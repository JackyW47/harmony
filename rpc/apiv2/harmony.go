package apiv2

import (
	"context"
	"math/big"

	"github.com/harmony-one/harmony/hmy"
	"github.com/harmony-one/harmony/internal/params"
	"github.com/harmony-one/harmony/rpc/common"
)

// PublicHarmonyAPI provides an API to access Harmony related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicHarmonyAPI struct {
	hmy *hmy.Harmony
}

// NewPublicHarmonyAPI ...
func NewPublicHarmonyAPI(hmy *hmy.Harmony) *PublicHarmonyAPI {
	return &PublicHarmonyAPI{hmy}
}

// ProtocolVersion returns the current Harmony protocol version this node supports
func (s *PublicHarmonyAPI) ProtocolVersion() int {
	return s.hmy.ProtocolVersion()
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicHarmonyAPI) Syncing() (interface{}, error) {
	// TODO(dm): find our Downloader module for syncing blocks
	return false, nil
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicHarmonyAPI) GasPrice(ctx context.Context) (*big.Int, error) {
	// TODO(ricl): add SuggestPrice API
	return big.NewInt(1), nil
}

// NodeMetadata captures select metadata of the RPC answering node
type NodeMetadata struct {
	BLSPublicKey   []string           `json:"blskey"`
	Version        string             `json:"version"`
	NetworkType    string             `json:"network"`
	ChainConfig    params.ChainConfig `json:"chain-config"`
	IsLeader       bool               `json:"is-leader"`
	ShardID        uint32             `json:"shard-id"`
	CurrentEpoch   uint64             `json:"current-epoch"`
	BlocksPerEpoch *uint64            `json:"blocks-per-epoch,omitempty"`
	Role           string             `json:"role"`
	DNSZone        string             `json:"dns-zone"`
	Archival       bool               `json:"is-archival"`
}

// GetNodeMetadata produces a NodeMetadata record, data is from the answering RPC node
func (s *PublicHarmonyAPI) GetNodeMetadata() common.NodeMetadata {
	return s.hmy.GetNodeMetadata()
}

// GetPeerInfo produces a NodePeerInfo record, containing peer info of the node
func (s *PublicHarmonyAPI) GetPeerInfo() common.NodePeerInfo {
	return s.hmy.GetPeerInfo()
}