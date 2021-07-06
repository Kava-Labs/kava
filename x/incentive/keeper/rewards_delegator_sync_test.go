package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

// SynchronizeDelegatorRewardTests runs unit tests for the keeper.SynchronizeDelegatorReward method
//
// inputs
// - claim in store if it exists (only claim.DelegatorRewardIndexes and claim.Reward)
// - global index in store
// - function args: delegator address, validator address, shouldIncludeValidator flag
// - delegator's delegations and the corresponding validators
//
// outputs
// - sets or creates a claim
type SynchronizeDelegatorRewardTests struct {
	unitTester
}

func TestSynchronizeDelegatorReward(t *testing.T) {
	suite.Run(t, new(SynchronizeDelegatorRewardTests))
}

func (suite *SynchronizeDelegatorRewardTests) storeGlobalDelegatorFactor(multiRewardIndexes types.MultiRewardIndexes) {
	multiRewardIndex, _ := multiRewardIndexes.GetRewardIndex(types.BondDenom)
	suite.keeper.SetDelegatorRewardIndexes(suite.ctx, types.BondDenom, multiRewardIndex.RewardIndexes)
}

func (suite *SynchronizeDelegatorRewardTests) TestClaimIndexesAreUnchangedWhenGlobalFactorUnchanged() {
	delegator := arbitraryAddress()

	stakingKeeper := fakeStakingKeeper{} // use an empty staking keeper that returns no delegations
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper)

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: delegator,
		},
		RewardIndexes: arbitraryDelegatorRewardIndexes,
	}
	suite.storeDelegatorClaim(claim)

	suite.storeGlobalDelegatorFactor(claim.RewardIndexes)

	suite.keeper.SynchronizeDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetDelegatorClaim(suite.ctx, claim.Owner)
	suite.Equal(claim.RewardIndexes, syncedClaim.RewardIndexes)
}

func (suite *SynchronizeDelegatorRewardTests) TestClaimIndexesAreUpdatedWhenGlobalFactorIncreased() {
	delegator := arbitraryAddress()

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, fakeStakingKeeper{})

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: delegator,
		},
		RewardIndexes: arbitraryDelegatorRewardIndexes,
	}
	suite.storeDelegatorClaim(claim)

	rewardIndexes, _ := claim.RewardIndexes.Get(types.BondDenom)
	globalIndexes := increaseRewardFactors(rewardIndexes)

	// Update the claim object with the new global factor
	bondIndex, _ := claim.RewardIndexes.GetRewardIndexIndex(types.BondDenom)
	claim.RewardIndexes[bondIndex].RewardIndexes = globalIndexes
	suite.storeGlobalDelegatorFactor(claim.RewardIndexes)

	suite.keeper.SynchronizeDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetDelegatorClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes[bondIndex].RewardIndexes)
}

func (suite *SynchronizeDelegatorRewardTests) TestRewardIsUnchangedWhenGlobalFactorUnchanged() {
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

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  delegator,
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{{
			CollateralType: types.BondDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard", RewardFactor: d("0.1"),
				},
				{
					CollateralType: "swp", RewardFactor: d("0.2"),
				},
			},
		}},
	}
	suite.storeDelegatorClaim(claim)

	suite.storeGlobalDelegatorFactor(claim.RewardIndexes)

	suite.keeper.SynchronizeDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetDelegatorClaim(suite.ctx, claim.Owner)

	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *SynchronizeDelegatorRewardTests) TestRewardIsIncreasedWhenNewRewardAdded() {
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

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  delegator,
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{},
	}
	suite.storeDelegatorClaim(claim)

	newGlobalIndexes := types.MultiRewardIndexes{{
		CollateralType: types.BondDenom,
		RewardIndexes: types.RewardIndexes{
			{
				CollateralType: "hard", RewardFactor: d("0.1"),
			},
			{
				CollateralType: "swp", RewardFactor: d("0.2"),
			},
		},
	}}
	suite.storeGlobalDelegatorFactor(newGlobalIndexes)

	suite.keeper.SynchronizeDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetDelegatorClaim(suite.ctx, claim.Owner)

	suite.Equal(newGlobalIndexes, syncedClaim.RewardIndexes)
	suite.Equal(
		cs(
			c(types.HardLiquidityRewardDenom, 100),
			c("swp", 200),
		).Add(claim.Reward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeDelegatorRewardTests) TestRewardIsIncreasedWhenGlobalFactorIncreased() {
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

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  delegator,
			Reward: arbitraryCoins(),
		},
		RewardIndexes: types.MultiRewardIndexes{{
			CollateralType: types.BondDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard", RewardFactor: d("0.1"),
				},
				{
					CollateralType: "swp", RewardFactor: d("0.2"),
				},
			},
		}},
	}
	suite.storeDelegatorClaim(claim)

	suite.storeGlobalDelegatorFactor(
		types.MultiRewardIndexes{
			types.NewMultiRewardIndex(
				types.BondDenom,
				types.RewardIndexes{
					{
						CollateralType: "hard", RewardFactor: d("0.2"),
					},
					{
						CollateralType: "swp", RewardFactor: d("0.4"),
					},
				},
			),
		},
	)

	suite.keeper.SynchronizeDelegatorRewards(suite.ctx, claim.Owner, nil, false)

	syncedClaim, _ := suite.keeper.GetDelegatorClaim(suite.ctx, claim.Owner)

	suite.Equal(
		cs(
			c(types.HardLiquidityRewardDenom, 100),
			c("swp", 200),
		).Add(claim.Reward...),
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

func (suite *SynchronizeDelegatorRewardTests) TestGetDelegatedWhenValAddrIsNil() {
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
func (suite *SynchronizeDelegatorRewardTests) TestGetDelegatedWhenExcludingAValidator() {
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
func (suite *SynchronizeDelegatorRewardTests) TestGetDelegatedWhenIncludingAValidator() {
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
