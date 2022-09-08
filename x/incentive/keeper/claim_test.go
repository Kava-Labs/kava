package keeper_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"

	acctypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// ClaimTests runs unit tests for the keeper Claim methods
type ClaimTests struct {
	unitTester
}

func TestClaim(t *testing.T) {
	suite.Run(t, new(ClaimTests))
}

func (suite *ClaimTests) ErrorIs(err, target error) bool {
	return suite.Truef(errors.Is(err, target), "err didn't match: %s, it was: %s", target, err)
}

func (suite *ClaimTests) TestCannotClaimWhenMultiplierNotRecognised() {
	subspace := &fakeParamSubspace{
		params: types.Params{
			ClaimMultipliers: types.MultipliersPerDenoms{
				{
					Denom: "hard",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.2")),
					},
				},
			},
		},
	}
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
	}
	suite.storeDelegatorClaim(claim)

	// multiplier not in params
	err := suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, "hard", "large")
	suite.ErrorIs(err, types.ErrInvalidMultiplier)

	// invalid multiplier name
	err = suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, "hard", "")
	suite.ErrorIs(err, types.ErrInvalidMultiplier)
}

func (suite *ClaimTests) TestCannotClaimAfterEndTime() {
	endTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	subspace := &fakeParamSubspace{
		params: types.Params{
			ClaimMultipliers: types.MultipliersPerDenoms{
				{
					Denom: "hard",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.2")),
					},
				},
			},
			ClaimEnd: endTime,
		},
	}
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	suite.ctx = suite.ctx.WithBlockTime(endTime.Add(time.Nanosecond))

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
	}
	suite.storeDelegatorClaim(claim)

	err := suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, "hard", "small")
	suite.ErrorIs(err, types.ErrClaimExpired)
}

func (suite *ClaimTests) TestClaimEarnNoLockup() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	suite.ctx = NewTestContext(
		suite.incentiveStoreKey,
		suite.app.GetStoreKey(acctypes.StoreKey),
		suite.app.GetStoreKey(banktypes.StoreKey),
	)

	err := suite.app.FundModuleAccount(suite.ctx, types.IncentiveMacc, cs(c("ukava", 100000000000)))
	suite.NoError(err)

	accAddr := arbitraryAddress()

	ak := suite.app.GetAccountKeeper()
	acc := ak.NewAccountWithAddress(suite.ctx, accAddr)
	ak.SetAccount(suite.ctx, acc)

	bk := suite.app.GetBankKeeper()

	earnKeeper := newFakeEarnKeeper().
		addVault(vaultDenom1, earntypes.NewVaultShare(vaultDenom1, d("1000000"))).
		addVault(vaultDenom2, earntypes.NewVaultShare(vaultDenom2, d("1000000")))

	liquidKeeper := newFakeLiquidKeeper().
		addDerivative(vaultDenom1, i(1000000)).
		addDerivative(vaultDenom2, i(1000000))

	suite.keeper = suite.NewKeeper(
		&fakeParamSubspace{},
		bk,
		nil, nil,
		ak,
		nil, nil, nil,
		liquidKeeper, earnKeeper,
	)

	suite.keeper.SetParams(suite.ctx, types.Params{
		ClaimMultipliers: types.MultipliersPerDenoms{
			{
				Denom: "ukava",
				Multipliers: types.Multipliers{
					types.NewMultiplier("large", 0, d("1")),
				},
				ModuleName: earntypes.ModuleName,
			},
			{
				Denom: "ukava",
				Multipliers: types.Multipliers{
					types.NewMultiplier("small", 1, d("0.2")),
				},
				// No module name to apply to other non-earn modules
				ModuleName: "",
			},
		},
		ClaimEnd: distantFuture,
	})

	earnClaim := types.EarnClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  accAddr,
			Reward: cs(c("earn", 100), c("ukava", 100)),
		},
	}
	suite.storeEarnClaim(earnClaim)

	claim := types.DelegatorClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  accAddr,
			Reward: cs(c("earn", 100), c("ukava", 100)),
		},
	}
	suite.storeDelegatorClaim(claim)

	balBefore := bk.GetAllBalances(suite.ctx, accAddr)
	suite.Equal(cs(), balBefore)

	// Claim for earn module
	err = suite.keeper.ClaimEarnReward(suite.ctx, earnClaim.Owner, earnClaim.Owner, "ukava", "large")
	suite.NoError(err)

	// Check balances
	balAfter := bk.GetAllBalances(suite.ctx, accAddr)
	suite.Equal(cs(c("ukava", 100)), balAfter)

	// Claim for non-earn module
	err = suite.keeper.ClaimDelegatorReward(suite.ctx, claim.Owner, claim.Owner, "ukava", "small")
	suite.NoError(err)

	balAfter2 := bk.GetAllBalances(suite.ctx, accAddr)
	suite.Equal(cs(c("ukava", 120)), balAfter2, "claiming ukava for non-earn is multiplied by 0.2")
}
