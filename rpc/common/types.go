package common

import (
	"encoding/json"

	"github.com/harmony-one/harmony/internal/params"
	"github.com/libp2p/go-libp2p-core/peer"
)

// StructuredResponse type of RPCs
type StructuredResponse = map[string]interface{}

// NewStructuredResponse creates a structured response from the given input
func NewStructuredResponse(input interface{}) (StructuredResponse, error) {
	var objMap StructuredResponse
	dat, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(dat, &objMap); err != nil {
		return nil, err
	}
	return objMap, nil
}

// BlockArgs is struct to include optional block formatting params.
type BlockArgs struct {
	WithSigners bool     `json:"withSigners"`
	InclTx      bool     `json:"inclTx"`
	FullTx      bool     `json:"fullTx"`
	Signers     []string `json:"signers"`
	InclStaking bool     `json:"inclStaking"`
}

// UnmarshalFromInterface ..
func (ba *BlockArgs) UnmarshalFromInterface(blockArgs interface{}) error {
	var args BlockArgs
	dat, err := json.Marshal(blockArgs)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(dat, &args); err != nil {
		return err
	}
	*ba = args
	return nil
}

// TxHistoryArgs is struct to include optional transaction formatting params.
type TxHistoryArgs struct {
	Address   string `json:"address"`
	PageIndex uint32 `json:"pageIndex"`
	PageSize  uint32 `json:"pageSize"`
	FullTx    bool   `json:"fullTx"`
	TxType    string `json:"txType"`
	Order     string `json:"order"`
}

// UnmarshalFromInterface ..
func (ta *TxHistoryArgs) UnmarshalFromInterface(blockArgs interface{}) error {
	var args TxHistoryArgs
	dat, err := json.Marshal(blockArgs)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(dat, &args); err != nil {
		return err
	}
	*ta = args
	return nil
}

// C ..
type C struct {
	TotalKnownPeers int `json:"total-known-peers"`
	Connected       int `json:"connected"`
	NotConnected    int `json:"not-connected"`
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
	NodeBootTime   int64              `json:"node-unix-start-time"`
	PeerID         peer.ID            `json:"peerid"`
	C              C                  `json:"p2p-connectivity"`
}

// P captures the connected peers per topic
type P struct {
	Topic string    `json:"topic"`
	Peers []peer.ID `json:"peers"`
}

// NodePeerInfo captures the peer connectivity info of the node
type NodePeerInfo struct {
	PeerID       peer.ID   `json:"peerid"`
	BlockedPeers []peer.ID `json:"blocked-peers"`
	P            []P       `json:"connected-peers"`
}