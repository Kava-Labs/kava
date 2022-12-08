package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// InitializeClaimTests runs unit tests for the keeper.InitializeClaim method
//
// inputs
// - claim in store if it exists
// - global indexes in store
//
// outputs
// - sets or creates a claim
type InitializeClaimTests struct {
	unitTester
}

func TestInitializeRewardTests(t *testing.T) {
	suite.Run(t, new(InitializeClaimTests))
}

func (suite *InitializeClaimTests) TestClaimAddedWhenClaimDoesNotExistAndNoRewards() {
	// When a claim doesn't exist, and a user deposits to a non-rewarded pool;
	// then a claim is added with no rewards and no indexes

	collateralType := "usdc"
	claimType := types.CLAIM_TYPE_SWAP
	owner := arbitraryAddress()

	// no global indexes stored as this pool is not rewarded

	suite.keeper.InitializeClaimSingleReward(suite.ctx, claimType, owner, collateralType)

	syncedClaim, found := suite.keeper.Store.GetClaim(suite.ctx, claimType, owner)
	suite.True(found)
	// A new claim should have empty indexes. It doesn't strictly need the collateralType either.
	expectedIndexes := types.MultiRewardIndexes{{
		CollateralType: collateralType,
		RewardIndexes:  nil,
	}}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeClaimTests) TestClaimAddedWhenClaimDoesNotExistAndRewardsExist() {
	// When a claim doesn't exist, and a user deposits to a rewarded pool;
	// then a claim is added with no rewards and indexes matching the global indexes

	collateralType := "usdc"
	claimType := types.CLAIM_TYPE_SWAP
	owner := arbitraryAddress()

	globalIndexes := types.MultiRewardIndexes{
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
	suite.storeGlobalIndexes(claimType, globalIndexes)

	suite.keeper.InitializeClaimSingleReward(suite.ctx, claimType, owner, collateralType)

	syncedClaim, found := suite.keeper.Store.GetClaim(suite.ctx, claimType, owner)
	suite.True(found)
	// a new claim should start with the current global indexes
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeClaimTests) TestClaimUpdatedWhenClaimExistsAndNoRewards() {
	// When a claim exists, and a user deposits to a new non-rewarded pool;
	// then the claim's rewards don't change

	preexistingCollateralType := "preexisting"

	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	newCollateralType := "usdc"
	claimType := types.CLAIM_TYPE_SWAP

	claim := types.Claim{
		Type:   claimType,
		Owner:  arbitraryAddress(),
		Reward: arbitraryCoins(),
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingCollateralType,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.keeper.Store.SetClaim(suite.ctx, claim)

	// no global indexes stored as the new pool is not rewarded

	suite.keeper.InitializeClaimSingleReward(suite.ctx, claimType, claim.Owner, newCollateralType)

	syncedClaim, _ := suite.keeper.Store.GetClaim(suite.ctx, claimType, claim.Owner)
	// The preexisting indexes shouldn't be changed. It doesn't strictly need the new collateralType either.
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingCollateralType,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newCollateralType,
			RewardIndexes:  nil,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *InitializeClaimTests) TestClaimUpdatedWhenClaimExistsAndRewardsExist() {
	// When a claim exists, and a user deposits to a new rewarded pool;
	// then the claim's rewards don't change and the indexes are updated to match the global indexes

	preexistingCollateralType := "preexisting"
	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	newCollateralType := "btcb:usdx"
	newIndexes := types.RewardIndexes{
		{
			CollateralType: "otherrewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	claimType := types.CLAIM_TYPE_SWAP

	claim := types.Claim{
		Type:   claimType,
		Owner:  arbitraryAddress(),
		Reward: arbitraryCoins(),
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingCollateralType,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.keeper.Store.SetClaim(suite.ctx, claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingCollateralType,
			RewardIndexes:  increaseRewardFactors(preexistingIndexes),
		},
		{
			CollateralType: newCollateralType,
			RewardIndexes:  newIndexes,
		},
	}
	suite.storeGlobalIndexes(claimType, globalIndexes)

	suite.keeper.InitializeClaimSingleReward(suite.ctx, claimType, claim.Owner, newCollateralType)

	syncedClaim, _ := suite.keeper.Store.GetClaim(suite.ctx, claimType, claim.Owner)
	// only the indexes for the new pool should be updated
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingCollateralType,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newCollateralType,
			RewardIndexes:  newIndexes,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}
