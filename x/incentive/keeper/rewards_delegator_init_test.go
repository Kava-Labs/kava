package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// InitializeHardDelegatorRewardTests runs unit tests for the keeper.InitializeHardDelegatorReward method
//
// inputs
// - claim in store if it exists (only claim.DelegatorRewardIndexes)
// - global indexes in store
// - delegator function arg
//
// outputs
// - sets or creates a claim
type InitializeHardDelegatorRewardTests struct {
	unitTester
}

func TestInitializeHardDelegatorReward(t *testing.T) {
	suite.Run(t, new(InitializeHardDelegatorRewardTests))
}

// Hardcoded to use bond denom
func (suite *InitializeHardDelegatorRewardTests) storeGlobalDelegatorFactor(multiRewardIndexes types.MultiRewardIndexes) {
	multiRewardIndex, _ := multiRewardIndexes.GetRewardIndex(types.BondDenom)
	suite.keeper.SetHardDelegatorRewardIndexes(suite.ctx, types.BondDenom, multiRewardIndex.RewardIndexes)
}

func (suite *InitializeHardDelegatorRewardTests) TestClaimIndexesAreSetWhenClaimDoesNotExist() {
	globalIndex := arbitraryDelegatorRewardIndexes
	suite.storeGlobalDelegatorFactor(globalIndex)

	delegator := arbitraryAddress()
	suite.keeper.InitializeHardDelegatorReward(suite.ctx, delegator)

	syncedClaim, f := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, delegator)
	suite.True(f)
	suite.Equal(globalIndex, syncedClaim.DelegatorRewardIndexes)
}

func (suite *InitializeHardDelegatorRewardTests) TestClaimIsSyncedAndIndexesAreSetWhenClaimDoesExist() {
	validatorAddress := arbitraryValidatorAddress()
	sk := fakeStakingKeeper{
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
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, sk)

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		DelegatorRewardIndexes: arbitraryDelegatorRewardIndexes,
	}
	suite.storeClaim(claim)

	// Set the global factor to a value different to one in claim so
	// we can detect if it is overwritten.
	rewardIndexes, _ := claim.DelegatorRewardIndexes.Get(types.BondDenom)
	globalIndexes := increaseRewardFactors(rewardIndexes)

	// Update the claim object with the new global factor
	bondIndex, _ := claim.DelegatorRewardIndexes.GetRewardIndexIndex(types.BondDenom)
	claim.DelegatorRewardIndexes[bondIndex].RewardIndexes = globalIndexes
	suite.storeGlobalDelegatorFactor(claim.DelegatorRewardIndexes)

	suite.keeper.InitializeHardDelegatorReward(suite.ctx, claim.Owner)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.DelegatorRewardIndexes[bondIndex].RewardIndexes)
	suite.Truef(syncedClaim.Reward.IsAllGT(claim.Reward), "'%s' not greater than '%s'", syncedClaim.Reward, claim.Reward)
}

// arbitraryDelegatorRewardIndexes contains only one reward index as there is only every one bond denom
var arbitraryDelegatorRewardIndexes = types.MultiRewardIndexes{
	types.NewMultiRewardIndex(
		types.BondDenom,
		types.RewardIndexes{
			types.NewRewardIndex("hard", d("0.2")),
			types.NewRewardIndex("swp", d("0.2")),
		},
	),
}

type fakeStakingKeeper struct {
	delegations stakingtypes.Delegations
	validators  stakingtypes.Validators
}

func (k fakeStakingKeeper) TotalBondedTokens(ctx sdk.Context) sdk.Int {
	panic("unimplemented")
}
func (k fakeStakingKeeper) GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) []stakingtypes.Delegation {
	return k.delegations
}
func (k fakeStakingKeeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	for _, val := range k.validators {
		if val.GetOperator().Equals(addr) {
			return val, true
		}
	}
	return stakingtypes.Validator{}, false
}
func (k fakeStakingKeeper) GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) []stakingtypes.Delegation {
	var delegations stakingtypes.Delegations
	for _, d := range k.delegations {
		if d.ValidatorAddress.Equals(valAddr) {
			delegations = append(delegations, d)
		}
	}
	return delegations
}
