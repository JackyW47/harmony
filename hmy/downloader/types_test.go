package downloader

import (
	"container/heap"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/harmony-one/harmony/block"
	headerV3 "github.com/harmony-one/harmony/block/v3"
	"github.com/harmony-one/harmony/core/types"
	sttypes "github.com/harmony-one/harmony/p2p/stream/types"
)

func TestResultQueue_AddBlockResults(t *testing.T) {
	tests := []struct {
		initBNs []uint64
		addBNs  []uint64
		expErr  error
		expSize int
	}{
		{
			initBNs: []uint64{},
			addBNs:  []uint64{1, 2, 3, 4},
			expErr:  nil,
			expSize: 4,
		},
		{
			initBNs: []uint64{1, 2, 3, 4},
			addBNs:  []uint64{5, 6, 7, 8},
			expErr:  nil,
			expSize: 8,
		},
		{
			initBNs: func() []uint64 {
				res := make([]uint64, 0, queueMaxSize-3)
				for i := uint64(0); i < uint64(queueMaxSize-3); i++ {
					res = append(res, i)
				}
				return res
			}(),
			addBNs:  []uint64{1, 2, 3, 4},
			expErr:  errResultQueueFull,
			expSize: queueMaxSize - 3,
		},
	}
	for i, test := range tests {
		rq := makeTestResultQueue(test.initBNs)
		err := rq.addBlockResults(makeTestBlocks(test.addBNs), "")
		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Errorf("Test %v: %v", i, assErr)
		}
		if rq.results.Len() != test.expSize {
			t.Errorf("Test %v: unexpected size: %v / %v", i, rq.results.Len(), test.expSize)
		}
	}
}

func TestResultQueue_PopBlockResults(t *testing.T) {
	tests := []struct {
		initBNs  []uint64
		cap      int
		expStart uint64
		expSize  int
	}{
		{
			initBNs:  []uint64{1, 2, 3, 4, 5},
			cap:      3,
			expStart: 1,
			expSize:  3,
		},
		{
			initBNs:  []uint64{1, 2, 3, 4, 5},
			cap:      10,
			expStart: 1,
			expSize:  5,
		},
		{
			initBNs:  []uint64{1, 3, 4, 5},
			cap:      10,
			expStart: 1,
			expSize:  1,
		},
		{
			initBNs:  []uint64{1, 2, 3, 4, 5},
			cap:      10,
			expStart: 0,
			expSize:  0,
		},
		{
			initBNs:  []uint64{1, 1, 1, 1, 2},
			cap:      10,
			expStart: 1,
			expSize:  2,
		},
		{
			initBNs:  []uint64{1, 2, 3, 4, 5},
			cap:      10,
			expStart: 2,
			expSize:  4,
		},
	}
	for i, test := range tests {
		rq := makeTestResultQueue(test.initBNs)
		res := rq.popBlockResults(test.expStart, test.cap)
		if len(res) != test.expSize {
			t.Errorf("Test %v: unexpect size %v / %v", i, len(res), test.expSize)
		}
	}
}

func TestResultQueue_RemoveResultsByStreamID(t *testing.T) {
	tests := []struct {
		rq         *resultQueue
		rmStreamID sttypes.StreamID
		removed    int
		expSize    int
	}{
		{
			rq:         makeTestResultQueue([]uint64{1, 2, 3, 4}),
			rmStreamID: "test stream id",
			removed:    4,
			expSize:    0,
		},
		{
			rq: func() *resultQueue {
				rq := makeTestResultQueue([]uint64{2, 3, 4, 5})
				rq.addBlockResults([]*types.Block{
					makeTestBlock(1),
					makeTestBlock(5),
					makeTestBlock(6),
				}, "another test stream id")
				return rq
			}(),
			rmStreamID: "test stream id",
			removed:    4,
			expSize:    3,
		},
		{
			rq: func() *resultQueue {
				rq := makeTestResultQueue([]uint64{2, 3, 4, 5})
				rq.addBlockResults([]*types.Block{
					makeTestBlock(1),
					makeTestBlock(5),
					makeTestBlock(6),
				}, "another test stream id")
				return rq
			}(),
			rmStreamID: "another test stream id",
			removed:    3,
			expSize:    4,
		},
	}
	for i, test := range tests {
		res := test.rq.removeResultsByStreamID(test.rmStreamID)
		if len(res) != test.removed {
			t.Errorf("Test %v: unexpected number removed %v / %v", i, len(res), test.removed)
		}
		if gotSize := test.rq.results.Len(); gotSize != test.expSize {
			t.Errorf("Test %v: unexpected number after removal %v / %v", i, gotSize, test.expSize)
		}
	}
}

func makeTestResultQueue(bns []uint64) *resultQueue {
	rq := newResultQueue()
	for _, bn := range bns {
		heap.Push(rq.results, &blockResult{
			block: makeTestBlock(bn),
			stid:  "test stream id",
		})
	}
	return rq
}

func TestPrioritizedBlocks(t *testing.T) {
	addBNs := []uint64{4, 7, 6, 9}

	bns := newPrioritizedNumbers()
	for _, bn := range addBNs {
		bns.push(bn)
	}
	prevBN := uint64(0)
	for len(*bns.q) > 0 {
		b := bns.pop()
		if b < prevBN {
			t.Errorf("number not incrementing")
		}
		prevBN = b
	}
}

func TestPriorityQueue(t *testing.T) {
	testBNs := []uint64{1, 9, 2, 4, 5, 12}
	pq := make(priorityQueue, 0, 10)
	heap.Init(&pq)
	for _, bn := range testBNs {
		heap.Push(&pq, &blockResult{
			block: makeTestBlock(bn),
			stid:  "",
		})
	}
	cmpBN := uint64(0)
	for pq.Len() > 0 {
		bn := heap.Pop(&pq).(*blockResult).block.NumberU64()
		if bn < cmpBN {
			t.Errorf("not incrementing")
		}
		cmpBN = bn
	}
	if pq.Len() != 0 {
		t.Errorf("after poping, size not 0")
	}
}

func makeTestBlocks(bns []uint64) []*types.Block {
	blocks := make([]*types.Block, 0, len(bns))
	for _, bn := range bns {
		blocks = append(blocks, makeTestBlock(bn))
	}
	return blocks
}

func makeTestBlock(bn uint64) *types.Block {
	testHeader := &block.Header{Header: headerV3.NewHeader()}
	testHeader.SetNumber(big.NewInt(int64(bn)))
	return types.NewBlockWithHeader(testHeader)
}

func assertError(got, expect error) error {
	if (got == nil) != (expect == nil) {
		return fmt.Errorf("unexpected error [%v] / [%v]", got, expect)
	}
	if (got == nil) || (expect == nil) {
		return nil
	}
	if !strings.Contains(got.Error(), expect.Error()) {
		return fmt.Errorf("unexpected error [%v] / [%v]", got, expect)
	}
	return nil
}
