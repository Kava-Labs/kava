package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// SynchronizeHardSupplyRewardTests runs unit tests for the keeper.SynchronizeHardSupplyReward method
//
// inputs
// - claim in store (only claim.SupplyRewardIndexes, claim.Reward)
// - global indexes in store
// - deposit function arg (only deposit.Amount)
//
// outputs
// - sets a claim
type SynchronizeHardSupplyRewardTests struct {
	unitTester
}

func TestSynchronizeHardSupplyReward(t *testing.T) {
	suite.Run(t, new(SynchronizeHardSupplyRewardTests))
}

func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := increaseAllRewardFactors(nonEmptyMultiRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)
	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(extractCollateralTypes(claim.SupplyRewardIndexes)...),
	}

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}
func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUnchangedWhenGlobalIndexesUnchanged() {
	// It should be safe to call SynchronizeHardSupplyReward multiple times

	unchangingIndexes := nonEmptyMultiRewardIndexes

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: unchangingIndexes,
	}
	suite.storeClaim(claim)

	suite.storeGlobalSupplyIndexes(unchangingIndexes)

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(extractCollateralTypes(unchangingIndexes)...),
	}

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(unchangingIndexes, syncedClaim.SupplyRewardIndexes)
}
func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUpdatedWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a hard deposit denom the user has already deposited, and the claim is synced;
	// Then the new reward's index should be added to the claim.
	suite.T().Skip("TODO fix this bug")

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := appendUniqueMultiRewardIndex(nonEmptyMultiRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}
func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded deposit denom (that the user has already deposited), and the claim is synced;
	// Then the new reward coin's index should be added to the claim.

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := appendUniqueRewardIndexToFirstItem(nonEmptyMultiRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    arbitraryCoinsWithDenoms(extractCollateralTypes(globalIndexes)...),
	}

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}

func (suite *SynchronizeHardSupplyRewardTests) TestRewardIsIncrementedWhenGlobalIndexesHaveIncreased() {
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
		SupplyRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "depositdenom",
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

	suite.storeGlobalSupplyIndexes(types.MultiRewardIndexes{
		{
			CollateralType: "depositdenom",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	})

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    cs(c("depositdenom", 1e9)),
	}

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	// new reward is (new index - old index) * deposit amount
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeHardSupplyRewardTests) TestRewardIsIncrementedWhenWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a hard deposit denom the user has already deposited, and the claim is synced
	// Then the user earns rewards for the time since the reward was added
	suite.T().Skip("TODO fix this bug")

	originalReward := arbitraryCoins()
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		SupplyRewardIndexes: types.MultiRewardIndexes{
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
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    cs(c("rewarded", 1e9), c("newlyrewarded", 1e9)),
	}

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	// new reward is (new index - old index) * deposit amount for each deposited denom
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("otherreward", 1_000_001_000_000), c("reward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}
func (suite *SynchronizeHardSupplyRewardTests) TestRewardIsIncrementedWhenWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded deposit denom (that the user has already deposited), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added

	originalReward := arbitraryCoins()
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		SupplyRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "deposited",
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
			CollateralType: "deposited",
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
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := hardtypes.Deposit{
		Depositor: claim.Owner,
		Amount:    cs(c("deposited", 1e9)),
	}

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	// new reward is (new index - old index) * deposit amount for each deposited denom
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}
