package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// SynchronizeHardDelegatorRewardTests runs unit tests for the keeper.SynchronizeHardDelegatorReward method
//
// inputs
// - claim in store if it exists (only claim.DelegatorRewardIndexes and claim.Reward)
// - global index in store
// - function args: delegator address, validator address, shouldIncludeValidator flag
// - delegator's delegations and the corresponding validators
//
// outputs
// - sets or creates a claim
type SynchronizeHardDelegatorRewardTests struct {
	unitTester
}

func TestSynchronizeHardDelegatorReward(t *testing.T) {
	suite.Run(t, new(SynchronizeHardDelegatorRewardTests))
}

func (suite *SynchronizeHardDelegatorRewardTests) storeGlobalDelegatorFactor(rewardIndexes types.RewardIndexes) {
	factor := rewardIndexes[0]
	suite.keeper.SetHardDelegatorRewardFactor(suite.ctx, factor.CollateralType, factor.RewardFactor)
}

func (suite *SynchronizeHardDelegatorRewardTests) TestClaimIndexesAreUnchangedWhenGlobalFactorUnchanged() {
	delegator := arbitraryAddress()

	stakingKeeper := fakeStakingKeeper{} // use an empty staking keeper that returns no delegations
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper)

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: delegator,
		},
		DelegatorRewardIndexes: arbitraryDelegatorRewardIndexes,
	}
	suite.storeClaim(claim)

	suite.storeGlobalDelegatorFactor(claim.DelegatorRewardIndexes)

	suite.keeper.SynchronizeHardDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(claim.DelegatorRewardIndexes, syncedClaim.DelegatorRewardIndexes)
}

func (suite *SynchronizeHardDelegatorRewardTests) TestClaimIndexesAreUpdatedWhenGlobalFactorIncreased() {
	delegator := arbitraryAddress()

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, fakeStakingKeeper{})

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: delegator,
		},
		DelegatorRewardIndexes: arbitraryDelegatorRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := increaseRewardFactors(claim.DelegatorRewardIndexes)
	suite.storeGlobalDelegatorFactor(globalIndexes)

	suite.keeper.SynchronizeHardDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.DelegatorRewardIndexes)
}
func (suite *SynchronizeHardDelegatorRewardTests) TestRewardIsUnchangedWhenGlobalFactorUnchanged() {
	delegator := arbitraryAddress()
	validatorAddress := arbitraryValidatorAddress()
	stakingKeeper := fakeStakingKeeper{
		delegations: stakingtypes.Delegations{
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddress,
				Shares:           d("1000"),
			},
		},
		validators: stakingtypes.Validators{
			unslashedBondedValidator(validatorAddress),
		},
	}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper)

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  delegator,
			Reward: arbitraryCoins(),
		},
		DelegatorRewardIndexes: types.RewardIndexes{{
			CollateralType: types.BondDenom,
			RewardFactor:   d("0.1"),
		}},
	}
	suite.storeClaim(claim)

	suite.storeGlobalDelegatorFactor(claim.DelegatorRewardIndexes)

	suite.keeper.SynchronizeHardDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)

	suite.Equal(claim.Reward, syncedClaim.Reward)
}

// No test for TestRewardIsIncrementedWhenNewRewardAdded as we can currently only have one reward denom for our one bond denom.
// So new rewards cannot be added.

func (suite *SynchronizeHardDelegatorRewardTests) TestRewardIsIncreasedWhenGlobalFactorIncreased() {
	delegator := arbitraryAddress()
	validatorAddress := arbitraryValidatorAddress()
	stakingKeeper := fakeStakingKeeper{
		delegations: stakingtypes.Delegations{
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddress,
				Shares:           d("1000"),
			},
		},
		validators: stakingtypes.Validators{
			unslashedBondedValidator(validatorAddress),
		},
	}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper)

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  delegator,
			Reward: arbitraryCoins(),
		},
		DelegatorRewardIndexes: types.RewardIndexes{{
			CollateralType: types.BondDenom,
			RewardFactor:   d("0.1"),
		}},
	}
	suite.storeClaim(claim)

	suite.storeGlobalDelegatorFactor(
		types.RewardIndexes{
			types.NewRewardIndex(types.BondDenom, d("0.2")),
		},
	)

	suite.keeper.SynchronizeHardDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)

	suite.Equal(
		cs(c(types.HardLiquidityRewardDenom, 100)).Add(claim.Reward...),
		syncedClaim.Reward,
	)
}

func unslashedBondedValidator(address sdk.ValAddress) stakingtypes.Validator {
	return stakingtypes.Validator{
		OperatorAddress: address,
		Status:          sdk.Bonded,

		// Set the tokens and shares equal so then
		// a _delegator's_ token amount is equal to their shares amount
		Tokens:          i(1e12),
		DelegatorShares: i(1e12).ToDec(),
	}
}
func unslashedNotBondedValidator(address sdk.ValAddress) stakingtypes.Validator {
	return stakingtypes.Validator{
		OperatorAddress: address,
		Status:          sdk.Unbonding,

		// Set the tokens and shares equal so then
		// a _delegator's_ token amount is equal to their shares amount
		Tokens:          i(1e12),
		DelegatorShares: i(1e12).ToDec(),
	}
}

func (suite *SynchronizeHardDelegatorRewardTests) TestGetDelegatedWhenValAddrIsNil() {
	// when valAddr is nil, get total delegated to bonded validators
	delegator := arbitraryAddress()
	validatorAddresses := generateValidatorAddresses(4)
	stakingKeeper := fakeStakingKeeper{
		delegations: stakingtypes.Delegations{
			//bonded
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[0],
				Shares:           d("1"),
			},
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[1],
				Shares:           d("10"),
			},
			// not bonded
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[2],
				Shares:           d("100"),
			},
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[3],
				Shares:           d("1000"),
			},
		},
		validators: stakingtypes.Validators{
			unslashedBondedValidator(validatorAddresses[0]),
			unslashedBondedValidator(validatorAddresses[1]),
			unslashedNotBondedValidator(validatorAddresses[2]),
			unslashedNotBondedValidator(validatorAddresses[3]),
		},
	}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper)

	suite.Equal(
		d("11"), // delegation to bonded validators
		suite.keeper.GetTotalDelegated(suite.ctx, delegator, nil, false),
	)
}
func (suite *SynchronizeHardDelegatorRewardTests) TestGetDelegatedWhenExcludingAValidator() {
	// when valAddr is x, get total delegated to bonded validators excluding those to x
	delegator := arbitraryAddress()
	validatorAddresses := generateValidatorAddresses(4)
	stakingKeeper := fakeStakingKeeper{
		delegations: stakingtypes.Delegations{
			//bonded
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[0],
				Shares:           d("1"),
			},
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[1],
				Shares:           d("10"),
			},
			// not bonded
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[2],
				Shares:           d("100"),
			},
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[3],
				Shares:           d("1000"),
			},
		},
		validators: stakingtypes.Validators{
			unslashedBondedValidator(validatorAddresses[0]),
			unslashedBondedValidator(validatorAddresses[1]),
			unslashedNotBondedValidator(validatorAddresses[2]),
			unslashedNotBondedValidator(validatorAddresses[3]),
		},
	}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper)

	suite.Equal(
		d("10"),
		suite.keeper.GetTotalDelegated(suite.ctx, delegator, validatorAddresses[0], false),
	)
}
func (suite *SynchronizeHardDelegatorRewardTests) TestGetDelegatedWhenIncludingAValidator() {
	// when valAddr is x, get total delegated to bonded validators including those to x
	delegator := arbitraryAddress()
	validatorAddresses := generateValidatorAddresses(4)
	stakingKeeper := fakeStakingKeeper{
		delegations: stakingtypes.Delegations{
			//bonded
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[0],
				Shares:           d("1"),
			},
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[1],
				Shares:           d("10"),
			},
			// not bonded
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[2],
				Shares:           d("100"),
			},
			{
				DelegatorAddress: delegator,
				ValidatorAddress: validatorAddresses[3],
				Shares:           d("1000"),
			},
		},
		validators: stakingtypes.Validators{
			unslashedBondedValidator(validatorAddresses[0]),
			unslashedBondedValidator(validatorAddresses[1]),
			unslashedNotBondedValidator(validatorAddresses[2]),
			unslashedNotBondedValidator(validatorAddresses[3]),
		},
	}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper)

	suite.Equal(
		d("111"),
		suite.keeper.GetTotalDelegated(suite.ctx, delegator, validatorAddresses[2], true),
	)
}
