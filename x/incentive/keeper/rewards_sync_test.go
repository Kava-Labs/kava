package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// SynchronizeClaimTests runs unit tests for the keeper.SynchronizeClaim method
//
// inputs
// - claim in store (only claim.RewardIndexes, claim.Reward)
// - global indexes in store
// - shares function arg
//
// outputs
// - sets a claim
type SynchronizeClaimTests struct {
	unitTester
}

func TestSynchronizeClaim(t *testing.T) {
	suite.Run(t, new(SynchronizeClaimTests))
}

func (suite *SynchronizeClaimTests) TestClaimUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced
	// The user earns rewards for the time passed, and the claim indexes are updated

	originalReward := arbitraryCoins()
	collateralType := "base:quote"
	claimType := types.CLAIM_TYPE_USDX_MINTING

	claim := types.Claim{
		Type:   claimType,
		Owner:  arbitraryAddress(),
		Reward: originalReward,
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: collateralType,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.keeper.SetClaim(suite.ctx, claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: collateralType,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalIndexes(claimType, globalIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeClaim(suite.ctx, claimType, collateralType, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetClaim(suite.ctx, claimType, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * user shares
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeClaimTests) TestClaimUnchangedWhenGlobalIndexesUnchanged() {
	// It should be safe to call SynchronizeClaim multiple times

	collateralType := "base:quote"
	claimType := types.CLAIM_TYPE_USDX_MINTING

	unchangingIndexes := types.MultiRewardIndexes{
		{
			CollateralType: collateralType,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("1000.001"),
				},
			},
		},
	}

	claim := types.Claim{
		Type:          claimType,
		Owner:         arbitraryAddress(),
		Reward:        arbitraryCoins(),
		RewardIndexes: unchangingIndexes,
	}
	suite.keeper.SetClaim(suite.ctx, claim)

	suite.storeGlobalIndexes(claimType, unchangingIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeClaim(suite.ctx, claimType, collateralType, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetClaim(suite.ctx, claimType, claim.Owner)
	// claim should have the same rewards and indexes as before
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeClaimTests) TestClaimUpdatedWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a pool the user has already deposited to, and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the indexes are added to the claim.

	originalReward := arbitraryCoins()
	newlyRewardcollateralType := "newlyRewardedPool"
	claimType := types.CLAIM_TYPE_USDX_MINTING

	claim := types.Claim{
		Type:   claimType,
		Owner:  arbitraryAddress(),
		Reward: originalReward,
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
	suite.keeper.SetClaim(suite.ctx, claim)

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
			CollateralType: newlyRewardcollateralType,
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
	suite.storeGlobalIndexes(claimType, globalIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeClaim(suite.ctx, claimType, newlyRewardcollateralType, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetClaim(suite.ctx, claimType, claim.Owner)
	// the new indexes should be added to the claim, but the old ones should be unchanged
	newlyRewrdedIndexes, _ := globalIndexes.Get(newlyRewardcollateralType)
	expectedIndexes := claim.RewardIndexes.With(newlyRewardcollateralType, newlyRewrdedIndexes)
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares for the synced pool
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeClaimTests) TestClaimUnchangedWhenNoReward() {
	// When a pool is not rewarded but the user has deposited to that pool, and the claim is synced;
	// Then the claim should be the same.

	collateralType := "nonRewardPool"
	claimType := types.CLAIM_TYPE_USDX_MINTING

	claim := types.Claim{
		Type:          claimType,
		Owner:         arbitraryAddress(),
		Reward:        arbitraryCoins(),
		RewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.keeper.SetClaim(suite.ctx, claim)

	// No global indexes stored as this pool is not rewarded

	userShares := i(1e9)

	suite.keeper.SynchronizeClaim(suite.ctx, claimType, collateralType, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetClaim(suite.ctx, claimType, claim.Owner)
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeClaimTests) TestClaimUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded pool (that the user has already deposited to), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the new indexes are added.

	originalReward := arbitraryCoins()
	collateralType := "base:quote"
	claimType := types.CLAIM_TYPE_USDX_MINTING

	claim := types.Claim{
		Type:   claimType,
		Owner:  arbitraryAddress(),
		Reward: originalReward,
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: collateralType,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.keeper.SetClaim(suite.ctx, claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: collateralType,
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
	suite.storeGlobalIndexes(claimType, globalIndexes)

	userShares := i(1e9)

	suite.keeper.SynchronizeClaim(suite.ctx, claimType, collateralType, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetClaim(suite.ctx, claimType, claim.Owner)
	// indexes should have the new reward denom added
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeClaimTests) TestClaimUpdatedWhenGlobalIndexesIncreasedAndSourceIsZero() {
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced, but the user has no shares
	// The user earns no rewards for the time passed, but the claim indexes are updated

	collateralType := "base:quote"
	claimType := types.CLAIM_TYPE_USDX_MINTING

	claim := types.Claim{
		Type:   claimType,
		Owner:  arbitraryAddress(),
		Reward: arbitraryCoins(),
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: collateralType,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.keeper.SetClaim(suite.ctx, claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: collateralType,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalIndexes(claimType, globalIndexes)

	userShares := i(0)

	suite.keeper.SynchronizeClaim(suite.ctx, claimType, collateralType, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetClaim(suite.ctx, claimType, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// reward is unchanged
	suite.Equal(claim.Reward, syncedClaim.Reward)
}
