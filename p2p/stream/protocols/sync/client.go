package sync

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/harmony-one/harmony/core/types"
	syncpb "github.com/harmony-one/harmony/p2p/stream/protocols/sync/message"
	sttypes "github.com/harmony-one/harmony/p2p/stream/types"
	"github.com/pkg/errors"
)

// GetBlocksByNumber do getBlocksByNumberRequest through sync stream protocol.
// Return the block as result, target stream id, and error
func (p *Protocol) GetBlocksByNumber(ctx context.Context, bns []uint64, opts ...Option) ([]*types.Block, sttypes.StreamID, error) {
	if len(bns) > GetBlocksByNumAmountCap {
		return nil, "", fmt.Errorf("number of blocks exceed cap of %v", GetBlocksByNumAmountCap)
	}

	req := newGetBlocksByNumberRequest(bns)
	resp, stid, err := p.rm.DoRequest(ctx, req, opts...)
	if err != nil {
		// At this point, error can be context canceled, context timed out, or waiting queue
		// is already full.
		return nil, stid, err
	}

	// Parse and return blocks
	blocks, err := req.getBlocksFromResponse(resp)
	if err != nil {
		return nil, stid, err
	}
	return blocks, stid, nil
}

// GetEpochState get the epoch block from querying the remote node running sync stream protocol.
// Currently, this method is only supported by beacon syncer.
// Note: use this after epoch chain is implemented.
func (p *Protocol) GetEpochState(ctx context.Context, epoch uint64, opts ...Option) (*EpochStateResult, sttypes.StreamID, error) {
	req := newGetEpochBlockRequest(epoch)

	resp, stid, err := p.rm.DoRequest(ctx, req, opts...)
	if err != nil {
		return nil, stid, err
	}

	res, err := epochStateResultFromResponse(resp)
	if err != nil {
		return nil, stid, err
	}
	return res, stid, nil
}

// GetCurrentBlockNumber get the current block number from remote node
func (p *Protocol) GetCurrentBlockNumber(ctx context.Context, opts ...Option) (uint64, sttypes.StreamID, error) {
	req := newGetBlockNumberRequest()

	resp, stid, err := p.rm.DoRequest(ctx, req, opts...)
	if err != nil {
		return 0, stid, err
	}

	bn, err := req.getNumberFromResponse(resp)
	if err != nil {
		return bn, stid, err
	}
	return bn, stid, nil
}

// GetBlockHashes do getBlockHashesRequest through sync stream protocol.
// Return the hash of the given block number. If a block is unknown, the hash will be emptyHash.
func (p *Protocol) GetBlockHashes(ctx context.Context, bns []uint64, opts ...Option) ([]common.Hash, sttypes.StreamID, error) {
	if len(bns) > GetBlockHashesAmountCap {
		return nil, "", fmt.Errorf("number of requested numbers exceed limit")
	}

	req := newGetBlockHashesRequest(bns)
	resp, stid, err := p.rm.DoRequest(ctx, req, opts...)
	if err != nil {
		return nil, stid, err
	}
	hashes, err := req.getHashesFromResponse(resp)
	if err != nil {
		return nil, stid, err
	}
	return hashes, stid, nil
}

// GetBlocksByHashes do getBlocksByHashesRequest through sync stream protocol.
func (p *Protocol) GetBlocksByHashes(ctx context.Context, hs []common.Hash, opts ...Option) ([]*types.Block, sttypes.StreamID, error) {
	if len(hs) > GetBlocksByHashesAmountCap {
		return nil, "", fmt.Errorf("number of requested hashes exceed limit")
	}
	req := newGetBlocksByHashesRequest(hs)
	resp, stid, err := p.rm.DoRequest(ctx, req, opts...)
	if err != nil {
		return nil, stid, err
	}
	blocks, err := req.getBlocksFromResponse(resp)
	if err != nil {
		return nil, stid, err
	}
	return blocks, stid, nil
}

// getBlocksByNumberRequest is the request for get block by numbers which implements
// sttypes.Request interface
type getBlocksByNumberRequest struct {
	bns   []uint64
	pbReq *syncpb.Request
}

func newGetBlocksByNumberRequest(bns []uint64) *getBlocksByNumberRequest {
	pbReq := syncpb.MakeGetBlocksByNumRequest(bns)
	return &getBlocksByNumberRequest{
		bns:   bns,
		pbReq: pbReq,
	}
}

func (req *getBlocksByNumberRequest) ReqID() uint64 {
	return req.pbReq.GetReqId()
}

func (req *getBlocksByNumberRequest) SetReqID(val uint64) {
	req.pbReq.ReqId = val
}

func (req *getBlocksByNumberRequest) String() string {
	ss := make([]string, 0, len(req.bns))
	for _, bn := range req.bns {
		ss = append(ss, strconv.Itoa(int(bn)))
	}
	bnsStr := strings.Join(ss, ",")
	return fmt.Sprintf("REQUEST [GetBlockByNumber: %s]", bnsStr)
}

func (req *getBlocksByNumberRequest) IsSupportedByProto(target sttypes.ProtoSpec) bool {
	return target.Version.GreaterThanOrEqual(MinVersion)
}

func (req *getBlocksByNumberRequest) Encode() ([]byte, error) {
	msg := syncpb.MakeMessageFromRequest(req.pbReq)
	return protobuf.Marshal(msg)
}

func (req *getBlocksByNumberRequest) getBlocksFromResponse(resp sttypes.Response) ([]*types.Block, error) {
	sResp, ok := resp.(*syncResponse)
	if !ok || sResp == nil {
		return nil, errors.New("not sync response")
	}
	blockBytes, err := req.parseBlockBytes(sResp)
	if err != nil {
		return nil, err
	}
	blocks := make([]*types.Block, 0, len(blockBytes))
	for _, bb := range blockBytes {
		var block *types.Block
		if err := rlp.DecodeBytes(bb, &block); err != nil {
			return nil, errors.Wrap(err, "[GetBlocksByNumResponse]")
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (req *getBlocksByNumberRequest) parseBlockBytes(resp *syncResponse) ([][]byte, error) {
	if errResp := resp.pb.GetErrorResponse(); errResp != nil {
		return nil, errors.New(errResp.Error)
	}
	gbResp := resp.pb.GetGetBlocksByNumResponse()
	if gbResp == nil {
		return nil, errors.New("response not GetBlockByNumber")
	}
	return gbResp.BlocksBytes, nil
}

type getEpochBlockRequest struct {
	epoch uint64
	pbReq *syncpb.Request
}

func newGetEpochBlockRequest(epoch uint64) *getEpochBlockRequest {
	pbReq := syncpb.MakeGetEpochStateRequest(epoch)
	return &getEpochBlockRequest{
		epoch: epoch,
		pbReq: pbReq,
	}
}

func (req *getEpochBlockRequest) ReqID() uint64 {
	return req.pbReq.GetReqId()
}

func (req *getEpochBlockRequest) SetReqID(val uint64) {
	req.pbReq.ReqId = val
}

func (req *getEpochBlockRequest) String() string {
	return fmt.Sprintf("REQUEST [GetEpochBlock: %v]", req.epoch)
}

func (req *getEpochBlockRequest) IsSupportedByProto(target sttypes.ProtoSpec) bool {
	return target.Version.GreaterThanOrEqual(MinVersion)
}

func (req *getEpochBlockRequest) Encode() ([]byte, error) {
	msg := syncpb.MakeMessageFromRequest(req.pbReq)
	return protobuf.Marshal(msg)
}

type getBlockNumberRequest struct {
	pbReq *syncpb.Request
}

func newGetBlockNumberRequest() *getBlockNumberRequest {
	pbReq := syncpb.MakeGetBlockNumberRequest()
	return &getBlockNumberRequest{
		pbReq: pbReq,
	}
}

func (req *getBlockNumberRequest) ReqID() uint64 {
	return req.pbReq.GetReqId()
}

func (req *getBlockNumberRequest) SetReqID(val uint64) {
	req.pbReq.ReqId = val
}

func (req *getBlockNumberRequest) String() string {
	return fmt.Sprintf("REQUEST [GetBlockNumber]")
}

func (req *getBlockNumberRequest) IsSupportedByProto(target sttypes.ProtoSpec) bool {
	return target.Version.GreaterThanOrEqual(MinVersion)
}

func (req *getBlockNumberRequest) Encode() ([]byte, error) {
	msg := syncpb.MakeMessageFromRequest(req.pbReq)
	return protobuf.Marshal(msg)
}

func (req *getBlockNumberRequest) getNumberFromResponse(resp sttypes.Response) (uint64, error) {
	sResp, ok := resp.(*syncResponse)
	if !ok || sResp == nil {
		return 0, errors.New("not sync response")
	}
	if errResp := sResp.pb.GetErrorResponse(); errResp != nil {
		return 0, errors.New(errResp.Error)
	}
	gnResp := sResp.pb.GetGetBlockNumberResponse()
	if gnResp == nil {
		return 0, errors.New("response not GetBlockNumber")
	}
	return gnResp.Number, nil
}

type getBlockHashesRequest struct {
	bns   []uint64
	pbReq *syncpb.Request
}

func newGetBlockHashesRequest(bns []uint64) *getBlockHashesRequest {
	pbReq := syncpb.MakeGetBlockHashesRequest(bns)
	return &getBlockHashesRequest{
		bns:   bns,
		pbReq: pbReq,
	}
}

func (req *getBlockHashesRequest) ReqID() uint64 {
	return req.pbReq.ReqId
}

func (req *getBlockHashesRequest) SetReqID(val uint64) {
	req.pbReq.ReqId = val
}

func (req *getBlockHashesRequest) String() string {
	ss := make([]string, 0, len(req.bns))
	for _, bn := range req.bns {
		ss = append(ss, strconv.Itoa(int(bn)))
	}
	bnsStr := strings.Join(ss, ",")
	return fmt.Sprintf("REQUEST [GetBlockHashes: %s]", bnsStr)
}

func (req *getBlockHashesRequest) IsSupportedByProto(target sttypes.ProtoSpec) bool {
	return target.Version.GreaterThanOrEqual(MinVersion)
}

func (req *getBlockHashesRequest) Encode() ([]byte, error) {
	msg := syncpb.MakeMessageFromRequest(req.pbReq)
	return protobuf.Marshal(msg)
}

func (req *getBlockHashesRequest) getHashesFromResponse(resp sttypes.Response) ([]common.Hash, error) {
	sResp, ok := resp.(*syncResponse)
	if !ok || sResp == nil {
		return nil, errors.New("not sync response")
	}
	if errResp := sResp.pb.GetErrorResponse(); errResp != nil {
		return nil, errors.New(errResp.Error)
	}
	bhResp := sResp.pb.GetGetBlockHashesResponse()
	if bhResp == nil {
		return nil, errors.New("response not GetBlockHashes")
	}
	hashBytes := bhResp.Hashes
	return bytesToHashes(hashBytes), nil
}

type getBlocksByHashesRequest struct {
	hashes []common.Hash
	pbReq  *syncpb.Request
}

func newGetBlocksByHashesRequest(hashes []common.Hash) *getBlocksByHashesRequest {
	pbReq := syncpb.MakeGetBlocksByHashesRequest(hashes)
	return &getBlocksByHashesRequest{
		hashes: hashes,
		pbReq:  pbReq,
	}
}

func (req *getBlocksByHashesRequest) ReqID() uint64 {
	return req.pbReq.GetReqId()
}

func (req *getBlocksByHashesRequest) SetReqID(val uint64) {
	req.pbReq.ReqId = val
}

func (req *getBlocksByHashesRequest) String() string {
	hashStrs := make([]string, 0, len(req.hashes))
	for _, h := range req.hashes {
		hashStrs = append(hashStrs, fmt.Sprintf("%x", h[:]))
	}
	hStr := strings.Join(hashStrs, ", ")
	return fmt.Sprintf("REQUEST [GetBlocksByHashes: %v]", hStr)
}

func (req *getBlocksByHashesRequest) IsSupportedByProto(target sttypes.ProtoSpec) bool {
	return target.Version.GreaterThanOrEqual(MinVersion)
}

func (req *getBlocksByHashesRequest) Encode() ([]byte, error) {
	msg := syncpb.MakeMessageFromRequest(req.pbReq)
	return protobuf.Marshal(msg)
}

func (req *getBlocksByHashesRequest) getBlocksFromResponse(resp sttypes.Response) ([]*types.Block, error) {
	sResp, ok := resp.(*syncResponse)
	if !ok || sResp == nil {
		return nil, errors.New("not sync response")
	}
	if errResp := sResp.pb.GetErrorResponse(); errResp != nil {
		return nil, errors.New(errResp.Error)
	}
	bhResp := sResp.pb.GetGetBlocksByHashesResponse()
	if bhResp == nil {
		return nil, errors.New("response not GetBlocksByHashes")
	}
	blockBytes := bhResp.BlocksBytes
	blocks := make([]*types.Block, 0, len(blockBytes))
	for _, bb := range blockBytes {
		var block *types.Block
		if err := rlp.DecodeBytes(bb, &block); err != nil {
			return nil, errors.Wrap(err, "[GetBlocksByHashesResponse]")
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}