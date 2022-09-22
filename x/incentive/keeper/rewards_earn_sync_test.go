package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// SynchronizeEarnRewardTests runs unit tests for the keeper.SynchronizeEarnReward method
//
// inputs
// - claim in store (only claim.RewardIndexes, claim.Reward)
// - global indexes in store
// - shares function arg
//
// outputs
// - sets a claim
type SynchronizeEarnRewardTests struct {
	unitTester
}

func TestSynchronizeEarnReward(t *testing.T) {
	suite.Run(t, new(SynchronizeEarnRewardTests))
}

func (suite *SynchronizeEarnRewardTests) TestClaimUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced
	// The user earns rewards for the time passed, and the claim indexes are updated

	originalReward := arbitraryCoins()
	vaultDenom := "cats"

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: vaultDenom,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeEarnClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(globalIndexes)

	userShares := d("1000000000")

	suite.keeper.SynchronizeEarnReward(suite.ctx, vaultDenom, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * user shares
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeEarnRewardTests) TestClaimUnchangedWhenGlobalIndexesUnchanged() {
	// It should be safe to call SynchronizeEarnReward multiple times

	vaultDenom := "cats"
	unchangingIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("1000.001"),
				},
			},
		},
	}

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: unchangingIndexes,
	}
	suite.storeEarnClaim(claim)

	suite.storeGlobalEarnIndexes(unchangingIndexes)

	userShares := d("1000000000")

	suite.keeper.SynchronizeEarnReward(suite.ctx, vaultDenom, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	// claim should have the same rewards and indexes as before
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeEarnRewardTests) TestClaimUpdatedWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a vault the user has already deposited to, and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the indexes are added to the claim.

	originalReward := arbitraryCoins()
	newlyRewardVaultDenom := "newlyRewardedVault"

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "currentlyRewardedVault",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeEarnClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "currentlyRewardedVault",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: newlyRewardVaultDenom,
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
	suite.storeGlobalEarnIndexes(globalIndexes)

	userShares := d("1000000000")

	suite.keeper.SynchronizeEarnReward(suite.ctx, newlyRewardVaultDenom, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	// the new indexes should be added to the claim, but the old ones should be unchanged
	newlyRewrdedIndexes, _ := globalIndexes.Get(newlyRewardVaultDenom)
	expectedIndexes := claim.RewardIndexes.With(newlyRewardVaultDenom, newlyRewrdedIndexes)
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares for the synced vault
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeEarnRewardTests) TestClaimUnchangedWhenNoReward() {
	// When a vault is not rewarded but the user has deposited to that vault, and the claim is synced;
	// Then the claim should be the same.

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeEarnClaim(claim)

	vaultDenom := "nonRewardVault"
	// No global indexes stored as this vault is not rewarded

	userShares := d("1000000000")

	suite.keeper.SynchronizeEarnReward(suite.ctx, vaultDenom, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeEarnRewardTests) TestClaimUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded vault (that the user has already deposited to), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the new indexes are added.

	originalReward := arbitraryCoins()
	vaultDenom := "cats"

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: vaultDenom,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeEarnClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
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
	suite.storeGlobalEarnIndexes(globalIndexes)

	userShares := d("1000000000")

	suite.keeper.SynchronizeEarnReward(suite.ctx, vaultDenom, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	// indexes should have the new reward denom added
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeEarnRewardTests) TestClaimUpdatedWhenGlobalIndexesIncreasedAndSourceIsZero() {
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced, but the user has no shares
	// The user earns no rewards for the time passed, but the claim indexes are updated

	vaultDenom := "cats"

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: vaultDenom,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeEarnClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(globalIndexes)

	userShares := d("0")

	suite.keeper.SynchronizeEarnReward(suite.ctx, vaultDenom, claim.Owner, userShares)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// reward is unchanged
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *SynchronizeEarnRewardTests) TestGetSyncedClaim_ClaimUnchangedWhenNoGlobalIndexes() {
	vaultDenom_1 := "usdx"
	owner := arbitraryAddress()

	earnKeeper := newFakeEarnKeeper().
		addDeposit(owner, earntypes.NewVaultShare("usdx", d("1000000000")))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  owner,
			Reward: nil,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: vaultDenom_1,
				RewardIndexes:  nil, // this state only happens because Init stores empty indexes
			},
		},
	}
	suite.storeEarnClaim(claim)

	// no global indexes for any vault

	syncedClaim, f := suite.keeper.GetSynchronizedEarnClaim(suite.ctx, claim.Owner)
	suite.True(f)

	// indexes are unchanged
	suite.Equal(claim.RewardIndexes, syncedClaim.RewardIndexes)
	// reward is unchanged
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *SynchronizeEarnRewardTests) TestGetSyncedClaim_ClaimUpdatedWhenMissingIndexAndHasNoSourceShares() {
	vaultDenom_1 := "usdx"
	vaultDenom_2 := "ukava"
	owner := arbitraryAddress()

	// owner has no shares in any vault
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, newFakeEarnKeeper())

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  owner,
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: vaultDenom_1,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom1",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeEarnClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom_1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom1",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: vaultDenom_2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom2",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(globalIndexes)

	syncedClaim, f := suite.keeper.GetSynchronizedEarnClaim(suite.ctx, claim.Owner)
	suite.True(f)

	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// reward is unchanged
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *SynchronizeEarnRewardTests) TestGetSyncedClaim_ClaimUpdatedWhenMissingIndexButHasSourceShares() {
	VaultDenom_1 := "usdx"
	VaultDenom_2 := "ukava"
	owner := arbitraryAddress()

	earnKeeper := newFakeEarnKeeper().
		addVault(VaultDenom_1, earntypes.NewVaultShare(VaultDenom_1, d("1000000000"))).
		addVault(VaultDenom_2, earntypes.NewVaultShare(VaultDenom_2, d("1000000000"))).
		addDeposit(owner, earntypes.NewVaultShare(VaultDenom_1, d("1000000000"))).
		addDeposit(owner, earntypes.NewVaultShare(VaultDenom_2, d("1000000000")))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  owner,
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: VaultDenom_1,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom1",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeEarnClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: VaultDenom_1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom1",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: VaultDenom_2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom2",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(globalIndexes)

	syncedClaim, f := suite.keeper.GetSynchronizedEarnClaim(suite.ctx, claim.Owner)
	suite.True(f)

	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// reward is incremented
	expectedReward := cs(c("rewarddenom1", 1_000_001_000_000), c("rewarddenom2", 2_000_002_000_000))
	suite.Equal(claim.Reward.Add(expectedReward...), syncedClaim.Reward)
}
