package v2

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/harmony-one/harmony/block"
	"github.com/harmony-one/harmony/core/types"
	internal_common "github.com/harmony-one/harmony/internal/common"
	"github.com/harmony-one/harmony/numeric"
	rpc_common "github.com/harmony-one/harmony/rpc/common"
	"github.com/harmony-one/harmony/shard"
	staking "github.com/harmony-one/harmony/staking/types"
)

// RPCBlock represents a block that will serialize to the RPC representation of a block
type RPCBlock struct {
	Number           *big.Int         `json:"number"`
	ViewID           *big.Int         `json:"viewID"`
	Epoch            *big.Int         `json:"epoch"`
	Hash             common.Hash      `json:"hash"`
	ParentHash       common.Hash      `json:"parentHash"`
	Nonce            types.BlockNonce `json:"nonce"`
	MixHash          common.Hash      `json:"mixHash"`
	LogsBloom        ethtypes.Bloom   `json:"logsBloom"`
	StateRoot        common.Hash      `json:"stateRoot"`
	Miner            string           `json:"miner"`
	Difficulty       *big.Int         `json:"difficulty"`
	ExtraData        []byte           `json:"extraData"`
	Size             uint64           `json:"size"`
	GasLimit         uint64           `json:"gasLimit"`
	GasUsed          uint64           `json:"gasUsed"`
	Timestamp        *big.Int         `json:"timestamp"`
	TransactionsRoot common.Hash      `json:"transactionsRoot"`
	ReceiptsRoot     common.Hash      `json:"receiptsRoot"`
	Transactions     []interface{}    `json:"transactions"`
	StakingTxs       []interface{}    `json:"stakingTxs"`
	Uncles           []common.Hash    `json:"uncles"`
	TotalDifficulty  *big.Int         `json:"totalDifficulty"`
	Signers          []string         `json:"signers"`
}

// RPCBlockWithTxHash represents a block that will serialize to the RPC representation of a block
// having ONLY transaction hashes in the Transaction & Staking transaction fields.
type RPCBlockWithTxHash struct {
	Number           *big.Int       `json:"number"`
	ViewID           *big.Int       `json:"viewID"`
	Epoch            *big.Int       `json:"epoch"`
	Hash             common.Hash    `json:"hash"`
	ParentHash       common.Hash    `json:"parentHash"`
	Nonce            uint64         `json:"nonce"`
	MixHash          common.Hash    `json:"mixHash"`
	LogsBloom        ethtypes.Bloom `json:"logsBloom"`
	StateRoot        common.Hash    `json:"stateRoot"`
	Miner            string         `json:"miner"`
	Difficulty       uint64         `json:"difficulty"`
	ExtraData        hexutil.Bytes  `json:"extraData"`
	Size             uint64         `json:"size"`
	GasLimit         uint64         `json:"gasLimit"`
	GasUsed          uint64         `json:"gasUsed"`
	Timestamp        *big.Int       `json:"timestamp"`
	TransactionsRoot common.Hash    `json:"transactionsRoot"`
	ReceiptsRoot     common.Hash    `json:"receiptsRoot"`
	Uncles           []common.Hash  `json:"uncles"`
	Transactions     []common.Hash  `json:"transactions"`
	StakingTxs       []common.Hash  `json:"stakingTransactions"`
	Signers          []string       `json:"signers,omitempty"`
}

// RPCBlockWithFullTx represents a block that will serialize to the RPC representation of a block
// having FULL transactions in the Transaction & Staking transaction fields.
type RPCBlockWithFullTx struct {
	Number           *big.Int                 `json:"number"`
	ViewID           *big.Int                 `json:"viewID"`
	Epoch            *big.Int                 `json:"epoch"`
	Hash             common.Hash              `json:"hash"`
	ParentHash       common.Hash              `json:"parentHash"`
	Nonce            uint64                   `json:"nonce"`
	MixHash          common.Hash              `json:"mixHash"`
	LogsBloom        ethtypes.Bloom           `json:"logsBloom"`
	StateRoot        common.Hash              `json:"stateRoot"`
	Miner            string                   `json:"miner"`
	Difficulty       uint64                   `json:"difficulty"`
	ExtraData        hexutil.Bytes            `json:"extraData"`
	Size             uint64                   `json:"size"`
	GasLimit         uint64                   `json:"gasLimit"`
	GasUsed          uint64                   `json:"gasUsed"`
	Timestamp        *big.Int                 `json:"timestamp"`
	TransactionsRoot common.Hash              `json:"transactionsRoot"`
	ReceiptsRoot     common.Hash              `json:"receiptsRoot"`
	Uncles           []common.Hash            `json:"uncles"`
	Transactions     []*RPCTransaction        `json:"transactions"`
	StakingTxs       []*RPCStakingTransaction `json:"stakingTransactions"`
	Signers          []string                 `json:"signers,omitempty"`
}

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        common.Hash   `json:"blockHash"`
	BlockNumber      *big.Int      `json:"blockNumber"`
	From             string        `json:"from"`
	Timestamp        uint64        `json:"timestamp"`
	Gas              uint64        `json:"gas"`
	GasPrice         *big.Int      `json:"gasPrice"`
	Hash             common.Hash   `json:"hash"`
	Input            hexutil.Bytes `json:"input"`
	Nonce            uint64        `json:"nonce"`
	To               string        `json:"to"`
	TransactionIndex uint64        `json:"transactionIndex"`
	Value            *big.Int      `json:"value"`
	ShardID          uint32        `json:"shardID"`
	ToShardID        uint32        `json:"toShardID"`
	V                *hexutil.Big  `json:"v"`
	R                *hexutil.Big  `json:"r"`
	S                *hexutil.Big  `json:"s"`
}

// RPCStakingTransaction represents a transaction that will serialize to the RPC representation of a staking transaction
type RPCStakingTransaction struct {
	BlockHash        common.Hash            `json:"blockHash"`
	BlockNumber      *big.Int               `json:"blockNumber"`
	From             string                 `json:"from"`
	Timestamp        uint64                 `json:"timestamp"`
	Gas              uint64                 `json:"gas"`
	GasPrice         *big.Int               `json:"gasPrice"`
	Hash             common.Hash            `json:"hash"`
	Nonce            uint64                 `json:"nonce"`
	TransactionIndex uint64                 `json:"transactionIndex"`
	V                *hexutil.Big           `json:"v"`
	R                *hexutil.Big           `json:"r"`
	S                *hexutil.Big           `json:"s"`
	Type             string                 `json:"type"`
	Msg              map[string]interface{} `json:"msg"`
}

// RPCCXReceipt represents a CXReceipt that will serialize to the RPC representation of a CXReceipt
type RPCCXReceipt struct {
	BlockHash   common.Hash `json:"blockHash"`
	BlockNumber *big.Int    `json:"blockNumber"`
	TxHash      common.Hash `json:"hash"`
	From        string      `json:"from"`
	To          string      `json:"to"`
	ShardID     uint32      `json:"shardID"`
	ToShardID   uint32      `json:"toShardID"`
	Amount      *big.Int    `json:"value"`
}

// HeaderInformation represents the latest consensus information
type HeaderInformation struct {
	BlockHash        common.Hash       `json:"blockHash"`
	BlockNumber      uint64            `json:"blockNumber"`
	ShardID          uint32            `json:"shardID"`
	Leader           string            `json:"leader"`
	ViewID           uint64            `json:"viewID"`
	Epoch            uint64            `json:"epoch"`
	Timestamp        string            `json:"timestamp"`
	UnixTime         uint64            `json:"unixtime"`
	LastCommitSig    string            `json:"lastCommitSig"`
	LastCommitBitmap string            `json:"lastCommitBitmap"`
	CrossLinks       *types.CrossLinks `json:"crossLinks,omitempty"`
}

// RPCDelegation represents a particular delegation to a validator
type RPCDelegation struct {
	ValidatorAddress string            `json:"validator_address"`
	DelegatorAddress string            `json:"delegator_address"`
	Amount           *big.Int          `json:"amount"`
	Reward           *big.Int          `json:"reward"`
	Undelegations    []RPCUndelegation `json:"Undelegations"`
}

// RPCUndelegation represents one undelegation entry
type RPCUndelegation struct {
	Amount *big.Int
	Epoch  *big.Int
}

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From     *common.Address `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Data     *hexutil.Bytes  `json:"data"`
}

// StakingNetworkInfo returns global staking info.
type StakingNetworkInfo struct {
	TotalSupply       numeric.Dec `json:"total-supply"`
	CirculatingSupply numeric.Dec `json:"circulating-supply"`
	EpochLastBlock    uint64      `json:"epoch-last-block"`
	TotalStaking      *big.Int    `json:"total-staking"`
	MedianRawStake    numeric.Dec `json:"median-raw-stake"`
}

func newHeaderInformation(header *block.Header, leader string) *HeaderInformation {
	if header == nil {
		return nil
	}

	result := &HeaderInformation{
		BlockHash:        header.Hash(),
		BlockNumber:      header.Number().Uint64(),
		ShardID:          header.ShardID(),
		Leader:           leader,
		ViewID:           header.ViewID().Uint64(),
		Epoch:            header.Epoch().Uint64(),
		UnixTime:         header.Time().Uint64(),
		Timestamp:        time.Unix(header.Time().Int64(), 0).UTC().String(),
		LastCommitBitmap: hex.EncodeToString(header.LastCommitBitmap()),
	}

	sig := header.LastCommitSignature()
	result.LastCommitSig = hex.EncodeToString(sig[:])

	if header.ShardID() == shard.BeaconChainShardID {
		decodedCrossLinks := &types.CrossLinks{}
		err := rlp.DecodeBytes(header.CrossLinks(), decodedCrossLinks)
		if err != nil {
			result.CrossLinks = &types.CrossLinks{}
		} else {
			result.CrossLinks = decodedCrossLinks
		}
	}

	return result
}

// newRPCCXReceipt returns a CXReceipt that will serialize to the RPC representation
func newRPCCXReceipt(cx *types.CXReceipt, blockHash common.Hash, blockNumber uint64) *RPCCXReceipt {
	result := &RPCCXReceipt{
		BlockHash: blockHash,
		TxHash:    cx.TxHash,
		Amount:    cx.Amount,
		ShardID:   cx.ShardID,
		ToShardID: cx.ToShardID,
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*big.Int)(new(big.Int).SetUint64(blockNumber))
	}

	fromAddr, err := internal_common.AddressToBech32(cx.From)
	if err != nil {
		return nil
	}
	toAddr := ""
	if cx.To != nil {
		if toAddr, err = internal_common.AddressToBech32(*cx.To); err != nil {
			return nil
		}
	}
	result.From = fromAddr
	result.To = toAddr

	return result
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
// Note that all txs on Harmony are replay protected (post EIP155 epoch).
func newRPCTransaction(
	tx *types.Transaction, blockHash common.Hash,
	blockNumber uint64, timestamp uint64, index uint64,
) *RPCTransaction {
	from, err := tx.SenderAddress()
	if err != nil {
		return nil
	}
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		Gas:       tx.Gas(),
		GasPrice:  tx.GasPrice(),
		Hash:      tx.Hash(),
		Input:     hexutil.Bytes(tx.Data()),
		Nonce:     tx.Nonce(),
		Value:     tx.Value(),
		ShardID:   tx.ShardID(),
		ToShardID: tx.ToShardID(),
		Timestamp: timestamp,
		V:         (*hexutil.Big)(v),
		R:         (*hexutil.Big)(r),
		S:         (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*big.Int)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = index
	}

	fromAddr, err := internal_common.AddressToBech32(from)
	if err != nil {
		return nil
	}
	toAddr := ""

	if tx.To() != nil {
		if toAddr, err = internal_common.AddressToBech32(*tx.To()); err != nil {
			return nil
		}
		result.From = fromAddr
	} else {
		result.From = strings.ToLower(from.Hex())
	}
	result.To = toAddr

	return result
}

// newRPCStakingTransaction returns a staking transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCStakingTransaction(
	tx *staking.StakingTransaction, blockHash common.Hash,
	blockNumber uint64, timestamp uint64, index uint64,
) *RPCStakingTransaction {
	from, err := tx.SenderAddress()
	if err != nil {
		return nil
	}
	v, r, s := tx.RawSignatureValues()

	fields := make(map[string]interface{})

	switch tx.StakingType() {
	case staking.DirectiveCreateValidator:
		rawMsg, err := staking.RLPDecodeStakeMsg(tx.Data(), staking.DirectiveCreateValidator)
		if err != nil {
			return nil
		}
		msg, ok := rawMsg.(*staking.CreateValidator)
		if !ok {
			return nil
		}
		validatorAddress, err := internal_common.AddressToBech32(msg.ValidatorAddress)
		if err != nil {
			return nil
		}
		fields = map[string]interface{}{
			"validatorAddress":   validatorAddress,
			"name":               msg.Description.Name,
			"commissionRate":     msg.CommissionRates.Rate.Int,
			"maxCommissionRate":  msg.CommissionRates.MaxRate.Int,
			"maxChangeRate":      msg.CommissionRates.MaxChangeRate.Int,
			"minSelfDelegation":  msg.MinSelfDelegation,
			"maxTotalDelegation": msg.MaxTotalDelegation,
			"amount":             msg.Amount,
			"website":            msg.Description.Website,
			"identity":           msg.Description.Identity,
			"securityContact":    msg.Description.SecurityContact,
			"details":            msg.Description.Details,
			"slotPubKeys":        msg.SlotPubKeys,
		}
	case staking.DirectiveEditValidator:
		rawMsg, err := staking.RLPDecodeStakeMsg(tx.Data(), staking.DirectiveEditValidator)
		if err != nil {
			return nil
		}
		msg, ok := rawMsg.(*staking.EditValidator)
		if !ok {
			return nil
		}
		validatorAddress, err := internal_common.AddressToBech32(msg.ValidatorAddress)
		if err != nil {
			return nil
		}
		// Edit validators txs need not have commission rates to edit
		commissionRate := &big.Int{}
		if msg.CommissionRate != nil {
			commissionRate = msg.CommissionRate.Int
		}
		fields = map[string]interface{}{
			"validatorAddress":   validatorAddress,
			"commisionRate":      commissionRate,
			"name":               msg.Description.Name,
			"minSelfDelegation":  msg.MinSelfDelegation,
			"maxTotalDelegation": msg.MaxTotalDelegation,
			"website":            msg.Description.Website,
			"identity":           msg.Description.Identity,
			"securityContact":    msg.Description.SecurityContact,
			"details":            msg.Description.Details,
			"slotPubKeyToAdd":    msg.SlotKeyToAdd,
			"slotPubKeyToRemove": msg.SlotKeyToRemove,
		}
	case staking.DirectiveCollectRewards:
		rawMsg, err := staking.RLPDecodeStakeMsg(tx.Data(), staking.DirectiveCollectRewards)
		if err != nil {
			return nil
		}
		msg, ok := rawMsg.(*staking.CollectRewards)
		if !ok {
			return nil
		}
		delegatorAddress, err := internal_common.AddressToBech32(msg.DelegatorAddress)
		if err != nil {
			return nil
		}
		fields = map[string]interface{}{
			"delegatorAddress": delegatorAddress,
		}
	case staking.DirectiveDelegate:
		rawMsg, err := staking.RLPDecodeStakeMsg(tx.Data(), staking.DirectiveDelegate)
		if err != nil {
			return nil
		}
		msg, ok := rawMsg.(*staking.Delegate)
		if !ok {
			return nil
		}
		delegatorAddress, err := internal_common.AddressToBech32(msg.DelegatorAddress)
		if err != nil {
			return nil
		}
		validatorAddress, err := internal_common.AddressToBech32(msg.ValidatorAddress)
		if err != nil {
			return nil
		}
		fields = map[string]interface{}{
			"delegatorAddress": delegatorAddress,
			"validatorAddress": validatorAddress,
			"amount":           msg.Amount,
		}
	case staking.DirectiveUndelegate:
		rawMsg, err := staking.RLPDecodeStakeMsg(tx.Data(), staking.DirectiveUndelegate)
		if err != nil {
			return nil
		}
		msg, ok := rawMsg.(*staking.Undelegate)
		if !ok {
			return nil
		}
		delegatorAddress, err := internal_common.AddressToBech32(msg.DelegatorAddress)
		if err != nil {
			return nil
		}
		validatorAddress, err := internal_common.AddressToBech32(msg.ValidatorAddress)
		if err != nil {
			return nil
		}
		fields = map[string]interface{}{
			"delegatorAddress": delegatorAddress,
			"validatorAddress": validatorAddress,
			"amount":           msg.Amount,
		}
	}

	result := &RPCStakingTransaction{
		Gas:       tx.Gas(),
		GasPrice:  tx.GasPrice(),
		Hash:      tx.Hash(),
		Nonce:     tx.Nonce(),
		Timestamp: timestamp,
		V:         (*hexutil.Big)(v),
		R:         (*hexutil.Big)(r),
		S:         (*hexutil.Big)(s),
		Type:      tx.StakingType().String(),
		Msg:       fields,
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = new(big.Int).SetUint64(blockNumber)
		result.TransactionIndex = index
	}

	fromAddr, err := internal_common.AddressToBech32(from)
	if err != nil {
		return nil
	}
	result.From = fromAddr

	return result
}

// NewRPCBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func NewRPCBlock(b *types.Block, blockArgs *rpc_common.BlockArgs, leader string) (interface{}, error) {
	if blockArgs.FullTx {
		return NewRPCBlockWithFullTx(b, blockArgs, leader)
	}
	return NewRPCBlockWithTxHash(b, blockArgs, leader)
}

// NewRPCBlockWithTxHash ..
func NewRPCBlockWithTxHash(
	b *types.Block, blockArgs *rpc_common.BlockArgs, leader string,
) (*RPCBlockWithTxHash, error) {
	if blockArgs.FullTx {
		return nil, fmt.Errorf("block args specifies full tx, but requested RPC block with only tx hash")
	}

	head := b.Header()
	blk := &RPCBlockWithTxHash{
		Number:           head.Number(),
		ViewID:           head.ViewID(),
		Epoch:            head.Epoch(),
		Hash:             b.Hash(),
		ParentHash:       head.ParentHash(),
		Nonce:            0, // Remove this because we don't have it in our header
		MixHash:          head.MixDigest(),
		LogsBloom:        head.Bloom(),
		StateRoot:        head.Root(),
		Miner:            leader,
		Difficulty:       0, // Remove this because we don't have it in our header
		ExtraData:        hexutil.Bytes(head.Extra()),
		Size:             uint64(b.Size()),
		GasLimit:         head.GasLimit(),
		GasUsed:          head.GasUsed(),
		Timestamp:        head.Time(),
		TransactionsRoot: head.TxHash(),
		ReceiptsRoot:     head.ReceiptHash(),
		Uncles:           []common.Hash{},
		Transactions:     []common.Hash{},
		StakingTxs:       []common.Hash{},
	}

	for _, tx := range b.Transactions() {
		blk.Transactions = append(blk.Transactions, tx.Hash())
	}

	if blockArgs.InclStaking {
		for _, stx := range b.StakingTransactions() {
			blk.StakingTxs = append(blk.StakingTxs, stx.Hash())
		}
	}

	if blockArgs.WithSigners {
		blk.Signers = blockArgs.Signers
	}
	return blk, nil
}

// NewRPCBlockWithFullTx ..
func NewRPCBlockWithFullTx(
	b *types.Block, blockArgs *rpc_common.BlockArgs, leader string,
) (*RPCBlockWithFullTx, error) {
	if !blockArgs.FullTx {
		return nil, fmt.Errorf("block args specifies NO full tx, but requested RPC block with full tx")
	}

	head := b.Header()
	blk := &RPCBlockWithFullTx{
		Number:           head.Number(),
		ViewID:           head.ViewID(),
		Epoch:            head.Epoch(),
		Hash:             b.Hash(),
		ParentHash:       head.ParentHash(),
		Nonce:            0, // Remove this because we don't have it in our header
		MixHash:          head.MixDigest(),
		LogsBloom:        head.Bloom(),
		StateRoot:        head.Root(),
		Miner:            leader,
		Difficulty:       0, // Remove this because we don't have it in our header
		ExtraData:        hexutil.Bytes(head.Extra()),
		Size:             uint64(b.Size()),
		GasLimit:         head.GasLimit(),
		GasUsed:          head.GasUsed(),
		Timestamp:        head.Time(),
		TransactionsRoot: head.TxHash(),
		ReceiptsRoot:     head.ReceiptHash(),
		Uncles:           []common.Hash{},
		Transactions:     []*RPCTransaction{},
		StakingTxs:       []*RPCStakingTransaction{},
	}

	for _, tx := range b.Transactions() {
		blk.Transactions = append(blk.Transactions, newRPCTransactionFromBlockHash(b, tx.Hash()))
	}

	if blockArgs.InclStaking {
		for _, stx := range b.StakingTransactions() {
			blk.StakingTxs = append(blk.StakingTxs, newRPCStakingTransactionFromBlockHash(b, stx.Hash()))
		}
	}

	if blockArgs.WithSigners {
		blk.Signers = blockArgs.Signers
	}
	return blk, nil
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCTransaction {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *types.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), b.Time().Uint64(), index)
}

// newRPCStakingTransactionFromBlockHash returns a staking transaction that will serialize to the RPC representation.
func newRPCStakingTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCStakingTransaction {
	for idx, tx := range b.StakingTransactions() {
		if tx.Hash() == hash {
			return newRPCStakingTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// newRPCStakingTransactionFromBlockIndex returns a staking transaction that will serialize to the RPC representation.
func newRPCStakingTransactionFromBlockIndex(b *types.Block, index uint64) *RPCStakingTransaction {
	txs := b.StakingTransactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCStakingTransaction(txs[index], b.Hash(), b.NumberU64(), b.Time().Uint64(), index)
}
