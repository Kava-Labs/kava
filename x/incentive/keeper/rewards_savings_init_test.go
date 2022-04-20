package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// InitializeSavingsRewardTests runs unit tests for the keeper.InitializeSavingsReward method
type InitializeSavingsRewardTests struct {
	unitTester
}

func TestInitializeSavingsRewardTests(t *testing.T) {
	suite.Run(t, new(InitializeSavingsRewardTests))
}

func (suite *InitializeSavingsRewardTests) TestClaimAddedWhenClaimDoesNotExistAndNoRewards() {
	// When a claim doesn't exist, and a user deposits to a non-rewarded pool;
	// then a claim is added with no rewards and no indexes

	// no global indexes stored as this pool is not rewarded

	owner := arbitraryAddress()

	amount := sdk.NewCoin("test", sdk.OneInt())
	deposit := savingstypes.NewDeposit(owner, sdk.NewCoins(amount))

	suite.keeper.InitializeSavingsReward(suite.ctx, deposit)

	syncedClaim, found := suite.keeper.GetSavingsClaim(suite.ctx, owner)
	suite.True(found)
	// A new claim should have empty indexes. It doesn't strictly need the poolID either.
	expectedIndexes := types.MultiRewardIndexes{{
		CollateralType: amount.Denom,
		RewardIndexes:  nil,
	}}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeSavingsRewardTests) TestClaimAddedWhenClaimDoesNotExistAndRewardsExist() {
	// When a claim doesn't exist, and a user deposits to a rewarded pool;
	// then a claim is added with no rewards and indexes matching the global indexes

	amount := sdk.NewCoin("test", sdk.OneInt())

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: amount.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("1000.001"),
				},
			},
		},
	}
	suite.storeGlobalSavingsIndexes(globalIndexes)

	owner := arbitraryAddress()

	deposit := savingstypes.NewDeposit(owner, sdk.NewCoins(amount))
	suite.keeper.InitializeSavingsReward(suite.ctx, deposit)

	syncedClaim, found := suite.keeper.GetSavingsClaim(suite.ctx, owner)
	suite.True(found)
	// a new claim should start with the current global indexes
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// a new claim should start with 0 rewards
	suite.Equal(sdk.Coins(nil), syncedClaim.Reward)
}

func (suite *InitializeSavingsRewardTests) TestClaimUpdatedWhenClaimExistsAndNoRewards() {
	// When a claim exists, and a user deposits to a new non-rewarded denom;
	// then the claim's rewards don't change

	preexistingDenom := "preexisting"
	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingDenom,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.storeSavingsClaim(claim)

	// no global indexes stored as the new denom is not rewarded
	newDenom := "test"
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(newDenom, sdk.OneInt())))
	suite.keeper.InitializeSavingsReward(suite.ctx, deposit)

	syncedClaim, found := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	suite.True(found)

	// The preexisting indexes shouldn't be changed. It doesn't strictly need the new denom either.
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingDenom,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newDenom,
			RewardIndexes:  nil,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *InitializeSavingsRewardTests) TestClaimUpdatedWhenClaimExistsAndRewardsExist() {
	// When a claim exists, and a user deposits to a new rewarded denom;
	// then the claim's rewards don't change and the indexes are updated to match the global indexes

	preexistingDenom := "preexisting"
	preexistingIndexes := types.RewardIndexes{
		{
			CollateralType: "rewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	newDenom := "test"
	newIndexes := types.RewardIndexes{
		{
			CollateralType: "otherrewarddenom",
			RewardFactor:   d("1000.001"),
		},
	}

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: preexistingDenom,
				RewardIndexes:  preexistingIndexes,
			},
		},
	}
	suite.storeSavingsClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingDenom,
			RewardIndexes:  increaseRewardFactors(preexistingIndexes),
		},
		{
			CollateralType: newDenom,
			RewardIndexes:  newIndexes,
		},
	}
	suite.storeGlobalSavingsIndexes(globalIndexes)

	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(newDenom, sdk.OneInt())))
	suite.keeper.InitializeSavingsReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// only the indexes for the new denom should be updated
	expectedIndexes := types.MultiRewardIndexes{
		{
			CollateralType: preexistingDenom,
			RewardIndexes:  preexistingIndexes,
		},
		{
			CollateralType: newDenom,
			RewardIndexes:  newIndexes,
		},
	}
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// init should never alter the rewards
	suite.Equal(claim.Reward, syncedClaim.Reward)
}
