package sync

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/harmony-one/harmony/block"
	"github.com/harmony-one/harmony/consensus/engine"
	"github.com/harmony-one/harmony/core/types"
	shardingconfig "github.com/harmony-one/harmony/internal/configs/sharding"
	"github.com/pkg/errors"
)

// chainHelper is the adapter for blockchain which is friendly to unit test.
type chainHelper interface {
	getCurrentBlockNumber() uint64
	getBlockHashes(bns []uint64) []common.Hash
	getBlocksByNumber(bns []uint64) ([]*types.Block, error)
	getBlocksByHashes(hs []common.Hash) ([]*types.Block, error)
	getEpochState(epoch uint64) (*EpochStateResult, error)
}

type chainHelperImpl struct {
	chain    engine.ChainReader
	schedule shardingconfig.Schedule
}

func newChainHelper(chain engine.ChainReader, schedule shardingconfig.Schedule) *chainHelperImpl {
	return &chainHelperImpl{
		chain:    chain,
		schedule: schedule,
	}
}

func (ch *chainHelperImpl) getCurrentBlockNumber() uint64 {
	return ch.chain.CurrentBlock().NumberU64()
}

func (ch *chainHelperImpl) getBlockHashes(bns []uint64) []common.Hash {
	hashes := make([]common.Hash, 0, len(bns))
	for _, bn := range bns {
		var h common.Hash
		header := ch.chain.GetHeaderByNumber(bn)
		if header != nil {
			h = header.Hash()
		}
		hashes = append(hashes, h)
	}
	return hashes
}

func (ch *chainHelperImpl) getBlocksByNumber(bns []uint64) ([]*types.Block, error) {
	var (
		blocks = make([]*types.Block, 0, len(bns))
	)
	for _, bn := range bns {
		var (
			block *types.Block
			err   error
		)
		header := ch.chain.GetHeaderByNumber(bn)
		if header != nil {
			block, err = ch.getBlockWithSigByHeader(header)
			if err != nil {
				return nil, errors.Wrapf(err, "get block %v at %v", header.Hash().String(), header.Number())
			}
		}
		blocks = append(blocks, block)

	}
	return blocks, nil
}

func (ch *chainHelperImpl) getBlocksByHashes(hs []common.Hash) ([]*types.Block, error) {
	var (
		blocks = make([]*types.Block, 0, len(hs))
	)
	for _, h := range hs {
		var (
			block *types.Block
			err   error
		)
		header := ch.chain.GetHeaderByHash(h)
		if header != nil {
			block, err = ch.getBlockWithSigByHeader(header)
			if err != nil {
				return nil, errors.Wrapf(err, "get block %v at %v", header.Hash().String(), header.Number())
			}
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

var errBlockNotFound = errors.New("block not found")

func (ch *chainHelperImpl) getBlockWithSigByHeader(header *block.Header) (*types.Block, error) {
	b := ch.chain.GetBlock(header.Hash(), header.Number().Uint64())
	if b == nil {
		return nil, nil
	}
	commitSig, err := ch.chain.ReadCommitSig(header.Number().Uint64())
	if err != nil {
		return nil, errors.New("missing commit signature")
	}
	if len(commitSig) != 0 {
		b.SetCurrentCommitSig(commitSig)
	}
	return b, nil
}

func (ch *chainHelperImpl) getEpochState(epoch uint64) (*EpochStateResult, error) {
	if ch.chain.ShardID() != 0 {
		return nil, errors.New("get epoch state currently unavailable on side chain")
	}
	if epoch == 0 {
		return nil, errors.New("nil shard state for epoch 0")
	}
	res := &EpochStateResult{}

	targetBN := ch.schedule.EpochLastBlock(epoch - 1)
	res.Header = ch.chain.GetHeaderByNumber(targetBN)
	if res.Header == nil {
		// we still don't have the given epoch
		return res, nil
	}
	epochBI := new(big.Int).SetUint64(epoch)
	if ch.chain.Config().IsPreStaking(epochBI) {
		// For epoch before preStaking, only hash is stored in header
		ss, err := ch.chain.ReadShardState(epochBI)
		if err != nil {
			return nil, err
		}
		if ss == nil {
			return nil, fmt.Errorf("missing shard state for [EPOCH-%v]", epoch)
		}
		res.State = ss
	}
	return res, nil
}
