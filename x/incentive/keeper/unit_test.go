package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// NewTestContext sets up a basic context with an in-memory db
func NewTestContext(requiredStoreKeys ...sdk.StoreKey) sdk.Context {
	memDB := db.NewMemDB()
	cms := store.NewCommitMultiStore(memDB)

	for _, key := range requiredStoreKeys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	cms.LoadLatestVersion()

	return sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
}

// unitTester is a wrapper around suite.Suite, with common functionality for keeper unit tests.
// It can be embedded in structs the same way as suite.Suite.
type unitTester struct {
	suite.Suite
	keeper keeper.Keeper
	ctx    sdk.Context

	cdc               *codec.Codec
	incentiveStoreKey sdk.StoreKey
}

func (suite *unitTester) SetupSuite() {
	suite.cdc = app.MakeCodec()

	suite.incentiveStoreKey = sdk.NewKVStoreKey(types.StoreKey)

}

func (suite *unitTester) SetupTest() {
	suite.ctx = NewTestContext(suite.incentiveStoreKey)
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil)
}

func (suite *unitTester) TearDownTest() {
	suite.keeper = keeper.Keeper{}
	suite.ctx = sdk.Context{}
}

func (suite *unitTester) NewKeeper(paramSubspace types.ParamSubspace, sk types.SupplyKeeper, cdpk types.CdpKeeper, hk types.HardKeeper, ak types.AccountKeeper, stk types.StakingKeeper, swk types.SwapKeeper) keeper.Keeper {
	return keeper.NewKeeper(suite.cdc, suite.incentiveStoreKey, paramSubspace, sk, cdpk, hk, ak, stk, swk)
}

func (suite *unitTester) storeGlobalBorrowIndexes(indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, i.CollateralType, i.RewardIndexes)
	}
}
func (suite *unitTester) storeGlobalSupplyIndexes(indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		suite.keeper.SetHardSupplyRewardIndexes(suite.ctx, i.CollateralType, i.RewardIndexes)
	}
}
func (suite *unitTester) storeGlobalDelegatorIndexes(multiRewardIndexes types.MultiRewardIndexes) {
	// Hardcoded to use bond denom
	multiRewardIndex, _ := multiRewardIndexes.GetRewardIndex(types.BondDenom)
	suite.keeper.SetDelegatorRewardIndexes(suite.ctx, types.BondDenom, multiRewardIndex.RewardIndexes)
}
func (suite *unitTester) storeGlobalSwapIndexes(indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		suite.keeper.SetSwapRewardIndexes(suite.ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (suite *unitTester) storeHardClaim(claim types.HardLiquidityProviderClaim) {
	suite.keeper.SetHardLiquidityProviderClaim(suite.ctx, claim)
}
func (suite *unitTester) storeDelegatorClaim(claim types.DelegatorClaim) {
	suite.keeper.SetDelegatorClaim(suite.ctx, claim)
}
func (suite *unitTester) storeSwapClaim(claim types.SwapClaim) {
	suite.keeper.SetSwapClaim(suite.ctx, claim)
}

// fakeParamSubspace is a stub paramSpace to simplify keeper unit test setup.
type fakeParamSubspace struct {
	params types.Params
}

func (subspace *fakeParamSubspace) GetParamSet(_ sdk.Context, ps params.ParamSet) {
	*(ps.(*types.Params)) = subspace.params
}
func (subspace *fakeParamSubspace) SetParamSet(_ sdk.Context, ps params.ParamSet) {
	subspace.params = *(ps.(*types.Params))
}
func (subspace *fakeParamSubspace) HasKeyTable() bool {
	// return true so the keeper does not try to call WithKeyTable, which does nothing
	return true
}
func (subspace *fakeParamSubspace) WithKeyTable(params.KeyTable) params.Subspace {
	// return an non-functional subspace to satisfy the interface
	return params.Subspace{}
}

// fakeSwapKeeper is a stub swap keeper.
// It can be used to return values to the incentive keeper without having to initialize a full swap keeper.
type fakeSwapKeeper struct {
	poolShares map[string]sdk.Int
}

var _ types.SwapKeeper = newFakeSwapKeeper()

func newFakeSwapKeeper() *fakeSwapKeeper {
	return &fakeSwapKeeper{
		poolShares: map[string]sdk.Int{},
	}
}
func (k *fakeSwapKeeper) addPool(id string, shares sdk.Int) *fakeSwapKeeper {
	k.poolShares[id] = shares
	return k
}
func (k *fakeSwapKeeper) GetPoolShares(_ sdk.Context, poolID string) (sdk.Int, bool) {
	shares, ok := k.poolShares[poolID]
	return shares, ok
}
func (k *fakeSwapKeeper) GetDepositorSharesAmount(_ sdk.Context, depositor sdk.AccAddress, poolID string) (sdk.Int, bool) {
	// This is just to implement the swap keeper interface.
	return sdk.Int{}, false
}

// fakeHardKeeper is a stub hard keeper.
// It can be used to return values to the incentive keeper without having to initialize a full hard keeper.
type fakeHardKeeper struct {
	borrows  fakeHardState
	deposits fakeHardState
}

type fakeHardState struct {
	total           sdk.Coins
	interestFactors map[string]sdk.Dec
}

func newFakeHardState() fakeHardState {
	return fakeHardState{
		total:           nil,
		interestFactors: map[string]sdk.Dec{}, // initialize map to avoid panics on read
	}
}

var _ types.HardKeeper = newFakeHardKeeper()

func newFakeHardKeeper() *fakeHardKeeper {
	return &fakeHardKeeper{
		borrows:  newFakeHardState(),
		deposits: newFakeHardState(),
	}
}

func (k *fakeHardKeeper) addTotalBorrow(coin sdk.Coin, factor sdk.Dec) *fakeHardKeeper {
	k.borrows.total = k.borrows.total.Add(coin)
	k.borrows.interestFactors[coin.Denom] = factor
	return k
}
func (k *fakeHardKeeper) addTotalSupply(coin sdk.Coin, factor sdk.Dec) *fakeHardKeeper {
	k.deposits.total = k.deposits.total.Add(coin)
	k.deposits.interestFactors[coin.Denom] = factor
	return k
}

func (k *fakeHardKeeper) GetBorrowedCoins(_ sdk.Context) (sdk.Coins, bool) {
	if k.borrows.total == nil {
		return nil, false
	}
	return k.borrows.total, true
}
func (k *fakeHardKeeper) GetSuppliedCoins(_ sdk.Context) (sdk.Coins, bool) {
	if k.deposits.total == nil {
		return nil, false
	}
	return k.deposits.total, true
}
func (k *fakeHardKeeper) GetBorrowInterestFactor(_ sdk.Context, denom string) (sdk.Dec, bool) {
	f, ok := k.borrows.interestFactors[denom]
	return f, ok
}
func (k *fakeHardKeeper) GetSupplyInterestFactor(_ sdk.Context, denom string) (sdk.Dec, bool) {
	f, ok := k.deposits.interestFactors[denom]
	return f, ok
}
func (k *fakeHardKeeper) GetBorrow(_ sdk.Context, _ sdk.AccAddress) (hardtypes.Borrow, bool) {
	panic("unimplemented")
}
func (k *fakeHardKeeper) GetDeposit(_ sdk.Context, _ sdk.AccAddress) (hardtypes.Deposit, bool) {
	panic("unimplemented")
}

// fakeStakingKeeper is a stub staking keeper.
// It can be used to return values to the incentive keeper without having to initialize a full staking keeper.
type fakeStakingKeeper struct {
	delegations stakingtypes.Delegations
	validators  stakingtypes.Validators
}

var _ types.StakingKeeper = newFakeStakingKeeper()

func newFakeStakingKeeper() *fakeStakingKeeper { return &fakeStakingKeeper{} }

func (k *fakeStakingKeeper) addBondedTokens(amount int64) *fakeStakingKeeper {
	if len(k.validators) != 0 {
		panic("cannot set total bonded if keeper already has validators set")
	}
	// add a validator with all the tokens
	k.validators = append(k.validators, stakingtypes.Validator{
		Status: sdk.Bonded,
		Tokens: sdk.NewInt(amount),
	})
	return k
}

func (k *fakeStakingKeeper) TotalBondedTokens(_ sdk.Context) sdk.Int {
	total := sdk.ZeroInt()
	for _, val := range k.validators {
		if val.GetStatus() == sdk.Bonded {
			total = total.Add(val.GetBondedTokens())
		}
	}
	return total
}
func (k *fakeStakingKeeper) GetDelegatorDelegations(_ sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) []stakingtypes.Delegation {
	return k.delegations
}
func (k *fakeStakingKeeper) GetValidator(_ sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	for _, val := range k.validators {
		if val.GetOperator().Equals(addr) {
			return val, true
		}
	}
	return stakingtypes.Validator{}, false
}
func (k *fakeStakingKeeper) GetValidatorDelegations(_ sdk.Context, valAddr sdk.ValAddress) []stakingtypes.Delegation {
	var delegations stakingtypes.Delegations
	for _, d := range k.delegations {
		if d.ValidatorAddress.Equals(valAddr) {
			delegations = append(delegations, d)
		}
	}
	return delegations
}

// fakeCDPKeeper is a stub cdp keeper.
// It can be used to return values to the incentive keeper without having to initialize a full cdp keeper.
type fakeCDPKeeper struct {
	interestFactor *sdk.Dec
	totalPrincipal sdk.Int
}

var _ types.CdpKeeper = newFakeCDPKeeper()

func newFakeCDPKeeper() *fakeCDPKeeper {
	return &fakeCDPKeeper{
		interestFactor: nil,
		totalPrincipal: sdk.ZeroInt(),
	}
}

func (k *fakeCDPKeeper) addInterestFactor(f sdk.Dec) *fakeCDPKeeper {
	k.interestFactor = &f
	return k
}
func (k *fakeCDPKeeper) addTotalPrincipal(p sdk.Int) *fakeCDPKeeper {
	k.totalPrincipal = p
	return k
}

func (k *fakeCDPKeeper) GetInterestFactor(_ sdk.Context, collateralType string) (sdk.Dec, bool) {
	if k.interestFactor != nil {
		return *k.interestFactor, true
	}
	return sdk.Dec{}, false
}
func (k *fakeCDPKeeper) GetTotalPrincipal(_ sdk.Context, collateralType string, principalDenom string) sdk.Int {
	return k.totalPrincipal
}
func (k *fakeCDPKeeper) GetCdpByOwnerAndCollateralType(_ sdk.Context, owner sdk.AccAddress, collateralType string) (cdptypes.CDP, bool) {
	return cdptypes.CDP{}, false
}
func (k *fakeCDPKeeper) GetCollateral(_ sdk.Context, collateralType string) (cdptypes.CollateralParam, bool) {
	return cdptypes.CollateralParam{}, false
}

// Assorted Testing Data

// note: amino panics when encoding times ≥ the start of year 10000.
var distantFuture = time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)

func arbitraryCoin() sdk.Coin {
	return c("hard", 1e9)
}

func arbitraryCoins() sdk.Coins {
	return cs(c("btcb", 1))
}

func arbitraryCoinsWithDenoms(denom ...string) sdk.Coins {
	const arbitraryAmount = 1 // must be > 0 as sdk.Coins type only stores positive amounts
	coins := sdk.NewCoins()
	for _, d := range denom {
		coins = coins.Add(sdk.NewInt64Coin(d, arbitraryAmount))
	}
	return coins
}

func arbitraryAddress() sdk.AccAddress {
	_, addresses := app.GeneratePrivKeyAddressPairs(1)
	return addresses[0]
}
func arbitraryValidatorAddress() sdk.ValAddress {
	return generateValidatorAddresses(1)[0]
}
func generateValidatorAddresses(n int) []sdk.ValAddress {
	_, addresses := app.GeneratePrivKeyAddressPairs(n)
	var valAddresses []sdk.ValAddress
	for _, a := range addresses {
		valAddresses = append(valAddresses, sdk.ValAddress(a))
	}
	return valAddresses
}

var nonEmptyMultiRewardIndexes = types.MultiRewardIndexes{
	{
		CollateralType: "bnb",
		RewardIndexes: types.RewardIndexes{
			{
				CollateralType: "hard",
				RewardFactor:   d("0.02"),
			},
			{
				CollateralType: "ukava",
				RewardFactor:   d("0.04"),
			},
		},
	},
	{
		CollateralType: "btcb",
		RewardIndexes: types.RewardIndexes{
			{
				CollateralType: "hard",
				RewardFactor:   d("0.2"),
			},
			{
				CollateralType: "ukava",
				RewardFactor:   d("0.4"),
			},
		},
	},
}

func extractCollateralTypes(indexes types.MultiRewardIndexes) []string {
	var denoms []string
	for _, ri := range indexes {
		denoms = append(denoms, ri.CollateralType)
	}
	return denoms
}

func increaseAllRewardFactors(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	increasedIndexes := make(types.MultiRewardIndexes, len(indexes))
	copy(increasedIndexes, indexes)

	for i := range increasedIndexes {
		increasedIndexes[i].RewardIndexes = increaseRewardFactors(increasedIndexes[i].RewardIndexes)
	}
	return increasedIndexes
}

func increaseRewardFactors(indexes types.RewardIndexes) types.RewardIndexes {
	increasedIndexes := make(types.RewardIndexes, len(indexes))
	copy(increasedIndexes, indexes)

	for i := range increasedIndexes {
		increasedIndexes[i].RewardFactor = increasedIndexes[i].RewardFactor.MulInt64(2)
	}
	return increasedIndexes
}

func appendUniqueMultiRewardIndex(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	const uniqueDenom = "uniquedenom"

	for _, mri := range indexes {
		if mri.CollateralType == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique multi reward index with denom '%s', but denom already existed", uniqueDenom))
		}
	}

	return append(indexes, types.NewMultiRewardIndex(
		uniqueDenom,
		types.RewardIndexes{
			{
				CollateralType: "hard",
				RewardFactor:   d("0.02"),
			},
			{
				CollateralType: "ukava",
				RewardFactor:   d("0.04"),
			},
		},
	),
	)
}

func appendUniqueEmptyMultiRewardIndex(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	const uniqueDenom = "uniquedenom"

	for _, mri := range indexes {
		if mri.CollateralType == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique multi reward index with denom '%s', but denom already existed", uniqueDenom))
		}
	}

	return append(indexes, types.NewMultiRewardIndex(uniqueDenom, nil))
}

func appendUniqueRewardIndexToFirstItem(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	newIndexes := make(types.MultiRewardIndexes, len(indexes))
	copy(newIndexes, indexes)

	newIndexes[0].RewardIndexes = appendUniqueRewardIndex(newIndexes[0].RewardIndexes)
	return newIndexes
}

func appendUniqueRewardIndex(indexes types.RewardIndexes) types.RewardIndexes {
	const uniqueDenom = "uniquereward"

	for _, mri := range indexes {
		if mri.CollateralType == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique reward index with denom '%s', but denom already existed", uniqueDenom))
		}
	}

	return append(
		indexes,
		types.NewRewardIndex(uniqueDenom, d("0.02")),
	)
}
