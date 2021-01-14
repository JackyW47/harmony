package sync

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	protobuf "github.com/golang/protobuf/proto"
	syncpb "github.com/harmony-one/harmony/p2p/stream/protocols/sync/message"
	sttypes "github.com/harmony-one/harmony/p2p/stream/types"
	libp2p_network "github.com/libp2p/go-libp2p-core/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// syncStream is the structure for a stream running sync protocol.
type syncStream struct {
	// Basic stream
	*sttypes.BaseStream

	protocol *Protocol
	chain    chainHelper

	// pipeline channels
	reqC  chan *syncpb.Request
	respC chan *syncpb.Response

	// close related fields. Concurrent call of close is possible.
	closeC    chan struct{}
	closeStat uint32

	logger zerolog.Logger
}

// wrapStream wraps the raw libp2p stream to syncStream
func (p *Protocol) wrapStream(raw libp2p_network.Stream) *syncStream {
	bs := sttypes.NewBaseStream(raw)
	logger := p.logger.With().
		Str("ID", string(bs.ID())).
		Str("Remote Protocol", string(bs.ProtoID())).
		Logger()

	return &syncStream{
		BaseStream: bs,
		protocol:   p,
		chain:      newChainHelper(p.chain, p.schedule),
		reqC:       make(chan *syncpb.Request, 100),
		respC:      make(chan *syncpb.Response, 100),
		closeC:     make(chan struct{}),
		closeStat:  0,
		logger:     logger,
	}
}

func (st *syncStream) run() {
	go st.readMsgLoop()
	go st.handleReqLoop()
	go st.handleRespLoop()
}

// readMsgLoop is the loop
func (st *syncStream) readMsgLoop() {
	for {
		msg, err := st.readMsg()
		if err != nil {
			if err := st.Close(); err != nil {
				st.logger.Err(err).Msg("failed to close sync stream")
			}
			fmt.Println(err)
			return
		}
		st.deliverMsg(msg)
	}
}

// deliverMsg process the delivered message and forward to the corresponding channel
func (st *syncStream) deliverMsg(msg protobuf.Message) {
	syncMsg := msg.(*syncpb.Message)
	if syncMsg == nil {
		st.logger.Info().Str("message", msg.String()).Msg("received unexpected sync message")
		return
	}
	if req := syncMsg.GetReq(); req != nil {
		go func() {
			select {
			case st.reqC <- req:
			case <-time.After(1 * time.Minute):
				st.logger.Warn().Str("request", req.String()).
					Msg("request handler severely jammed, message dropped")
			}
		}()
	}
	if resp := syncMsg.GetResp(); resp != nil {
		go func() {
			select {
			case st.respC <- resp:
			case <-time.After(1 * time.Minute):
				st.logger.Warn().Str("response", resp.String()).
					Msg("response handler severely jammed, message dropped")
			}
		}()
	}
	return
}

func (st *syncStream) handleReqLoop() {
	for {
		select {
		case req := <-st.reqC:
			st.protocol.rl.LimitRequest(st.ID())
			err := st.handleReq(req)

			if err != nil {
				st.logger.Info().Err(err).Str("request", req.String()).
					Msg("handle request error. Closing stream")
				if err := st.Close(); err != nil {
					st.logger.Err(err).Msg("failed to close sync stream")
				}
				return
			}

		case <-st.closeC:
			return
		}
	}
}

func (st *syncStream) handleRespLoop() {
	for {
		select {
		case resp := <-st.respC:
			st.handleResp(resp)

		case <-st.closeC:
			return
		}
	}
}

// Close stops the stream handling and closes the underlying stream
func (st *syncStream) Close() error {
	// Hack here: Close is called in two cases:
	// 1. error happened when running the stream (readMsgLoop, handleMsgLoop)
	// 2. error happened when validating the result delivered by the stream, thus
	//    close the stream at stream manager.
	// Thus we only need to close for only once to prevent recursive call of the
	// Close method (syncStream -> StreamManager -> syncStream -> ...)
	notClosed := atomic.CompareAndSwapUint32(&st.closeStat, 0, 1)
	if !notClosed {
		// Already closed by another goroutine. Directly return
		return nil
	}
	err := st.BaseStream.Close()
	if err := st.protocol.sm.RemoveStream(st.ID()); err != nil {
		st.logger.Err(err).Str("stream ID", string(st.ID())).
			Msg("failed to remove sync stream on close")
	}
	close(st.closeC)
	return err
}

func (st *syncStream) handleReq(req *syncpb.Request) error {
	if bnReq := req.GetGetBlocksByNumRequest(); bnReq != nil {
		return st.handleGetBlocksByNumRequest(req.ReqId, bnReq)
	}
	if esReq := req.GetGetEpochStateRequest(); esReq != nil {
		return st.handleEpochStateRequest(req.ReqId, esReq)
	}
	// unsupported request type
	resp := syncpb.MakeErrorResponseMessage(req.ReqId, errUnknownReqType)
	return st.writeMsg(resp)
}

func (st *syncStream) handleGetBlocksByNumRequest(rid uint64, req *syncpb.GetBlocksByNumRequest) error {
	resp, err := st.computeRespFromBlockNumber(rid, req.Nums)
	if resp == nil && err != nil {
		resp = syncpb.MakeErrorResponseMessage(rid, err)
	}
	if writeErr := st.writeMsg(resp); writeErr != nil {
		if err == nil {
			err = writeErr
		} else {
			err = fmt.Errorf("%v; [writeMsg] %v", err.Error(), writeErr)
		}
	}
	return errors.Wrap(err, "[GetBlocksByNumber]")
}

func (st *syncStream) handleEpochStateRequest(rid uint64, req *syncpb.GetEpochStateRequest) error {
	resp, err := st.computeEpochStateResp(rid, req.Epoch)
	if resp == nil && err != nil {
		resp = syncpb.MakeErrorResponseMessage(rid, err)
	}
	if writeErr := st.writeMsg(resp); writeErr != nil {
		if err == nil {
			err = writeErr
		} else {
			err = fmt.Errorf("%v; [writeMsg] %v", err.Error(), writeErr)
		}
	}
	return errors.Wrap(err, "[GetEpochState]")
}

func (st *syncStream) handleResp(resp *syncpb.Response) {
	st.protocol.rm.DeliverResponse(st.ID(), &syncResponse{resp})
}

func (st *syncStream) readMsg() (*syncpb.Message, error) {
	b, err := st.ReadBytes()
	if err != nil {
		return nil, err
	}
	var msg = &syncpb.Message{}
	if err := protobuf.Unmarshal(b, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (st *syncStream) writeMsg(msg *syncpb.Message) error {
	b, err := protobuf.Marshal(msg)
	if err != nil {
		return err
	}
	return st.WriteBytes(b)
}

func (st *syncStream) computeRespFromBlockNumber(rid uint64, bns []uint64) (*syncpb.Message, error) {
	if len(bns) > GetBlocksByNumAmountCap {
		err := fmt.Errorf("GetBlocksByNum amount exceed cap: %v/%v", len(bns), GetBlocksByNumAmountCap)
		return nil, err
	}
	blocks := st.chain.getBlocks(bns)

	blocksBytes := make([][]byte, 0, len(blocks))
	for _, block := range blocks {
		bb, err := rlp.EncodeToBytes(block)
		if err != nil {
			return nil, err
		}
		blocksBytes = append(blocksBytes, bb)
	}
	return syncpb.MakeGetBlocksByNumResponseMessage(rid, blocksBytes)
}

func (st *syncStream) computeEpochStateResp(rid uint64, epoch uint64) (*syncpb.Message, error) {
	if epoch == 0 {
		return nil, errors.New("Epoch 0 does not have shard state")
	}
	esRes, err := st.chain.getEpochState(epoch)
	if err != nil {
		return nil, err
	}
	return esRes.toMessage(rid)
}
