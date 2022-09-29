package keeper_test

import (
	"time"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *AccumulateEarnRewardsTests) TestStakingRewardsDistributed() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	vaultDenom1Supply := i(800000)
	vaultDenom2Supply := i(200000)

	liquidKeeper := newFakeLiquidKeeper().
		addDerivative(suite.ctx, vaultDenom1, vaultDenom1Supply).
		addDerivative(suite.ctx, vaultDenom2, vaultDenom2Supply)

	vault1Shares := d("700000")
	vault2Shares := d("100000")

	// More bkava minted than deposited into earn
	// Rewards are higher per-share as a result
	earnKeeper := newFakeEarnKeeper().
		addVault(vaultDenom1, earntypes.NewVaultShare(vaultDenom1, vault1Shares)).
		addVault(vaultDenom2, earntypes.NewVaultShare(vaultDenom2, vault2Shares))

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, liquidKeeper, earnKeeper)

	initialVault1RewardFactor := d("0.04")
	initialVault2RewardFactor := d("0.04")

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "ukava",
					RewardFactor:   initialVault1RewardFactor,
				},
			},
		},
		{
			CollateralType: vaultDenom2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "ukava",
					RewardFactor:   initialVault2RewardFactor,
				},
			},
		},
	}

	suite.storeGlobalEarnIndexes(globalIndexes)

	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom1, previousAccrualTime)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom2, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(newAccrualTime)

	rewardPeriod := types.NewMultiRewardPeriod(
		true,
		"bkava",         // reward period is set for "bkava" to apply to all vaults
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(), // no incentives, so only the staking rewards are distributed
	)
	suite.keeper.AccumulateEarnRewards(suite.ctx, rewardPeriod)

	// check time and factors

	suite.storedTimeEquals(vaultDenom1, newAccrualTime)
	suite.storedTimeEquals(vaultDenom2, newAccrualTime)

	// Only contains staking rewards
	suite.storedIndexesEqual(vaultDenom1, types.RewardIndexes{
		{
			CollateralType: "ukava",
			RewardFactor: initialVault1RewardFactor.
				Add(vaultDenom1Supply.ToDec().
					QuoInt64(10).
					MulInt64(3600).
					Quo(vault1Shares)),
		},
	})

	suite.storedIndexesEqual(vaultDenom2, types.RewardIndexes{
		{
			CollateralType: "ukava",
			RewardFactor: initialVault2RewardFactor.
				Add(vaultDenom2Supply.ToDec().
					QuoInt64(10).
					MulInt64(3600).
					Quo(vault2Shares)),
		},
	})
}
