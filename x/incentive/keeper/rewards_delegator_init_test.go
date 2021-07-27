package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// InitializeDelegatorRewardTests runs unit tests for the keeper.InitializeDelegatorReward method
//
// inputs
// - claim in store if it exists (only claim.DelegatorRewardIndexes)
// - global indexes in store
// - delegator function arg
//
// outputs
// - sets or creates a claim
type InitializeDelegatorRewardTests struct {
	unitTester
}

func TestInitializeDelegatorReward(t *testing.T) {
	suite.Run(t, new(InitializeDelegatorRewardTests))
}

// Hardcoded to use bond denom
func (suite *InitializeDelegatorRewardTests) storeGlobalDelegatorFactor(multiRewardIndexes types.MultiRewardIndexes) {
	multiRewardIndex, _ := multiRewardIndexes.GetRewardIndex(types.BondDenom)
	suite.keeper.SetDelegatorRewardIndexes(suite.ctx, types.BondDenom, multiRewardIndex.RewardIndexes)
}

func (suite *InitializeDelegatorRewardTests) TestClaimIndexesAreSetWhenClaimDoesNotExist() {
	globalIndex := arbitraryDelegatorRewardIndexes
	suite.storeGlobalDelegatorIndexes(globalIndex)

	delegator := arbitraryAddress()
	suite.keeper.InitializeDelegatorReward(suite.ctx, delegator)

	syncedClaim, f := suite.keeper.GetDelegatorClaim(suite.ctx, delegator)
	suite.True(f)
	suite.Equal(globalIndex, syncedClaim.RewardIndexes)
}

func (suite *InitializeDelegatorRewardTests) TestClaimIsSyncedAndIndexesAreSetWhenClaimDoesExist() {
	validatorAddress := arbitraryValidatorAddress()
	sk := &fakeStakingKeeper{
		delegations: stakingtypes.Delegations{{
			ValidatorAddress: validatorAddress,
			Shares:           d("1000"),
		}},
		validators: stakingtypes.Validators{{
			OperatorAddress: validatorAddress,
			Status:          sdk.Bonded,
			Tokens:          i(1000),
			DelegatorShares: d("1000"),
		}},
	}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, sk, nil)

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		RewardIndexes: arbitraryDelegatorRewardIndexes,
	}
	suite.storeDelegatorClaim(claim)

	// Set the global factor to a value different to one in claim so
	// we can detect if it is overwritten.
	rewardIndexes, _ := claim.RewardIndexes.Get(types.BondDenom)
	globalIndexes := increaseRewardFactors(rewardIndexes)

	// Update the claim object with the new global factor
	bondIndex, _ := claim.RewardIndexes.GetRewardIndexIndex(types.BondDenom)
	claim.RewardIndexes[bondIndex].RewardIndexes = globalIndexes
	suite.storeGlobalDelegatorFactor(claim.RewardIndexes)

	suite.keeper.InitializeDelegatorReward(suite.ctx, claim.Owner)

	syncedClaim, _ := suite.keeper.GetDelegatorClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes[bondIndex].RewardIndexes)
	suite.Truef(syncedClaim.Reward.IsAllGT(claim.Reward), "'%s' not greater than '%s'", syncedClaim.Reward, claim.Reward)
}

// arbitraryDelegatorRewardIndexes contains only one reward index as there is only ever one bond denom
var arbitraryDelegatorRewardIndexes = types.MultiRewardIndexes{
	types.NewMultiRewardIndex(
		types.BondDenom,
		types.RewardIndexes{
			types.NewRewardIndex("hard", d("0.2")),
			types.NewRewardIndex("swp", d("0.2")),
		},
	),
}
