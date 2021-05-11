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
	// SynchronizeHardBorrowReward does not use param subspace. The store key needs to be initialized in the multistore for the subspace to function.
	paramSubspace := params.NewKeeper(
		cdc,
		sdk.NewKVStoreKey(params.StoreKey),
		sdk.NewTransientStoreKey(params.StoreKey),
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

func (suite *SynchronizeHardBorrowRewardTests) TestClaimIndexesAreUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := increaseAllRewardFactors(nonEmptyMultiRewardIndexes)
	suite.storeGlobalIndexes(globalIndexes)
	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(claim.BorrowRewardIndexes)...),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *SynchronizeHardBorrowRewardTests) TestClaimIndexesAreUnchangedWhenGlobalIndexesUnchanged() {
	// It should be safe to call SynchronizeHardBorrowReward multiple times

	unchangingIndexes := nonEmptyMultiRewardIndexes

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: unchangingIndexes,
	}
	suite.storeClaim(claim)

	suite.storeGlobalIndexes(unchangingIndexes)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(unchangingIndexes)...),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(unchangingIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *SynchronizeHardBorrowRewardTests) TestClaimIndexesAreUpdatedWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a hard borrow denom the user has already borrowed, and the claim is synced;
	// Then the new reward's index should be added to the claim.
	suite.T().Skip("TODO fix this bug")

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := appendUniqueMultiRewardIndex(nonEmptyMultiRewardIndexes)
	suite.storeGlobalIndexes(globalIndexes)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}
func (suite *SynchronizeHardBorrowRewardTests) TestClaimIndexesAreUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded borrow denom (that the user has already borrowed), and the claim is synced;
	// Then the new reward coin's index should be added to the claim.

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		BorrowRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := appendUniqueRewardIndexToFirstItem(nonEmptyMultiRewardIndexes)
	suite.storeGlobalIndexes(globalIndexes)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.BorrowRewardIndexes)
}

func (suite *SynchronizeHardBorrowRewardTests) TestRewardIsIncrementedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced
	// The user earns rewards for the time passed

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

func (suite *SynchronizeHardBorrowRewardTests) TestRewardIsIncrementedWhenWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a hard borrow denom the user has already borrowed, and the claim is synced
	// Then the user earns rewards for the time since the reward was added
	suite.T().Skip("TODO fix this bug")

	originalReward := arbitraryCoins()
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		BorrowRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "rewarded",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "rewarded",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: "newlyrewarded",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "otherreward",
					// Indexes start at 0 when the reward is added by gov,
					// so this represents the syncing happening some time later.
					RewardFactor: d("1000.001"),
				},
			},
		},
	}
	suite.storeGlobalIndexes(globalIndexes)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   cs(c("rewarded", 1e9), c("newlyrewarded", 1e9)),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	// new reward is (new index - old index) * borrow amount for each borrowed denom
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("otherreward", 1_000_001_000_000), c("reward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}
func (suite *SynchronizeHardBorrowRewardTests) TestRewardIsIncrementedWhenWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded borrow denom (that the user has already borrowed), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added

	originalReward := arbitraryCoins()
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		BorrowRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "borrowed",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "borrowed",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
				{
					CollateralType: "otherreward",
					// Indexes start at 0 when the reward is added by gov,
					// so this represents the syncing happening some time later.
					RewardFactor: d("1000.001"),
				},
			},
		},
	}
	suite.storeGlobalIndexes(globalIndexes)

	borrow := hardtypes.Borrow{
		Borrower: claim.Owner,
		Amount:   cs(c("borrowed", 1e9)),
	}

	suite.keeper.SynchronizeHardBorrowReward(suite.ctx, borrow)

	// new reward is (new index - old index) * borrow amount for each borrowed denom
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
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
		coins = coins.Add(sdk.NewInt64Coin(d, arbitraryAmount))
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
