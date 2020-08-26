package pubsub

import (
	"context"

	nodeconfig "github.com/harmony-one/harmony/internal/configs/node"
	"github.com/harmony-one/harmony/p2p"
	libp2p_host "github.com/libp2p/go-libp2p-core/host"
	libp2p_peer "github.com/libp2p/go-libp2p-core/peer"
	libp2p_pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Host is the client + server in p2p network.
type host interface {
	PubSubHost

	GetSelfPeer() p2p.Peer
	AddPeer(*p2p.Peer) error
	GetID() libp2p_peer.ID
	GetP2PHost() libp2p_host.Host
	GetPeerCount() int
	ConnectHostPeer(p2p.Peer) error
	PubSub() *libp2p_pubsub.PubSub
	C() (int, int, int)
	ListPeer(topic string) []libp2p_peer.ID
	ListTopic() []string
	ListBlockedPeer() []libp2p_peer.ID
}

// PubSubHost is the interface for pubSubHost.
type PubSubHost interface {
	AddPubSubHandler(psh PubSubHandler) error
	RemovePubSubHandler(spec HandlerSpecifier) error
	StartPubSubHandler(spec HandlerSpecifier) error
	StopPubSubHandler(spec HandlerSpecifier) error

	// SendMessageToGroups sends a message to one or more multicast groups.
	SendMessageToGroups(groups []nodeconfig.GroupID, msg []byte) error
}

// PubSubHandler is the pub sub message handler of a certain topic
// TODO: Add version string to topic to enable compatibility
// TODO: add decode algorithm with protobuf. Change arg in ValidateMsg and DeliverMsg from
//       []byte to decoded message
type PubSubHandler interface {
	// Topic is the topic the handler is subscribed to
	Topic() Topic

	// Specifier defines the uniques specifier of a pub sub handler
	Specifier() HandlerSpecifier

	// ValidateMsg validate the message from the peer with PeerID. Return ValidateResult.
	// Cache in the result is parsed to DeliverMsg and can be used for further message handling.
	ValidateMsg(ctx context.Context, peer PeerID, rawData []byte) ValidateResult

	// DeliverMsg deliver the message to target object with the validationCache from ValidateMsg
	// Note: For the same handler under a topic, DeliverMsg are executed concurrently.
	// And the error should be handled in HandleMsg for each handler.
	DeliverMsg(ctx context.Context, rawData []byte, cache ValidateCache)
}

// ValidateOptionProvider is the interface to provide pub-sub option with respect to the topic
type ValidateOptionProvider interface {
	getValidateOptions(topic Topic) []libp2p_pubsub.ValidatorOpt
}

// pubSub is the interface used within the module to mock out the PubSub
type pubSub interface {
	Join(topic Topic) (topicHandle, error)
	RegisterTopicValidator(topic Topic, val interface{}, opts ...libp2p_pubsub.ValidatorOpt) error
	UnregisterTopicValidator(topic Topic) error
}

// topicHandle is the interface for libp2p_pubsub.Topic
type topicHandle interface {
	Subscribe() (subscription, error)
}

// subscription is the interface for libp2p_pubsub.Subscription
type subscription interface {
	Next(ctx context.Context) (*libp2p_pubsub.Message, error)
	Cancel()
}
