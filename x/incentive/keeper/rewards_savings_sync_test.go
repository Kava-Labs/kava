package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// SynchronizeSavingsRewardTests runs unit tests for the keeper.SynchronizeSavingsReward method
type SynchronizeSavingsRewardTests struct {
	unitTester
}

func TestSynchronizeSavingsReward(t *testing.T) {
	suite.Run(t, new(SynchronizeSavingsRewardTests))
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced
	// The user earns rewards for the time passed, and the claim indexes are updated

	originalReward := arbitraryCoins()
	denom := "test"

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: denom,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeSavingsClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalSavingsIndexes(globalIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(denom, userShares)))
	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, []string{})

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * user shares
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUnchangedWhenGlobalIndexesUnchanged() {
	denom := "test"
	unchangingIndexes := types.MultiRewardIndexes{
		{
			CollateralType: denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("1000.001"),
				},
			},
		},
	}

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: unchangingIndexes,
	}
	suite.storeSavingsClaim(claim)

	suite.storeGlobalSavingsIndexes(unchangingIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(denom, userShares)))
	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, []string{})

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// claim should have the same rewards and indexes as before
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUpdatedWhenNewRewardAdded() {
	originalReward := arbitraryCoins()
	newlyRewardedDenom := "newlyRewardedDenom"

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "currentlyRewardedDenom",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeSavingsClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "currentlyRewardedDenom",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: newlyRewardedDenom,
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
	suite.storeGlobalSavingsIndexes(globalIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner,
		sdk.NewCoins(
			sdk.NewCoin("currentlyRewardedDenom", userShares),
			sdk.NewCoin(newlyRewardedDenom, userShares),
		),
	)

	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, []string{newlyRewardedDenom})

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// the new indexes should be added to the claim and the old ones should be updated
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares for the synced deposit
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("reward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded denom (that the user has already deposited to), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the new indexes are added.

	originalReward := arbitraryCoins()
	denom := "base"

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: denom,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeSavingsClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: denom,
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
	suite.storeGlobalSavingsIndexes(globalIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(denom, userShares)))
	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, []string{})

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// indexes should have the new reward denom added
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func getDenoms(coins sdk.Coins) []string {
	denoms := []string{}
	for _, coin := range coins {
		denoms = append(denoms, coin.Denom)
	}
	return denoms
}
