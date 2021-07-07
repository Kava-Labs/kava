package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// SynchronizeSwapRewardTests runs unit tests for the keeper.SynchronizeSwapReward method
//
// inputs
// - claim in store (only claim.RewardIndexes, claim.Reward)
// - global indexes in store
// - shares function arg
//
// outputs
// - sets a claim
type SynchronizeSwapRewardTests struct {
	unitTester
}

func TestSynchronizeSwapReward(t *testing.T) {
	suite.Run(t, new(SynchronizeSwapRewardTests))
}

// TODO reuse across swap tests
func (suite *SynchronizeSwapRewardTests) storeClaim(claim types.SwapClaim) {
	suite.keeper.SetSwapClaim(suite.ctx, claim)
}

func (suite *SynchronizeSwapRewardTests) TestClaimUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced
	// The user earns rewards for the time passed, and the claim indexes are updated

	originalReward := arbitraryCoins()
	poolID := "base/quote"

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: poolID,
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

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: poolID,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalSwapIndexes(globalIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeSwapReward(suite.ctx, poolID, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * user shares
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeSwapRewardTests) TestClaimUnchangedWhenGlobalIndexesUnchanged() {
	// It should be safe to call SynchronizeSwapReward multiple times

	poolID := "base/quote"
	unchangingIndexes := types.MultiRewardIndexes{
		{
			CollateralType: poolID,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("1000.001"),
				},
			},
		},
	}

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: unchangingIndexes,
	}
	suite.storeClaim(claim)

	suite.storeGlobalSwapIndexes(unchangingIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeSwapReward(suite.ctx, poolID, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	// claim should have the same rewards and indexes as before
	suite.Equal(claim, syncedClaim)
}
func (suite *SynchronizeSwapRewardTests) TestClaimUpdatedWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a pool the user has already deposited to, and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the indexes are added to the claim.

	originalReward := arbitraryCoins()
	newlyRewardPoolID := "newlyRewardedPool"

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "currentlyRewardedPool",
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
			CollateralType: "currentlyRewardedPool",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: newlyRewardPoolID,
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
	suite.storeGlobalSwapIndexes(globalIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeSwapReward(suite.ctx, newlyRewardPoolID, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	// the new indexes should be added to the claim, but the old ones should be unchanged
	newlyRewrdedIndexes, _ := globalIndexes.Get(newlyRewardPoolID)
	expectedIndexes := claim.RewardIndexes.With(newlyRewardPoolID, newlyRewrdedIndexes)
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares for the synced pool
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeSwapRewardTests) TestClaimUnchangedWhenNoReward() {
	// When a pool is not rewarded but the user has deposited to that pool, and the claim is synced;
	// Then the claim should be the same.

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeClaim(claim)

	poolID := "nonRewardPool"
	// No global indexes stored as this pool is not rewarded

	userShares := i(1e9)

	suite.keeper.SynchronizeSwapReward(suite.ctx, poolID, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeSwapRewardTests) TestClaimUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded pool (that the user has already deposited to), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the new indexes are added.

	originalReward := arbitraryCoins()
	poolID := "base/quote"

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: poolID,
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
			CollateralType: poolID,
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
	suite.storeGlobalSwapIndexes(globalIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeSwapReward(suite.ctx, poolID, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	// indexes should have the new reward denom added
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeSwapRewardTests) TestClaimUpdatedWhenGlobalIndexesIncreasedAndSourceIsZero() {
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced, but the user has no shares
	// The user earns no rewards for the time passed, but the claim indexes are updated

	poolID := "base/quote"

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: poolID,
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

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: poolID,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalSwapIndexes(globalIndexes)

	userShares := i(0)

	suite.keeper.SynchronizeSwapReward(suite.ctx, poolID, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// reward is unchanged
	suite.Equal(claim.Reward, syncedClaim.Reward)
}
