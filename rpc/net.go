package rpc

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/harmony-one/harmony/internal/utils"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/harmony-one/harmony/p2p"
)

// PublicNetService offers network related RPC methods
type PublicNetService struct {
	net     p2p.Host
	chainID uint64
	version Version
}

// NewPublicNetAPI creates a new net API instance.
func NewPublicNetAPI(net p2p.Host, chainID uint64, version Version) rpc.API {
	// manually set different namespace to preserve legacy behavior
	var namespace string
	switch version {
	case V1:
		namespace = "net"
	case V2:
		namespace = "netv2"
	default:
		utils.Logger().Error().Msgf("Unknown version %v, ignoring API.", version)
		return rpc.API{}
	}

	return rpc.API{
		Namespace: namespace,
		Version:   APIVersion,
		Service:   &PublicNetService{net, chainID, version},
		Public:    true,
	}
}

// PeerCount returns the number of connected peers
func (s *PublicNetService) PeerCount() hexutil.Uint {
	return hexutil.Uint(s.net.GetPeerCount())
}

// Version returns the network version, i.e. ChainID identifying which network we are using
func (s *PublicNetService) Version() string {
	return fmt.Sprintf("%d", s.chainID)
}