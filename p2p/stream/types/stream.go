package sttypes

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"

	libp2p_network "github.com/libp2p/go-libp2p-core/network"
)

// Stream is the interface for streams implemented in each service.
// The stream interface is used for stream management as well as rate limiters
type Stream interface {
	ID() StreamID
	ProtoID() ProtoID
	ProtoSpec() (ProtoSpec, error)
	WriteBytes([]byte) error
	ReadBytes() ([]byte, error)
	Close() error // Make sure streams can handle multiple calls of Close
}

// BaseStream is the wrapper around
type BaseStream struct {
	raw libp2p_network.Stream
	rw  *bufio.ReadWriter

	// parse protocol spec fields
	spec     ProtoSpec
	specErr  error
	specOnce sync.Once
}

// NewBaseStream creates BaseStream as the wrapper of libp2p Stream
func NewBaseStream(st libp2p_network.Stream) *BaseStream {
	rw := bufio.NewReadWriter(bufio.NewReader(st), bufio.NewWriter(st))
	return &BaseStream{
		raw: st,
		rw:  rw,
	}
}

// StreamID is the unique identifier for the stream. It has the value of
// libp2p_network.Stream.ID()
type StreamID string

// Meta return the StreamID of the stream
func (st *BaseStream) ID() StreamID {
	return StreamID(st.raw.Conn().ID())
}

// ProtoID return the remote protocol ID of the stream
func (st *BaseStream) ProtoID() ProtoID {
	return ProtoID(st.raw.Protocol())
}

// ProtoSpec get the parsed protocol Specifier of the stream
func (st *BaseStream) ProtoSpec() (ProtoSpec, error) {
	st.specOnce.Do(func() {
		st.spec, st.specErr = ProtoIDToProtoSpec(st.ProtoID())
	})
	return st.spec, st.specErr
}

// Close close the stream on both sides.
func (st *BaseStream) Close() error {
	return st.raw.Reset()
}

const (
	maxMsgBytes = 20 * 1024 * 1024 // 20MB
	sizeBytes   = 4                // uint32
)

// WriteBytes write the bytes to the stream.
// First 4 bytes is used as the size bytes, and the rest is the content
func (st *BaseStream) WriteBytes(b []byte) error {
	fmt.Println("write bytes", len(b))
	if len(b) > maxMsgBytes {
		return errors.New("message too long")
	}
	if _, err := st.rw.Write(intToBytes(len(b))); err != nil {
		return errors.Wrap(err, "write size bytes")
	}
	if _, err := st.rw.Write(b); err != nil {
		return errors.Wrap(err, "write content")
	}
	return st.rw.Flush()
}

// ReadMsg read the bytes from the stream
func (st *BaseStream) ReadBytes() ([]byte, error) {
	sb := make([]byte, sizeBytes)
	_, err := st.rw.Read(sb)
	if err != nil {
		return nil, errors.Wrap(err, "read size")
	}
	size := bytesToInt(sb)

	cb := make([]byte, size)
	n, err := io.ReadFull(st.rw, cb)
	if err != nil {
		fmt.Println("size prefix", size, n)
		return nil, errors.Wrap(err, "read content")
	}

	if n != size {
		fmt.Println("size prefix", size, n)
		return nil, errors.New("ReadBytes sanity failed: byte size")
	}
	return cb, nil
}

func intToBytes(val int) []byte {
	b := make([]byte, sizeBytes) // uint32
	binary.LittleEndian.PutUint32(b, uint32(val))
	return b
}

func bytesToInt(b []byte) int {
	val := binary.LittleEndian.Uint32(b)
	return int(val)
}
