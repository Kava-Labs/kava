package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive"
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

type unitTester struct {
	suite.Suite
	keeper keeper.Keeper
	ctx    sdk.Context
}

func (suite *unitTester) SetupTest() {
	incentiveStoreKey := sdk.NewKVStoreKey(types.StoreKey)
	suite.keeper = suite.setupKeeper(incentiveStoreKey)
	suite.ctx = NewTestContext(incentiveStoreKey)
}

func (suite *unitTester) TearDownTest() {
	suite.keeper = keeper.Keeper{}
	suite.ctx = sdk.Context{}
}

func (suite *unitTester) setupKeeper(incentiveStoreKey sdk.StoreKey) keeper.Keeper {
	cdc := app.MakeCodec()
	// TODO The param store key needs to be initialized in the multistore for the subspace to function.
	paramSubspace := params.NewKeeper(
		cdc,
		sdk.NewKVStoreKey(params.StoreKey),
		sdk.NewTransientStoreKey(params.StoreKey),
	).Subspace(incentive.DefaultParamspace)

	return keeper.NewKeeper(cdc, incentiveStoreKey, paramSubspace, nil, nil, nil, nil, nil)
}

func (suite *unitTester) storeGlobalIndexes(indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (suite *unitTester) storeClaim(claim types.HardLiquidityProviderClaim) {
	suite.keeper.SetHardLiquidityProviderClaim(suite.ctx, claim)
}

// InitializeHardBorrowRewardTests runs unit tests for the keeper.InitializeHardBorrowReward method
//
// inputs
// - claim in store if it exists (only claim.BorrowRewardIndexes)
// - global indexes in store
// - borrow function arg (only borrow.Amount)
//
// outputs
// - sets or creates a claim
type InitializeHardBorrowRewardTests struct {
	unitTester
}

func TestInitializeHardBorrowReward(t *testing.T) {
	suite.Run(t, new(InitializeHardBorrowRewardTests))
}

func (suite *InitializeHardBorrowRewardTests) TestClaimIndexesAreSetWhenClaimExists() {
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		// Indexes should always be empty when initialize is called.
		// If initialize is called then the user must have repaid their borrow positions,
		// which means UpdateHardBorrowIndexDenoms was called and should have remove indexes.
		BorrowRewardIndexes: types.MultiRewardIndexes{},
	}
	suite.storeClaim(claim)

	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalIndexes(globalIndexes)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

	suite.keeper.InitializeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *InitializeHardBorrowRewardTests) TestClaimIndexesAreSetWhenClaimDoesNotExist() {
	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalIndexes(globalIndexes)

	owner := arbitraryAddress()
	borrow := hardtypes.Borrow{
		Borrower: owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

	suite.keeper.InitializeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, found := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, owner)
	suite.True(found)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *InitializeHardBorrowRewardTests) TestClaimIndexesAreSetEmptyForMissingIndexes() {

	globalIndexes := nonEmptyMultiRewardIndexes
	suite.storeGlobalIndexes(globalIndexes)

	owner := arbitraryAddress()
	// Borrow a denom that is not in the global indexes.
	// This happens when a borrow denom has no rewards associated with it.
	expectedIndexes := appendUniqueEmptyMultiRewardIndex(globalIndexes)
	borrowedDenoms := extractCollateralTypes(expectedIndexes)
	borrow := hardtypes.Borrow{
		Borrower: owner,
		Amount:   arbitraryCoinsWithDenoms(borrowedDenoms...),
	}

	suite.keeper.InitializeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, owner)
	suite.Equal(expectedIndexes, syncedClaim.BorrowRewardIndexes)
}

func appendUniqueDenom(denoms []string) []string {
	uniqueDenom := "uniquedenom"
	for _, d := range denoms {
		if d == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique denom '%s', but denom already existed", uniqueDenom))
		}
	}
	return append(denoms, uniqueDenom)
}
