package consensus

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	bls_core "github.com/harmony-one/bls/ffi/go/bls"
	msg_pb "github.com/harmony-one/harmony/api/proto/message"
	"github.com/harmony-one/harmony/core/types"
	"github.com/harmony-one/harmony/crypto/bls"
	bls_cosi "github.com/harmony-one/harmony/crypto/bls"
	"github.com/harmony-one/harmony/crypto/hash"
	"github.com/harmony-one/harmony/internal/utils"
)

// FBFTMessage is the record of pbft messages received by a node during FBFT process
type FBFTMessage struct {
	MessageType        msg_pb.MessageType
	ViewID             uint64
	BlockNum           uint64
	BlockHash          common.Hash
	Block              []byte
	SenderPubkeys      []*bls.PublicKeyWrapper
	SenderPubkeyBitmap []byte
	LeaderPubkey       *bls.PublicKeyWrapper
	Payload            []byte
	ViewchangeSig      *bls_core.Sign
	ViewidSig          *bls_core.Sign
	M2AggSig           *bls_core.Sign
	M2Bitmap           *bls_cosi.Mask
	M3AggSig           *bls_core.Sign
	M3Bitmap           *bls_cosi.Mask
}

// String ..
func (m *FBFTMessage) String() string {
	sender := ""
	for _, key := range m.SenderPubkeys {
		if sender == "" {
			sender = key.Bytes.Hex()
		} else {
			sender = sender + ";" + key.Bytes.Hex()
		}
	}
	leader := ""
	if m.LeaderPubkey != nil {
		leader = m.LeaderPubkey.Bytes.Hex()
	}
	return fmt.Sprintf(
		"[Type:%s ViewID:%d Num:%d BlockHash:%s Sender:%s Leader:%s]",
		m.MessageType.String(),
		m.ViewID,
		m.BlockNum,
		m.BlockHash.Hex(),
		sender,
		leader,
	)
}

// Hash return the hash of the FBFT message for the unique identifier for FBFT message
func (m *FBFTMessage) Hash() common.Hash {
	return hash.FromRLPNew256(m)
}

// FBFTLog represents the log stored by a node during FBFT process
type FBFTLog struct {
	blocks         map[common.Hash]*types.Block // store blocks received in FBFT
	messages       map[common.Hash]*FBFTMessage // store messages received in FBFT
	verifiedBlocks map[common.Hash]struct{}     // store block hashes for blocks that has already been verified

	lock sync.Mutex
}

// NewFBFTLog returns new instance of FBFTLog
func NewFBFTLog() *FBFTLog {
	pbftLog := FBFTLog{
		blocks:         make(map[common.Hash]*types.Block),
		messages:       make(map[common.Hash]*FBFTMessage),
		verifiedBlocks: make(map[common.Hash]struct{}),
	}
	return &pbftLog
}

// AddBlock add a new block into the log
func (log *FBFTLog) AddBlock(block *types.Block) {
	log.lock.Lock()
	defer log.lock.Unlock()

	log.blocks[block.Hash()] = block
}

// MarkBlockVerified marks the block as verified
func (log *FBFTLog) MarkBlockVerified(block *types.Block) {
	log.lock.Lock()
	defer log.lock.Unlock()

	log.verifiedBlocks[block.Hash()] = struct{}{}
}

// IsBlockVerified checks whether the block is verified
func (log *FBFTLog) IsBlockVerified(block *types.Block) bool {
	log.lock.Lock()
	defer log.lock.Unlock()

	_, exist := log.verifiedBlocks[block.Hash()]
	return exist
}

// GetBlockByHash returns the block matches the given block hash
func (log *FBFTLog) GetBlockByHash(hash common.Hash) *types.Block {
	log.lock.Lock()
	defer log.lock.Unlock()

	return log.blocks[hash]
}

// GetBlocksByNumber returns the blocks match the given block number
func (log *FBFTLog) GetBlocksByNumber(number uint64) []*types.Block {
	log.lock.Lock()
	defer log.lock.Unlock()

	var blocks []*types.Block
	for _, block := range log.blocks {
		if block.NumberU64() == number {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

// DeleteBlocksLessThan deletes blocks less than given block number
func (log *FBFTLog) DeleteBlocksLessThan(number uint64) {
	log.lock.Lock()
	defer log.lock.Unlock()

	for h, block := range log.blocks {
		if block.NumberU64() < number {
			delete(log.blocks, h)
			delete(log.verifiedBlocks, h)
		}
	}
}

// DeleteBlockByNumber deletes block of specific number
func (log *FBFTLog) DeleteBlockByNumber(number uint64) {
	log.lock.Lock()
	defer log.lock.Unlock()

	for h, block := range log.blocks {
		if block.NumberU64() == number {
			delete(log.blocks, h)
			delete(log.verifiedBlocks, h)
		}
	}
}

// DeleteMessagesLessThan deletes messages less than given block number
func (log *FBFTLog) DeleteMessagesLessThan(number uint64) {
	log.lock.Lock()
	defer log.lock.Unlock()

	for h, msg := range log.messages {
		if msg.BlockNum < number {
			delete(log.messages, h)
		}
	}
}

// AddMessage adds a pbft message into the log
func (log *FBFTLog) AddMessage(msg *FBFTMessage) {
	log.lock.Lock()
	defer log.lock.Unlock()

	log.messages[msg.Hash()] = msg
}

// GetMessagesByTypeSeqViewHash returns pbft messages with matching type, blockNum, viewID and blockHash
func (log *FBFTLog) GetMessagesByTypeSeqViewHash(typ msg_pb.MessageType, blockNum uint64, viewID uint64, blockHash common.Hash) []*FBFTMessage {
	log.lock.Lock()
	defer log.lock.Unlock()

	var found []*FBFTMessage
	for _, msg := range log.messages {
		if msg.MessageType == typ && msg.BlockNum == blockNum && msg.ViewID == viewID && msg.BlockHash == blockHash {
			found = append(found, msg)
		}
	}
	return found
}

// GetMessagesByTypeSeq returns pbft messages with matching type, blockNum
func (log *FBFTLog) GetMessagesByTypeSeq(typ msg_pb.MessageType, blockNum uint64) []*FBFTMessage {
	log.lock.Lock()
	defer log.lock.Unlock()

	var found []*FBFTMessage
	for _, msg := range log.messages {
		if msg.MessageType == typ && msg.BlockNum == blockNum {
			found = append(found, msg)
		}
	}
	return found
}

// GetMessagesByTypeSeqHash returns pbft messages with matching type, blockNum
func (log *FBFTLog) GetMessagesByTypeSeqHash(typ msg_pb.MessageType, blockNum uint64, blockHash common.Hash) []*FBFTMessage {
	log.lock.Lock()
	defer log.lock.Unlock()

	var found []*FBFTMessage
	for _, msg := range log.messages {
		if msg.MessageType == typ && msg.BlockNum == blockNum && msg.BlockHash == blockHash {
			found = append(found, msg)
		}
	}
	return found
}

// HasMatchingAnnounce returns whether the log contains announce type message with given blockNum, blockHash
func (log *FBFTLog) HasMatchingAnnounce(blockNum uint64, blockHash common.Hash) bool {
	found := log.GetMessagesByTypeSeqHash(msg_pb.MessageType_ANNOUNCE, blockNum, blockHash)
	return len(found) >= 1
}

// HasMatchingViewAnnounce returns whether the log contains announce type message with given blockNum, viewID and blockHash
func (log *FBFTLog) HasMatchingViewAnnounce(blockNum uint64, viewID uint64, blockHash common.Hash) bool {
	found := log.GetMessagesByTypeSeqViewHash(msg_pb.MessageType_ANNOUNCE, blockNum, viewID, blockHash)
	return len(found) >= 1
}

// HasMatchingPrepared returns whether the log contains prepared message with given blockNum, viewID and blockHash
func (log *FBFTLog) HasMatchingPrepared(blockNum uint64, blockHash common.Hash) bool {
	found := log.GetMessagesByTypeSeqHash(msg_pb.MessageType_PREPARED, blockNum, blockHash)
	return len(found) >= 1
}

// HasMatchingViewPrepared returns whether the log contains prepared message with given blockNum, viewID and blockHash
func (log *FBFTLog) HasMatchingViewPrepared(blockNum uint64, viewID uint64, blockHash common.Hash) bool {
	found := log.GetMessagesByTypeSeqViewHash(msg_pb.MessageType_PREPARED, blockNum, viewID, blockHash)
	return len(found) >= 1
}

// GetMessagesByTypeSeqView returns pbft messages with matching type, blockNum and viewID
func (log *FBFTLog) GetMessagesByTypeSeqView(typ msg_pb.MessageType, blockNum uint64, viewID uint64) []*FBFTMessage {
	log.lock.Lock()
	defer log.lock.Unlock()

	var found []*FBFTMessage
	for _, msg := range log.messages {
		if msg.MessageType != typ || msg.BlockNum != blockNum || msg.ViewID != viewID {
			continue
		}
		found = append(found, msg)
	}
	return found
}

// FindMessageByMaxViewID returns the message that has maximum ViewID
func (log *FBFTLog) FindMessageByMaxViewID(msgs []*FBFTMessage) *FBFTMessage {
	if len(msgs) == 0 {
		return nil
	}
	maxIdx := -1
	maxViewID := uint64(0)
	for k, v := range msgs {
		if v.ViewID >= maxViewID {
			maxIdx = k
			maxViewID = v.ViewID
		}
	}
	return msgs[maxIdx]
}

// ParseFBFTMessage parses FBFT message into FBFTMessage structure
func (consensus *Consensus) ParseFBFTMessage(msg *msg_pb.Message) (*FBFTMessage, error) {
	// TODO Have this do sanity checks on the message please
	pbftMsg := FBFTMessage{}
	pbftMsg.MessageType = msg.GetType()
	consensusMsg := msg.GetConsensus()
	pbftMsg.ViewID = consensusMsg.ViewId
	pbftMsg.BlockNum = consensusMsg.BlockNum
	copy(pbftMsg.BlockHash[:], consensusMsg.BlockHash[:])
	pbftMsg.Payload = make([]byte, len(consensusMsg.Payload))
	copy(pbftMsg.Payload[:], consensusMsg.Payload[:])
	pbftMsg.Block = make([]byte, len(consensusMsg.Block))
	copy(pbftMsg.Block[:], consensusMsg.Block[:])
	pbftMsg.SenderPubkeyBitmap = make([]byte, len(consensusMsg.SenderPubkeyBitmap))
	copy(pbftMsg.SenderPubkeyBitmap[:], consensusMsg.SenderPubkeyBitmap[:])

	if len(consensusMsg.SenderPubkey) != 0 {
		// If SenderPubKey is populated, treat it as a single key message
		pubKey, err := bls_cosi.BytesToBLSPublicKey(consensusMsg.SenderPubkey)
		if err != nil {
			return nil, err
		}
		pbftMsg.SenderPubkeys = []*bls.PublicKeyWrapper{{Object: pubKey}}
		copy(pbftMsg.SenderPubkeys[0].Bytes[:], consensusMsg.SenderPubkey[:])
	} else {
		// else, it should be a multi-key message where the bitmap is populated
		consensus.multiSigMutex.RLock()
		pubKeys, err := consensus.multiSigBitmap.GetSignedPubKeysFromBitmap(pbftMsg.SenderPubkeyBitmap)
		consensus.multiSigMutex.RUnlock()
		if err != nil {
			return nil, err
		}
		pbftMsg.SenderPubkeys = pubKeys
	}

	return &pbftMsg, nil
}

// ParseViewChangeMessage parses view change message into FBFTMessage structure
func ParseViewChangeMessage(msg *msg_pb.Message) (*FBFTMessage, error) {
	pbftMsg := FBFTMessage{}
	pbftMsg.MessageType = msg.GetType()
	if pbftMsg.MessageType != msg_pb.MessageType_VIEWCHANGE {
		return nil, fmt.Errorf("ParseViewChangeMessage: incorrect message type %s", pbftMsg.MessageType)
	}

	vcMsg := msg.GetViewchange()
	pbftMsg.ViewID = vcMsg.ViewId
	pbftMsg.BlockNum = vcMsg.BlockNum
	pbftMsg.Block = make([]byte, len(vcMsg.PreparedBlock))
	copy(pbftMsg.Block[:], vcMsg.PreparedBlock[:])
	pbftMsg.Payload = make([]byte, len(vcMsg.Payload))
	copy(pbftMsg.Payload[:], vcMsg.Payload[:])

	pubKey, err := bls_cosi.BytesToBLSPublicKey(vcMsg.SenderPubkey)
	if err != nil {
		utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to parse senderpubkey")
		return nil, err
	}
	leaderKey, err := bls_cosi.BytesToBLSPublicKey(vcMsg.LeaderPubkey)
	if err != nil {
		utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to parse leaderpubkey")
		return nil, err
	}

	vcSig := bls_core.Sign{}
	err = vcSig.Deserialize(vcMsg.ViewchangeSig)
	if err != nil {
		utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to deserialize the viewchange signature")
		return nil, err
	}

	vcSig1 := bls_core.Sign{}
	err = vcSig1.Deserialize(vcMsg.ViewidSig)
	if err != nil {
		utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to deserialize the viewid signature")
		return nil, err
	}

	pbftMsg.SenderPubkeys = []*bls.PublicKeyWrapper{{Object: pubKey}}
	copy(pbftMsg.SenderPubkeys[0].Bytes[:], vcMsg.SenderPubkey[:])
	pbftMsg.LeaderPubkey = &bls.PublicKeyWrapper{Object: leaderKey}
	copy(pbftMsg.LeaderPubkey.Bytes[:], vcMsg.LeaderPubkey[:])
	pbftMsg.ViewchangeSig = &vcSig
	pbftMsg.ViewidSig = &vcSig1
	return &pbftMsg, nil
}

// ParseNewViewMessage parses new view message into FBFTMessage structure
func (consensus *Consensus) ParseNewViewMessage(msg *msg_pb.Message) (*FBFTMessage, error) {
	FBFTMsg := FBFTMessage{}
	FBFTMsg.MessageType = msg.GetType()

	if FBFTMsg.MessageType != msg_pb.MessageType_NEWVIEW {
		return nil, fmt.Errorf("ParseNewViewMessage: incorrect message type %s", FBFTMsg.MessageType)
	}

	vcMsg := msg.GetViewchange()
	FBFTMsg.ViewID = vcMsg.ViewId
	FBFTMsg.BlockNum = vcMsg.BlockNum
	FBFTMsg.Payload = make([]byte, len(vcMsg.Payload))
	copy(FBFTMsg.Payload[:], vcMsg.Payload[:])
	FBFTMsg.Block = make([]byte, len(vcMsg.PreparedBlock))
	copy(FBFTMsg.Block[:], vcMsg.PreparedBlock[:])

	pubKey, err := bls_cosi.BytesToBLSPublicKey(vcMsg.SenderPubkey)
	if err != nil {
		utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to parse senderpubkey")
		return nil, err
	}

	FBFTMsg.SenderPubkeys = []*bls.PublicKeyWrapper{{Object: pubKey}}
	copy(FBFTMsg.SenderPubkeys[0].Bytes[:], vcMsg.SenderPubkey[:])

	members := consensus.Decider.Participants()
	if len(vcMsg.M3Aggsigs) > 0 {
		m3Sig := bls_core.Sign{}
		err = m3Sig.Deserialize(vcMsg.M3Aggsigs)
		if err != nil {
			utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to deserialize the multi signature for M3 viewID signature")
			return nil, err
		}
		m3mask, err := bls_cosi.NewMask(members, nil)
		if err != nil {
			utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to create mask for multi signature")
			return nil, err
		}
		m3mask.SetMask(vcMsg.M3Bitmap)
		FBFTMsg.M3AggSig = &m3Sig
		FBFTMsg.M3Bitmap = m3mask
	}

	if len(vcMsg.M2Aggsigs) > 0 {
		m2Sig := bls_core.Sign{}
		err = m2Sig.Deserialize(vcMsg.M2Aggsigs)
		if err != nil {
			utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to deserialize the multi signature for M2 aggregated signature")
			return nil, err
		}
		m2mask, err := bls_cosi.NewMask(members, nil)
		if err != nil {
			utils.Logger().Warn().Err(err).Msg("ParseViewChangeMessage failed to create mask for multi signature")
			return nil, err
		}
		m2mask.SetMask(vcMsg.M2Bitmap)
		FBFTMsg.M2AggSig = &m2Sig
		FBFTMsg.M2Bitmap = m2mask
	}

	return &FBFTMsg, nil
}

var errFBFTLogNotFound = errors.New("FBFT log not found")

// GetCommittedBlockAndMsgsFromNumber get committed block and message starting from block number bn.
func (log *FBFTLog) GetCommittedBlockAndMsgsFromNumber(bn uint64, logger *zerolog.Logger) (*types.Block, *FBFTMessage, error) {
	msgs := log.GetMessagesByTypeSeq(
		msg_pb.MessageType_COMMITTED, bn,
	)
	if len(msgs) == 0 {
		return nil, nil, errFBFTLogNotFound
	}
	if len(msgs) > 1 {
		logger.Error().Int("numMsgs", len(msgs)).Err(errors.New("DANGER!!! multiple COMMITTED message in PBFT log observed"))
	}
	for i := range msgs {
		block := log.GetBlockByHash(msgs[i].BlockHash)
		if block == nil {
			logger.Debug().
				Uint64("blockNum", msgs[i].BlockNum).
				Uint64("viewID", msgs[i].ViewID).
				Str("blockHash", msgs[i].BlockHash.Hex()).
				Err(errors.New("failed finding a matching block for committed message"))
			continue
		}
		return block, msgs[i], nil
	}
	return nil, nil, errFBFTLogNotFound
}

// PruneCacheBeforeBlock prune all blocks before bn
func (log *FBFTLog) PruneCacheBeforeBlock(bn uint64) {
	log.DeleteBlocksLessThan(bn - 1)
	log.DeleteMessagesLessThan(bn - 1)
}
