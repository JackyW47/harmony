package sync

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/harmony-one/harmony/block"
	headerV3 "github.com/harmony-one/harmony/block/v3"
	"github.com/harmony-one/harmony/core/types"
	"github.com/harmony-one/harmony/p2p/stream/common/ratelimiter"
	"github.com/harmony-one/harmony/p2p/stream/common/streammanager"
	syncpb "github.com/harmony-one/harmony/p2p/stream/protocols/sync/message"
	sttypes "github.com/harmony-one/harmony/p2p/stream/types"
	"github.com/harmony-one/harmony/shard"
)

var (
	_ sttypes.Request  = &getBlocksByNumberRequest{}
	_ sttypes.Request  = &getEpochBlockRequest{}
	_ sttypes.Response = &syncResponse{&syncpb.Response{}}
)

var (
	initStreamIDs = []sttypes.StreamID{
		makeTestStreamID(0),
		makeTestStreamID(1),
		makeTestStreamID(2),
		makeTestStreamID(3),
	}
)

var (
	testHeader           = &block.Header{Header: headerV3.NewHeader()}
	testBlock            = types.NewBlockWithHeader(testHeader)
	testHeaderBytes, _   = rlp.EncodeToBytes(testHeader)
	testBlockBytes, _    = rlp.EncodeToBytes(testBlock)
	testBlockResponse, _ = syncpb.MakeGetBlocksByNumResponse(0, [][]byte{testBlockBytes})

	testEpochState = &shard.State{
		Epoch:  new(big.Int).SetInt64(1),
		Shards: []shard.Committee{},
	}
	testEpochStateBytes, _ = rlp.EncodeToBytes(testEpochState)
	testEpochStateResponse = syncpb.MakeGetEpochStateResponse(0, testHeaderBytes, testEpochStateBytes)

	testErrorResponse = syncpb.MakeErrorResponse(0, errors.New("test error"))
)

func TestProtocol_GetBlocksByNumber(t *testing.T) {
	tests := []struct {
		getResponse   getResponseFn
		expErr        error
		expStID       sttypes.StreamID
		streamRemoved bool
	}{
		{
			getResponse: func(request sttypes.Request) (sttypes.Response, sttypes.StreamID) {
				return &syncResponse{
					pb: testBlockResponse,
				}, makeTestStreamID(0)
			},
			expErr:        nil,
			expStID:       makeTestStreamID(0),
			streamRemoved: false,
		},
		{
			getResponse: func(request sttypes.Request) (sttypes.Response, sttypes.StreamID) {
				return &syncResponse{
					pb: testEpochStateResponse,
				}, makeTestStreamID(0)
			},
			expErr:        errors.New("not GetBlockByNumber"),
			expStID:       makeTestStreamID(0),
			streamRemoved: true,
		},
		{
			getResponse:   nil,
			expErr:        errors.New("get response error"),
			expStID:       "",
			streamRemoved: true, // Does not exist at the first place
		},
		{
			getResponse: func(request sttypes.Request) (sttypes.Response, sttypes.StreamID) {
				return &syncResponse{
					pb: testErrorResponse,
				}, makeTestStreamID(0)
			},
			expErr:        errors.New("test error"),
			expStID:       makeTestStreamID(0),
			streamRemoved: true,
		},
	}

	for i, test := range tests {
		protocol := makeTestProtocol(test.getResponse)
		blocks, stid, err := protocol.GetBlocksByNumber(context.Background(), []uint64{0})

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Errorf("Test %v: %v", i, assErr)
			continue
		}
		if stid != test.expStID {
			t.Errorf("Test %v: unexpected st id: %v / %v", i, stid, test.expStID)
		}
		streamExist := protocol.sm.(*testStreamManager).isStreamExist(stid)
		if streamExist == test.streamRemoved {
			t.Errorf("Test %v: after request stream exist: %v / %v", i, streamExist, !test.streamRemoved)
		}
		if test.expErr == nil && (len(blocks) == 0) {
			t.Errorf("Test %v: zero blocks delivered", i)
		}
	}
}

func TestProtocol_GetEpochState(t *testing.T) {
	tests := []struct {
		getResponse   getResponseFn
		expErr        error
		expStID       sttypes.StreamID
		streamRemoved bool
	}{
		{
			getResponse: func(request sttypes.Request) (sttypes.Response, sttypes.StreamID) {
				return &syncResponse{
					pb: testEpochStateResponse,
				}, makeTestStreamID(0)
			},
			expErr:        nil,
			expStID:       makeTestStreamID(0),
			streamRemoved: false,
		},
		{
			getResponse: func(request sttypes.Request) (sttypes.Response, sttypes.StreamID) {
				return &syncResponse{
					pb: testBlockResponse,
				}, makeTestStreamID(0)
			},
			expErr:        errors.New("not GetEpochStateResponse"),
			expStID:       makeTestStreamID(0),
			streamRemoved: true,
		},
		{
			getResponse:   nil,
			expErr:        errors.New("get response error"),
			expStID:       "",
			streamRemoved: true, // Does not exist at the first place
		},
		{
			getResponse: func(request sttypes.Request) (sttypes.Response, sttypes.StreamID) {
				return &syncResponse{
					pb: testErrorResponse,
				}, makeTestStreamID(0)
			},
			expErr:        errors.New("test error"),
			expStID:       makeTestStreamID(0),
			streamRemoved: true,
		},
	}

	for i, test := range tests {
		protocol := makeTestProtocol(test.getResponse)
		res, stid, err := protocol.GetEpochState(context.Background(), 0)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Errorf("Test %v: %v", i, assErr)
			continue
		}
		if stid != test.expStID {
			t.Errorf("Test %v: unexpected st id: %v / %v", i, stid, test.expStID)
		}
		streamExist := protocol.sm.(*testStreamManager).isStreamExist(stid)
		if streamExist == test.streamRemoved {
			t.Errorf("Test %v: after request stream exist: %v / %v", i, streamExist, !test.streamRemoved)
		}
		if test.expErr == nil {
			if gotEpoch := res.state.Epoch; gotEpoch.Cmp(new(big.Int).SetUint64(1)) != 0 {
				t.Errorf("Test %v: unexpected epoch delivered: %v / %v", i, gotEpoch.String(), 1)
			}
		}
	}
}

type getResponseFn func(request sttypes.Request) (sttypes.Response, sttypes.StreamID)

type testHostRequestManager struct {
	getResponse getResponseFn
}

func makeTestProtocol(f getResponseFn) *Protocol {
	rm := &testHostRequestManager{f}

	streamIDs := make([]sttypes.StreamID, len(initStreamIDs))
	copy(streamIDs, initStreamIDs)
	sm := &testStreamManager{streamIDs}

	rl := ratelimiter.NewRateLimiter(10)

	return &Protocol{
		rm: rm,
		rl: rl,
		sm: sm,
	}
}

func (rm *testHostRequestManager) Start()                                             {}
func (rm *testHostRequestManager) Close()                                             {}
func (rm *testHostRequestManager) DeliverResponse(sttypes.StreamID, sttypes.Response) {}

func (rm *testHostRequestManager) DoRequest(ctx context.Context, request sttypes.Request) (sttypes.Response, sttypes.StreamID, error) {
	if rm.getResponse == nil {
		return nil, "", errors.New("get response error")
	}
	resp, stid := rm.getResponse(request)
	return resp, stid, nil
}

func makeTestStreamID(index int) sttypes.StreamID {
	id := fmt.Sprintf("[test stream %v]", index)
	return sttypes.StreamID(id)
}

// mock stream manager
type testStreamManager struct {
	streamIDs []sttypes.StreamID
}

func (sm *testStreamManager) Start() {}
func (sm *testStreamManager) Close() {}
func (sm *testStreamManager) SubscribeAddStreamEvent(chan<- streammanager.EvtStreamAdded) event.Subscription {
	return nil
}
func (sm *testStreamManager) SubscribeRemoveStreamEvent(chan<- streammanager.EvtStreamRemoved) event.Subscription {
	return nil
}

func (sm *testStreamManager) NewStream(stream sttypes.Stream) error {
	stid := stream.ID()
	for _, id := range sm.streamIDs {
		if id == stid {
			return errors.New("stream already exist")
		}
	}
	sm.streamIDs = append(sm.streamIDs, stid)
	return nil
}

func (sm *testStreamManager) RemoveStream(stID sttypes.StreamID) error {
	for i, id := range sm.streamIDs {
		if id == stID {
			sm.streamIDs = append(sm.streamIDs[:i], sm.streamIDs[i+1:]...)
		}
	}
	return errors.New("stream not exist")
}

func (sm *testStreamManager) isStreamExist(stid sttypes.StreamID) bool {
	for _, id := range sm.streamIDs {
		if id == stid {
			return true
		}
	}
	return false
}

func assertError(got, expect error) error {
	if (got == nil) != (expect == nil) {
		return fmt.Errorf("unexpected error: %v / %v", got, expect)
	}
	if got == nil {
		return nil
	}
	if !strings.Contains(got.Error(), expect.Error()) {
		return fmt.Errorf("unexpected error: %v/ %v", got, expect)
	}
	return nil
}
