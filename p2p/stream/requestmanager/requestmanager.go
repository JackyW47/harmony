package requestmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/harmony-one/harmony/internal/utils"
	"github.com/harmony-one/harmony/p2p/stream/message"
	"github.com/harmony-one/harmony/p2p/stream/streammanager"
	sttypes "github.com/harmony-one/harmony/p2p/stream/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// requestManager implements RequestManager. It is responsible for matching response
// with requests.
// TODO: each peer is able to have a queue of requests instead of one request at a time.
// TODO: add QoS evaluation for each stream
type requestManager struct {
	streams   map[sttypes.StreamID]*stream  // All streams
	available map[sttypes.StreamID]struct{} // Streams that are available for request
	pendings  map[uint64]*request           //requests that
	waitings  requestQueue                  // double linked list of requests that are on the waiting list

	// Stream events
	newStreamC <-chan streammanager.EvtStreamAdded
	rmStreamC  <-chan streammanager.EvtStreamRemoved
	// Request events
	cancelReqC  chan uint64 // request being canceled
	retryReqC   chan uint64 // request to be retried
	deliveryC   chan deliverData
	newRequestC chan *request

	subs   []event.Subscription
	logger zerolog.Logger
	stopC  chan struct{}
}

func newRequestManager(sm streammanager.Subscriber) *requestManager {
	// subscribe at initialize to prevent misuse of upper function which might cause
	// the bootstrap peers are ignored
	newStreamC := make(chan streammanager.EvtStreamAdded)
	rmStreamC := make(chan streammanager.EvtStreamRemoved)
	sub1 := sm.SubscribeAddStreamEvent(newStreamC)
	sub2 := sm.SubscribeRemoveStreamEvent(rmStreamC)

	logger := utils.Logger().With().Str("module", "request manager").Logger()

	return &requestManager{
		streams: make(map[sttypes.StreamID]*stream),

		newStreamC:  newStreamC,
		rmStreamC:   rmStreamC,
		cancelReqC:  make(chan uint64, 16),
		retryReqC:   make(chan uint64, 16),
		deliveryC:   make(chan deliverData, 128),
		newRequestC: make(chan *request, 128),

		subs:   []event.Subscription{sub1, sub2},
		logger: logger,
		stopC:  make(chan struct{}),
	}
}

func (rm *requestManager) Start() {
	go rm.loop()
}

func (rm *requestManager) Close() {
	rm.stopC <- struct{}{}
}

func (rm *requestManager) DoRequest(ctx context.Context, raw sttypes.Request) (*message.Response, error) {
	resp := <-rm.doRequestAsync(ctx, raw)
	return resp.resp, resp.err
}

func (rm *requestManager) doRequestAsync(ctx context.Context, raw sttypes.Request) <-chan response {
	req := &request{
		Request: raw,
		respC:   make(chan *message.Response),
		waitCh:  make(chan struct{}),
	}
	rm.newRequestC <- req

	resC := make(chan response, 1)
	go func() {
		defer close(req.waitCh)
		select {
		case <-ctx.Done(): // canceled or timeout in upper function calls
			rm.cancelReqC <- req.ReqID()
			resC <- response{err: ctx.Err()}

		case resp := <-req.respC:
			resC <- response{resp: resp, err: req.err}
		}
	}()
	return resC
}

// DeliverResponse delivers the response to the corresponding request.
// The function behaves non-block
func (rm *requestManager) DeliverResponse(stID sttypes.StreamID, resp *message.Response) {
	dlv := deliverData{
		resp: resp,
		stID: stID,
	}
	go func() {
		select {
		case rm.deliveryC <- dlv:
		case <-time.After(deliverTimeout):
			rm.logger.Error().Msg("WARNING: delivery timeout. Possible stuck in loop")
		}
	}()
}

func (rm *requestManager) loop() {
	var (
		throttleC = make(chan struct{}, 1) // throttle the waiting requests periodically
		ticker    = time.NewTicker(throttleInterval)
	)
	throttle := func() {
		select {
		case throttleC <- struct{}{}:
		default:
		}
	}

	for {
		select {
		case <-ticker.C:
			throttle()

		case <-throttleC:
		loop:
			for i := 0; i != throttleBatch; i++ {
				req, st := rm.getNextRequest()
				if req == nil {
					break loop
				}
				rm.addPendingRequest(req, st)

				go func(reqID uint64, reqMsg *message.Request) {
					if err := st.SendRequest(reqMsg); err != nil {
						rm.logger.Warn().Str("streamID", st.ID().String()).Err(err).
							Msg("failed to send request")
						rm.retryReqC <- reqID
					}
					go func() {
						select {
						case <-time.After(reqTimeOut):
							// request still not received after reqTimeOut, try again.
							rm.retryReqC <- reqID
						case <-req.waitCh:
							// request cancelled or response received. Do nothing and return
						}
					}()
				}(req.ReqID(), req.GetRequestMessage())
			}

		case req := <-rm.newRequestC:
			added := rm.handleNewRequest(req)
			if added {
				throttle()
			}

		case data := <-rm.deliveryC:
			rm.handleDeliverData(data)

		case reqID := <-rm.retryReqC:
			added := rm.handleRetryRequest(reqID)
			if added {
				throttle()
			}

		case reqID := <-rm.cancelReqC:
			rm.handleCancelRequest(reqID)

		case evt := <-rm.newStreamC:
			rm.logger.Info().Str("streamID", evt.Stream.ID().String()).Msg("add new stream")
			rm.addNewStream(evt.Stream)

		case evt := <-rm.rmStreamC:
			rm.logger.Info().Str("streamID", evt.ID.String()).Msg("remove stream")
			reqCanceled := rm.removeStream(evt.ID)
			if reqCanceled {
				throttle()
			}

		case <-rm.stopC:
			rm.logger.Info().Msg("request manager stopped")
			rm.close()
			return
		}
	}
}

func (rm *requestManager) handleNewRequest(req *request) bool {
	err := rm.addNewRequestToWaitings(req, reqPriorityLow)
	if err != nil {
		rm.logger.Warn().Err(err).Msg("failed to add new request to waitings")
		req.err = errors.Wrap(err, "failed to add new request to waitings")
		req.respC <- nil
		return false
	}
	return true
}

func (rm *requestManager) handleDeliverData(data deliverData) {
	if err := rm.validateDelivery(data); err != nil {
		// if error happens in delivery, most likely it's a stale delivery. No action needed
		// and return
		rm.logger.Warn().Err(err).Interface("response", data.resp).Msg("unable to validate deliver")
		return
	}
	// req and st is ensured not to be empty in validateDelivery
	req := rm.pendings[data.resp.ReqId]
	req.respC <- data.resp
	rm.removePendingRequest(req)
}

func (rm *requestManager) validateDelivery(data deliverData) error {
	st := rm.streams[data.stID]
	if st == nil {
		return fmt.Errorf("data delivered from dead stream: %v", data.stID)
	}
	req := rm.pendings[data.resp.ReqId]
	if req == nil {
		return fmt.Errorf("stale p2p response delivery")
	}
	if req.owner == nil || req.owner.ID() != data.stID {
		return fmt.Errorf("unexpected delivery stream")
	}
	if st.req == nil || st.req.ReqID() != data.resp.ReqId {
		// Possible when request is canceled
		return fmt.Errorf("unexpected deliver request")
	}
	return nil
}

func (rm *requestManager) handleCancelRequest(reqID uint64) {
	req, ok := rm.pendings[reqID]
	if !ok {
		return
	}
	rm.removePendingRequest(req)
}

func (rm *requestManager) handleRetryRequest(reqID uint64) bool {
	req, ok := rm.pendings[reqID]
	if !ok {
		return false
	}
	rm.removePendingRequest(req)

	if err := rm.addNewRequestToWaitings(req, reqPriorityHigh); err != nil {
		rm.logger.Warn().Err(err).Msg("cannot add request to waitings during retry")
		req.err = errors.Wrap(err, "cannot add request to waitings during retry")
		req.respC <- nil
		return false
	}
	return true
}

func (rm *requestManager) getNextRequest() (*request, *stream) {
	req := rm.waitings.pop()
	if req == nil {
		return nil, nil
	}
	st, err := rm.pickAvailableStream()
	if err != nil {
		rm.addNewRequestToWaitings(req, reqPriorityHigh)
		return nil, nil
	}
	return req, st
}

func (rm *requestManager) genReqID() uint64 {
	for {
		rid := sttypes.GenReqID()
		if _, ok := rm.pendings[rid]; !ok {
			return rid
		}
	}
}

func (rm *requestManager) addPendingRequest(req *request, st *stream) {
	reqID := rm.genReqID()
	req.SetReqID(reqID)

	req.owner = st
	st.req = req

	delete(rm.available, st.ID())
	rm.pendings[req.ReqID()] = req
}

func (rm *requestManager) removePendingRequest(req *request) {
	delete(rm.pendings, req.ReqID())

	if st := req.owner; st != nil {
		st.clearPendingRequest()
		rm.available[st.ID()] = struct{}{}
	}
}

func (rm *requestManager) pickAvailableStream() (*stream, error) {
	for id := range rm.available {
		st, ok := rm.streams[id]
		if !ok {
			return nil, errors.New("sanity error: available stream not registered")
		}
		if st.req != nil {
			return nil, errors.New("sanity error: available stream has pending requests")
		}
		return st, nil
	}
	return nil, errors.New("no more available streams")
}

func (rm *requestManager) addNewStream(st sttypes.Stream) {
	if _, ok := rm.streams[st.ID()]; !ok {
		rm.streams[st.ID()] = &stream{Stream: st}
		rm.available[st.ID()] = struct{}{}
	}
}

// removeStream remove the stream from request manager, clear the pending request
// of the stream. Return whether a pending request is canceled in the stream,
func (rm *requestManager) removeStream(id sttypes.StreamID) bool {
	st, ok := rm.streams[id]
	if !ok {
		return false
	}
	delete(rm.available, id)
	delete(rm.streams, id)

	cleared := st.clearPendingRequest()
	if cleared != nil {
		if err := rm.addNewRequestToWaitings(cleared, reqPriorityHigh); err != nil {
			rm.logger.Err(err).Msg("cannot add new request to waitings in removeStream")
			return false
		}
		return true
	}
	return false
}

func (rm *requestManager) close() {
	for _, sub := range rm.subs {
		sub.Unsubscribe()
	}
	for _, req := range rm.pendings {
		req.err = errors.New("request manager module closed")
		req.respC <- nil
	}
	close(rm.stopC)
}

type reqPriority int

const (
	reqPriorityLow reqPriority = iota
	reqPriorityHigh
)

func (rm *requestManager) addNewRequestToWaitings(req *request, priority reqPriority) error {
	switch priority {
	case reqPriorityHigh:
		return rm.waitings.pushFront(req)
	case reqPriorityLow:
		return rm.waitings.pushBack(req)
	}
	return nil
}