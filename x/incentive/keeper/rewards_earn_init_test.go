package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// InitializeEarnRewardTests runs unit tests for the keeper.InitializeEarnReward method
//
// inputs
// - claim in store if it exists
// - global indexes in store
//
// outputs
// - sets or creates a claim
type InitializeEarnRewardTests struct {
	unitTester
}

func TestInitializeEarnReward(t *testing.T) {
	suite.Run(t, new(InitializeEarnRewardTests))
}

func (suite *InitializeEarnRewardTests) TestClaimAddedWhenClaimDoesNotExistAndNoRewards() {
	// When a claim doesn't exist, and a user deposits to a non-rewarded pool;
	// then a claim is added with no rewards and no indexes

	vaultDenom := "usdx"

	// no global indexes stored as this pool is not rewarded

	owner := arbitraryAddress()

	suite.keeper.InitializeEarnReward(suite.ctx, vaultDenom, owner)

	syncedClaim, found := suite.keeper.GetEarnClaim(suite.ctx, owner)
	suite.True(found)
	// A new claim should have empty indexes. It doesn't strictly need the vaultDenom either.
	expectedIndexes := types.MultiRewardIndexes{{
		CollateralType: vaultDenom,
		RewardIndexes:  nil,
	}}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeEarnRewardTests) TestClaimAddedWhenClaimDoesNotExistAndRewardsExist() {
	// When a claim doesn't exist, and a user deposits to a rewarded pool;
	// then a claim is added with no rewards and indexes matching the global indexes

	vaultDenom := "usdx"

	globalIndexes := types.MultiRewardIndexes{
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
	suite.storeGlobalEarnIndexes(globalIndexes)

	owner := arbitraryAddress()

	suite.keeper.InitializeEarnReward(suite.ctx, vaultDenom, owner)

	syncedClaim, found := suite.keeper.GetEarnClaim(suite.ctx, owner)
	suite.True(found)
	// a new claim should start with the current global indexes
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeEarnRewardTests) TestClaimUpdatedWhenClaimExistsAndNoRewards() {
	// When a claim exists, and a user deposits to a new non-rewarded pool;
	// then the claim's rewards don't change

	preexistingvaultDenom := "preexisting"
	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	newVaultDenom := "btcb:usdx"

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingvaultDenom,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.storeEarnClaim(claim)

	// no global indexes stored as the new pool is not rewarded

	suite.keeper.InitializeEarnReward(suite.ctx, newVaultDenom, claim.Owner)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	// The preexisting indexes shouldn't be changed. It doesn't strictly need the new vaultDenom either.
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingvaultDenom,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newVaultDenom,
			RewardIndexes:  nil,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *InitializeEarnRewardTests) TestClaimUpdatedWhenClaimExistsAndRewardsExist() {
	// When a claim exists, and a user deposits to a new rewarded pool;
	// then the claim's rewards don't change and the indexes are updated to match the global indexes

	preexistingvaultDenom := "preexisting"
	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	newVaultDenom := "btcb:usdx"
	newIndexes := types.RewardIndexes{
		{
			CollateralType: "otherrewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	claim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingvaultDenom,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.storeEarnClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingvaultDenom,
			RewardIndexes:  increaseRewardFactors(preexistingIndexes),
		},
		{
			CollateralType: newVaultDenom,
			RewardIndexes:  newIndexes,
		},
	}
	suite.storeGlobalEarnIndexes(globalIndexes)

	suite.keeper.InitializeEarnReward(suite.ctx, newVaultDenom, claim.Owner)

	syncedClaim, _ := suite.keeper.GetEarnClaim(suite.ctx, claim.Owner)
	// only the indexes for the new pool should be updated
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingvaultDenom,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newVaultDenom,
			RewardIndexes:  newIndexes,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}
