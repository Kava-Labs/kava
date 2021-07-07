package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// InitializeSwapRewardTests runs unit tests for the keeper.InitializeSwapReward method
//
// inputs
// - claim in store if it exists
// - global indexes in store
//
// outputs
// - sets or creates a claim
type InitializeSwapRewardTests struct {
	unitTester
}

func TestInitializeSwapReward(t *testing.T) {
	suite.Run(t, new(InitializeSwapRewardTests))
}

func (suite *InitializeSwapRewardTests) TestClaimAddedWhenClaimDoesNotExistAndNoRewards() {
	// When a claim doesn't exist, and a user deposits to a non-rewarded pool;
	// then a claim is added with no rewards and no indexes

	poolID := "base/quote"

	// no global indexes stored as this pool is not rewarded

	owner := arbitraryAddress()

	suite.keeper.InitializeSwapReward(suite.ctx, poolID, owner)

	syncedClaim, found := suite.keeper.GetSwapClaim(suite.ctx, owner)
	suite.True(found)
	// A new claim should have empty indexes. It doesn't strictly need the poolID either.
	expectedIndexes := types.MultiRewardIndexes{{
		CollateralType: poolID,
		RewardIndexes:  nil,
	}}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeSwapRewardTests) TestClaimAddedWhenClaimDoesNotExistAndRewardsExist() {
	// When a claim doesn't exist, and a user deposits to a rewarded pool;
	// then a claim is added with no rewards and indexes matching the global indexes

	poolID := "base/quote"

	globalIndexes := types.MultiRewardIndexes{
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
	suite.storeGlobalSwapIndexes(globalIndexes)

	owner := arbitraryAddress()

	suite.keeper.InitializeSwapReward(suite.ctx, poolID, owner)

	syncedClaim, found := suite.keeper.GetSwapClaim(suite.ctx, owner)
	suite.True(found)
	// a new claim should start with the current global indexes
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeSwapRewardTests) TestClaimUpdatedWhenClaimExistsAndNoRewards() {
	// When a claim exists, and a user deposits to a new non-rewarded pool;
	// then the claim's rewards don't change

	preexistingPoolID := "preexisting"
	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	newPoolID := "btcb/usdx"

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingPoolID,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.storeSwapClaim(claim)

	// no global indexes stored as the new pool is not rewarded

	suite.keeper.InitializeSwapReward(suite.ctx, newPoolID, claim.Owner)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	// The preexisting indexes shouldn't be changed. It doesn't strictly need the new poolID either.
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingPoolID,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newPoolID,
			RewardIndexes:  nil,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *InitializeSwapRewardTests) TestClaimUpdatedWhenClaimExistsAndRewardsExist() {
	// When a claim exists, and a user deposits to a new rewarded pool;
	// then the claim's rewards don't change and the indexes are updated to match the global indexes

	preexistingPoolID := "preexisting"
	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	newPoolID := "btcb/usdx"
	newIndexes := types.RewardIndexes{
		{
			CollateralType: "otherrewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	claim := types.SwapClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingPoolID,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.storeSwapClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingPoolID,
			RewardIndexes:  increaseRewardFactors(preexistingIndexes),
		},
		{
			CollateralType: newPoolID,
			RewardIndexes:  newIndexes,
		},
	}
	suite.storeGlobalSwapIndexes(globalIndexes)

	suite.keeper.InitializeSwapReward(suite.ctx, newPoolID, claim.Owner)

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	// only the indexes for the new pool should be updated
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingPoolID,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newPoolID,
			RewardIndexes:  newIndexes,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}
