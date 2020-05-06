package slash

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/harmony-one/bls/ffi/go/bls"
	blockfactory "github.com/harmony-one/harmony/block/factory"
	"github.com/harmony-one/harmony/common/denominations"
	consensus_sig "github.com/harmony-one/harmony/consensus/signature"
	"github.com/harmony-one/harmony/core/state"
	"github.com/harmony-one/harmony/core/types"
	bls2 "github.com/harmony-one/harmony/crypto/bls"
	"github.com/harmony-one/harmony/internal/params"
	"github.com/harmony-one/harmony/numeric"
	"github.com/harmony-one/harmony/shard"
	"github.com/harmony-one/harmony/staking/effective"
	staking "github.com/harmony-one/harmony/staking/types"
)

var (
	fiveKOnes       = new(big.Int).Mul(big.NewInt(5000), big.NewInt(1e18))
	tenKOnes        = new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18))
	nineteenKOnes   = new(big.Int).Mul(big.NewInt(19000), big.NewInt(1e18))
	twentyKOnes     = new(big.Int).Mul(big.NewInt(20000), big.NewInt(1e18))
	twentyfiveKOnes = new(big.Int).Mul(big.NewInt(25000), big.NewInt(1e18))
	thirtyKOnes     = new(big.Int).Mul(big.NewInt(30000), big.NewInt(1e18))
	hundredKOnes    = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18))
)

var (
	scenarioTwoPercent    = defaultFundingScenario()
	scenarioEightyPercent = defaultFundingScenario()
)

const (
	// validator creation parameters
	doubleSignShardID     = 0
	doubleSignEpoch       = 3
	doubleSignBlockNumber = 37
	doubleSignViewID      = 38

	creationHeight  = 33
	lastEpochInComm = 5
)

var (
	doubleSignBlock1 = makeBlockForTest(doubleSignEpoch, 1)
	doubleSignBlock2 = makeBlockForTest(doubleSignEpoch, 2)
)

const (
	numShard        = 4
	numNodePerShard = 5

	offenderShard      = doubleSignShardID
	offenderShardIndex = 0
)

var (
	keyPairs = genKeyPairs(100)

	blsPair1, blsPair2 = keyPairs[0], keyPairs[1]
	blsPub1, blsPub2   = keyPairs[0].Pub(), keyPairs[0].Pub()

	offIndex = offenderShard*numNodePerShard + offenderShardIndex
	offAddr  = makeTestAddress(offIndex)
	offKey   = keyPairs[offIndex]
	offPub   = offKey.Pub()

	reporterAddr = makeTestAddress("reporter")
	otherAddr    = makeTestAddress("somebody")
)

var (
	testChain *fakeBlockChain

	commonCommission  staking.Commission
	commonDescription staking.Description
)

func totalSlashedExpected(slashRate float64) *big.Int {
	t := int64(50000 * slashRate)
	res := new(big.Int).Mul(big.NewInt(t), big.NewInt(denominations.One))
	return res
}

func totalSnitchRewardExpected(slashRate float64) *big.Int {
	t := int64(25000 * slashRate)
	res := new(big.Int).Mul(big.NewInt(t), big.NewInt(denominations.One))
	return res
}

func init() {
	commonDataSetup()
	{
		s := scenarioTwoPercent
		s.slashRate = 0.02
		s.result = &Application{
			TotalSlashed:      totalSlashedExpected(s.slashRate),      // big.NewInt(int64(s.slashRate * 5.0 * denominations.One)),
			TotalSnitchReward: totalSnitchRewardExpected(s.slashRate), // big.NewInt(int64(s.slashRate * 2.5 * denominations.One)),
		}
		s.snapshot = defaultTestValidatorWrapper([]shard.BLSPublicKey{offPub})
		s.current = defaultTestCurrentValidatorWrapper([]shard.BLSPublicKey{offPub})
	}
	{
		s := scenarioEightyPercent
		s.slashRate = 0.80
		s.result = &Application{
			TotalSlashed:      totalSlashedExpected(s.slashRate),      // big.NewInt(int64(s.slashRate * 5.0 * denominations.One)),
			TotalSnitchReward: totalSnitchRewardExpected(s.slashRate), // big.NewInt(int64(s.slashRate * 2.5 * denominations.One)),
		}
		s.snapshot = defaultTestValidatorWrapper([]shard.BLSPublicKey{offPub})
		s.current = defaultTestCurrentValidatorWrapper([]shard.BLSPublicKey{offPub})
	}
}

func defaultSlashRecord() Record {
	return Record{
		Evidence: Evidence{
			ConflictingVotes: ConflictingVotes{
				FirstVote:  makeVoteData(offKey, doubleSignBlock1),
				SecondVote: makeVoteData(offKey, doubleSignBlock2),
			},
			Moment: Moment{
				Epoch:   big.NewInt(doubleSignEpoch),
				ShardID: doubleSignShardID,
				Height:  doubleSignBlockNumber,
				ViewID:  doubleSignViewID,
			},
			Offender: offAddr,
		},
		Reporter: reporterAddr,
	}
}

func makeVoteData(kp blsKeyPair, block *types.Block) Vote {
	return Vote{
		SignerPubKey:    kp.Pub(),
		BlockHeaderHash: block.Hash(),
		Signature:       kp.Sign(block),
	}
}

func exampleSlashRecords() Records {
	return Records{defaultSlashRecord()}
}

func defaultStateWithAccountsApplied() *state.DB {
	st := ethdb.NewMemDatabase()
	stateHandle, _ := state.New(common.Hash{}, state.NewDatabase(st))
	for _, addr := range []common.Address{reporterAddr, offAddr, otherAddr} {
		stateHandle.CreateAccount(addr)
	}
	stateHandle.SetBalance(offAddr, big.NewInt(0).SetUint64(1994680320000000000))
	stateHandle.SetBalance(otherAddr, big.NewInt(0).SetUint64(1999975592000000000))
	return stateHandle
}

type scenario struct {
	snapshot, current *staking.ValidatorWrapper
	slashRate         float64
	result            *Application
}

func defaultFundingScenario() *scenario {
	return &scenario{
		snapshot:  nil,
		current:   nil,
		slashRate: 0.02,
		result:    nil,
	}
}

// ======== start of new test case codes ==========

func commonDataSetup() {
	commonCommission = staking.Commission{
		CommissionRates: staking.CommissionRates{
			Rate:          numeric.MustNewDecFromStr("0.167983520183826780"),
			MaxRate:       numeric.MustNewDecFromStr("0.179184469782137200"),
			MaxChangeRate: numeric.MustNewDecFromStr("0.152212761523253600"),
		},
		UpdateHeight: big.NewInt(10),
	}

	commonDescription = staking.Description{
		Name:            "someoneA",
		Identity:        "someoneB",
		Website:         "someoneC",
		SecurityContact: "someoneD",
		Details:         "someoneE",
	}
}

func makeTestAddress(item interface{}) common.Address {
	s := fmt.Sprintf("harmony.one.%s", item)
	return common.BytesToAddress([]byte(s))
}

// makeCommitteeFromKeyPairs makes a shard state for testing.
//  address is generated by makeTestAddress
//  bls key is get from the input []blsKeyPair
func makeDefaultCommitteeFromKeyPairs(kps []blsKeyPair) shard.State {
	epoch := big.NewInt(doubleSignEpoch)
	maker := newShardSlotMaker(kps)
	return makeCommitteeBySlotMaker(epoch, maker)
}

func makeCommitteeBySlotMaker(epoch *big.Int, maker shardSlotMaker) shard.State {
	sstate := shard.State{
		Epoch:  epoch,
		Shards: make([]shard.Committee, 0, int(numShard)),
	}
	for sid := uint32(0); sid != numNodePerShard; sid++ {
		sstate.Shards = append(sstate.Shards, makeShardBySlotMaker(sid, maker))
	}
	return sstate
}

func makeShardBySlotMaker(shardID uint32, maker shardSlotMaker) shard.Committee {
	cmt := shard.Committee{
		ShardID: shardID,
		Slots:   make(shard.SlotList, 0, numNodePerShard),
	}
	for nid := 0; nid != numNodePerShard; nid++ {
		cmt.Slots = append(cmt.Slots, maker.makeSlot())
	}
	return cmt
}

type shardSlotMaker struct {
	kps []blsKeyPair
	i   int
}

func newShardSlotMaker(kps []blsKeyPair) shardSlotMaker {
	return shardSlotMaker{kps, 0}
}

func (maker *shardSlotMaker) makeSlot() shard.Slot {
	s := shard.Slot{
		EcdsaAddress: makeTestAddress(maker.i),
		BLSPublicKey: maker.kps[maker.i].Pub(), // Yes, panic when not enough kps
	}
	maker.i++
	return s
}

type blsKeyPair struct {
	pri *bls.SecretKey
	pub *bls.PublicKey
}

func genKeyPairs(size int) []blsKeyPair {
	kps := make([]blsKeyPair, 0, size)
	kps = append(kps, genKeyPair())
	return kps
}

func genKeyPair() blsKeyPair {
	pri := bls2.RandPrivateKey()
	pub := pri.GetPublicKey()
	return blsKeyPair{
		pri: pri,
		pub: pub,
	}
}

func (kp blsKeyPair) Pub() shard.BLSPublicKey {
	var pub shard.BLSPublicKey
	copy(pub[:], kp.pub.Serialize())
	return pub
}

func (kp blsKeyPair) Sign(block *types.Block) []byte {
	msg := consensus_sig.ConstructCommitPayload(testChain, testChain.Config().StakingEpoch,
		block.Hash(), block.Number().Uint64(), block.Header().ViewID().Uint64())

	sig := kp.pri.SignHash(msg)

	return sig.Serialize()
}

func makeBlockForTest(epoch int64, index int) *types.Block {
	h := blockfactory.NewTestHeader()

	h.SetEpoch(big.NewInt(epoch))
	h.SetNumber(big.NewInt(doubleSignBlockNumber))
	h.SetViewID(big.NewInt(doubleSignViewID))
	h.SetRoot(common.BigToHash(big.NewInt(int64(index))))

	return types.NewBlockWithHeader(h)
}

func defaultTestValidatorWrapper(pubKeys []shard.BLSPublicKey) *staking.ValidatorWrapper {
	v := defaultTestValidator(pubKeys)
	ds := defaultTestDelegations()

	return &staking.ValidatorWrapper{
		Validator:   v,
		Delegations: ds,
	}
}

func defaultTestCurrentValidatorWrapper(pubKeys []shard.BLSPublicKey) *staking.ValidatorWrapper {
	v := defaultTestValidator(pubKeys)
	ds := defaultTestDelegationsWithUndelegates()

	return &staking.ValidatorWrapper{
		Validator:   v,
		Delegations: ds,
	}
}

// defaultTestValidator makes a valid Validator kps structure
func defaultTestValidator(pubKeys []shard.BLSPublicKey) staking.Validator {
	return staking.Validator{
		Address:              offAddr,
		SlotPubKeys:          pubKeys,
		LastEpochInCommittee: big.NewInt(lastEpochInComm),
		MinSelfDelegation:    tenKOnes,
		MaxTotalDelegation:   hundredKOnes,
		Status:               effective.Active,
		Commission:           commonCommission,
		Description:          commonDescription,
		CreationHeight:       big.NewInt(creationHeight),
	}
}

func defaultTestDelegations() staking.Delegations {
	return staking.Delegations{
		staking.Delegation{
			DelegatorAddress: offAddr,
			Amount:           twentyKOnes,
			Reward:           common.Big0,
			Undelegations:    staking.Undelegations{},
		},
		staking.Delegation{
			DelegatorAddress: otherAddr,
			Amount:           thirtyKOnes,
			Reward:           common.Big0,
			Undelegations:    staking.Undelegations{},
		},
	}
}

func defaultTestDelegationsWithUndelegates() staking.Delegations {
	return staking.Delegations{
		staking.Delegation{
			DelegatorAddress: offAddr,
			Amount:           nineteenKOnes,
			Reward:           common.Big0,
			Undelegations: staking.Undelegations{
				staking.Undelegation{
					Amount: tenKOnes,
					Epoch:  big.NewInt(doubleSignEpoch + 2),
				},
			},
		},
		staking.Delegation{
			DelegatorAddress: otherAddr,
			Amount:           fiveKOnes,
			Reward:           common.Big0,
			Undelegations: staking.Undelegations{
				staking.Undelegation{
					Amount: twentyfiveKOnes,
					Epoch:  big.NewInt(doubleSignEpoch + 2),
				},
			},
		},
	}
}

type fakeBlockChain struct {
	config         params.ChainConfig
	currentBlock   types.Block
	superCommittee shard.State
	snapshots      map[common.Address]staking.ValidatorWrapper
}

func (bc *fakeBlockChain) Config() *params.ChainConfig {
	return &bc.config
}

func (bc *fakeBlockChain) CurrentBlock() *types.Block {
	return &bc.currentBlock
}

func (bc *fakeBlockChain) ReadShardState(epoch *big.Int) (*shard.State, error) {
	if epoch.Cmp(big.NewInt(doubleSignEpoch)) != 0 {
		return nil, fmt.Errorf("epoch not expected")
	}
	return &bc.superCommittee, nil
}

func (bc *fakeBlockChain) ReadValidatorSnapshotAtEpoch(epoch *big.Int, addr common.Address) (*staking.ValidatorSnapshot, error) {
	vw, ok := bc.snapshots[addr]
	if !ok {
		return nil, errors.New("missing snapshot")
	}
	return &staking.ValidatorSnapshot{
		Validator: &vw,
		Epoch:     new(big.Int).Set(epoch),
	}, nil
}

// Simply testing serialization / deserialization of slash records working correctly
//func TestRoundTripSlashRecord(t *testing.T) {
//	slashes := exampleSlashRecords()
//	serializedA := slashes.String()
//	data, err := rlp.EncodeToBytes(slashes)
//	if err != nil {
//		t.Errorf("encoding slash records failed %s", err.Error())
//	}
//	roundTrip := Records{}
//	if err := rlp.DecodeBytes(data, &roundTrip); err != nil {
//		t.Errorf("decoding slash records failed %s", err.Error())
//	}
//	serializedB := roundTrip.String()
//	if serializedA != serializedB {
//		t.Error("rlp encode/decode round trip records failed")
//	}
//}
//
//func TestSetDifference(t *testing.T) {
//	setA, setB := exampleSlashRecords(), exampleSlashRecords()
//	additionalSlash := defaultSlashRecord()
//	additionalSlash.Evidence.Epoch.Add(additionalSlash.Evidence.Epoch, common.Big1)
//	setB = append(setB, additionalSlash)
//	diff := setA.SetDifference(setB)
//	if diff[0].Hash() != additionalSlash.Hash() {
//		t.Errorf("did not get set difference of slash")
//	}
//}

// TODO bytes used for this example are stale, need to update RLP dump
// func TestApply(t *testing.T) {
// 	slashes := exampleSlashRecords()
// {
// 	stateHandle := defaultStateWithAccountsApplied()
// 	testScenario(t, stateHandle, slashes, scenarioRealWorldSample1())
// }
// }
//
//func TestVerify(t *testing.T) {
//	stateHandle := defaultStateWithAccountsApplied()
//	// TODO: test this
//}

//func TestTwoPercentSlashed(t *testing.T) {
//	slashes := exampleSlashRecords()
//	stateHandle := defaultStateWithAccountsApplied()
//	testScenario(t, stateHandle, slashes, scenarioTwoPercent)
//}
//
//// func TestEightyPercentSlashed(t *testing.T) {
//// 	slashes := exampleSlashRecords()
//// 	stateHandle := defaultStateWithAccountsApplied()
//// 	testScenario(t, stateHandle, slashes, scenarioEightyPercent)
//// }
//
//func TestDoubleSignSlashRates(t *testing.T) {
//	for _, scenario := range doubleSignScenarios {
//		slashes := exampleSlashRecords()
//		stateHandle := defaultStateWithAccountsApplied()
//		testScenario(t, stateHandle, slashes, scenario)
//	}
//}

//func testScenario(
//	t *testing.T, stateHandle *state.DB, slashes Records, s *scenario,
//) {
//	if err := stateHandle.UpdateValidatorWrapper(
//		offAddr, s.snapshot,
//	); err != nil {
//		t.Fatalf("creation of validator failed %s", err.Error())
//	}
//
//	stateHandle.IntermediateRoot(false)
//	stateHandle.Commit(false)
//
//	if err := stateHandle.UpdateValidatorWrapper(
//		offAddr, s.current,
//	); err != nil {
//		t.Fatalf("update of validator failed %s", err.Error())
//	}
//
//	stateHandle.IntermediateRoot(false)
//	stateHandle.Commit(false)
//	// NOTE See dump.json to see what account
//	// state looks like as of this point
//
//	slashResult, err := Apply(
//		mockOutSnapshotReader{staking.ValidatorSnapshot{s.snapshot, big.NewInt(0)}},
//		stateHandle,
//		slashes,
//		numeric.MustNewDecFromStr(
//			fmt.Sprintf("%f", s.slashRate),
//		),
//	)
//
//	if err != nil {
//		t.Fatalf("rate: %v, slash application failed %s", s.slashRate, err.Error())
//	}
//
//	if sn := slashResult.TotalSlashed; sn.Cmp(
//		s.result.TotalSlashed,
//	) != 0 {
//		t.Errorf(
//			"total slash incorrect have %v want %v",
//			sn,
//			s.result.TotalSlashed,
//		)
//	}
//
//	if sn := slashResult.TotalSnitchReward; sn.Cmp(
//		s.result.TotalSnitchReward,
//	) != 0 {
//		t.Errorf(
//			"total snitch incorrect have %v want %v",
//			sn,
//			s.result.TotalSnitchReward,
//		)
//	}
//}
