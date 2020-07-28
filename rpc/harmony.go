package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/harmony-one/harmony/hmy"
	rpc_common "github.com/harmony-one/harmony/rpc/common"
)

// PublicHarmonyService provides an API to access Harmony related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicHarmonyService struct {
	hmy     *hmy.Harmony
	version Version
}

// NewPublicHarmonyAPI creates a new API for the RPC interface
func NewPublicHarmonyAPI(hmy *hmy.Harmony, version Version) rpc.API {
	return rpc.API{
		Namespace: version.Namespace(),
		Version:   APIVersion,
		Service:   &PublicHarmonyService{hmy, version},
		Public:    true,
	}
}

// ProtocolVersion returns the current Harmony protocol version this node supports
func (s *PublicHarmonyService) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(s.hmy.ProtocolVersion())
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicHarmonyService) Syncing() (interface{}, error) {
	// TODO(dm): find our Downloader module for syncing blocks
	return false, nil
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicHarmonyService) GasPrice(ctx context.Context) (*hexutil.Big, error) {
	// TODO(dm): add SuggestPrice API
	return (*hexutil.Big)(big.NewInt(1)), nil
}

// GetNodeMetadata produces a NodeMetadata record, data is from the answering RPC node
func (s *PublicHarmonyService) GetNodeMetadata() (rpc_common.StructuredResponse, error) {
	// Response output is the same for all versions
	return rpc_common.NewStructuredResponse(s.hmy.GetNodeMetadata())
}

// GetPeerInfo produces a NodePeerInfo record
func (s *PublicHarmonyService) GetPeerInfo() (rpc_common.StructuredResponse, error) {
	// Response output is the same for all versions
	return rpc_common.NewStructuredResponse(s.hmy.GetPeerInfo())
}