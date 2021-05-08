package keeper_test

import (
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

// SynchronizeHardBorrowRewardTests runs unit tests for the keeper.SynchronizeHardBorrowReward method
//
// inputs
// - claim in store (only claim.BorrowRewardIndexes, claim.Reward)
// - global indexes in store
// - borrow function arg (only borrow.Amount)
//
// outputs
// - sets a claim
type SynchronizeHardBorrowRewardTests struct {
	suite.Suite
	keeper keeper.Keeper
	ctx    sdk.Context
}

func TestSynchronizeHardBorrowReward(t *testing.T) {
	suite.Run(t, new(SynchronizeHardBorrowRewardTests))
}

func (suite *SynchronizeHardBorrowRewardTests) SetupTest() {
	incentveStoreKey := sdk.NewKVStoreKey(types.StoreKey)
	suite.keeper = suite.setupKeeper(incentveStoreKey)
	suite.ctx = NewTestContext(incentveStoreKey)
}

func (suite *SynchronizeHardBorrowRewardTests) TearDownTest() {
	suite.keeper = keeper.Keeper{}
	suite.ctx = sdk.Context{}
}

func (suite *SynchronizeHardBorrowRewardTests) setupKeeper(incentiveStoreKey sdk.StoreKey) keeper.Keeper {
	cdc := app.MakeCodec()
	paramSubspace := params.NewKeeper(
		cdc,
		sdk.NewKVStoreKey(params.StoreKey),
		sdk.NewTransientStoreKey(params.StoreKey), // TODO param subspace isn't used in the tests, replace with interface and use nil instead
	).Subspace(incentive.DefaultParamspace)

	return keeper.NewKeeper(cdc, incentiveStoreKey, paramSubspace, nil, nil, nil, nil, nil)
}

func (suite *SynchronizeHardBorrowRewardTests) storeGlobalIndexes(indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (suite *SynchronizeHardBorrowRewardTests) storeClaim(claim types.HardLiquidityProviderClaim) {
	suite.keeper.SetHardLiquidityProviderClaim(suite.ctx, claim)
}

func (suite *SynchronizeHardBorrowRewardTests) TestClaimIndexesAreUpdatedWhenGlobalIndexesAmountsIncreased() {

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	suite.storeGlobalIndexes(
		increaseAllRewardFactors(nonEmptyMultiRewardIndexes),
	)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(claim.BorrowRewardIndexes)...),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(claim.BorrowRewardIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *SynchronizeHardBorrowRewardTests) TestRewardIsIncrementedWhenGlobalIndexesHaveIncreased() {
	originalReward := arbitraryCoins()

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		BorrowRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "borrowdenom",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeClaim(claim)

	suite.storeGlobalIndexes(types.MultiRewardIndexes{
		{
			CollateralType: "borrowdenom",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	})

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   cs(c("borrowdenom", 1e9)),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	// new reward is (new index - old index) * borrow amount
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func arbitraryCoins() sdk.Coins {
	return cs(c("btcb", 1))
}

func arbitraryCoinsWithDenoms(denom ...string) sdk.Coins {
	const arbitraryAmount = 1 // must be > 0 as sdk.Coins type only stores positive amounts
	coins := sdk.NewCoins()
	for _, d := range denom {
		coins.Add(sdk.NewInt64Coin(d, arbitraryAmount))
	}
	return coins
}

func arbitraryAddress() sdk.AccAddress {
	_, addresses := app.GeneratePrivKeyAddressPairs(1)
	return addresses[0]
}

var nonEmptyMultiRewardIndexes = types.MultiRewardIndexes{
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
