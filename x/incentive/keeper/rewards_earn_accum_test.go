package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

type AccumulateEarnRewardsTests struct {
	unitTester
}

func (suite *AccumulateEarnRewardsTests) storedTimeEquals(vaultDenom string, expected time.Time) {
	storedTime, found := suite.keeper.GetEarnRewardAccrualTime(suite.ctx, vaultDenom)
	suite.Equal(found, expected != time.Time{}, "expected time is %v but time found = %v", expected, found)
	if found {
		suite.Equal(expected, storedTime)
	} else {
		suite.Empty(storedTime)
	}
}

func (suite *AccumulateEarnRewardsTests) storedIndexesEqual(vaultDenom string, expected types.RewardIndexes) {
	storedIndexes, found := suite.keeper.GetEarnRewardIndexes(suite.ctx, vaultDenom)
	suite.Equal(found, expected != nil, "expected indexes is %v but indexes found = %v", expected, found)
	if found {
		suite.Equal(expected, storedIndexes)
	} else {
		suite.Empty(storedIndexes)
	}
}

func TestAccumulateEarnRewards(t *testing.T) {
	suite.Run(t, new(AccumulateEarnRewardsTests))
}

func (suite *AccumulateEarnRewardsTests) TestStateUpdatedWhenBlockTimeHasIncreased() {
	vaultDenom := "usdx"

	earnKeeper := newFakeEarnKeeper().addVault(vaultDenom, earntypes.NewVaultShare(vaultDenom, d("1000000")))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	suite.storeGlobalEarnIndexes(types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	})
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(newAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		vaultDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(vaultDenom, newAccrualTime)
	suite.storedIndexesEqual(vaultDenom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("3.64"),
		},
	})
}

func (suite *AccumulateEarnRewardsTests) TestStateUpdatedWhenBlockTimeHasIncreased_bkava() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	earnKeeper := newFakeEarnKeeper().
		addVault(vaultDenom1, earntypes.NewVaultShare(vaultDenom1, d("800000"))).
		addVault(vaultDenom2, earntypes.NewVaultShare(vaultDenom2, d("200000")))

	liquidKeeper := newFakeLiquidKeeper().
		addDerivative(suite.ctx, vaultDenom1, i(800000)).
		addDerivative(suite.ctx, vaultDenom2, i(200000))

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, liquidKeeper, earnKeeper)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: vaultDenom2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
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
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)
	suite.keeper.AccumulateEarnRewards(suite.ctx, rewardPeriod)

	// check time and factors

	suite.storedTimeEquals(vaultDenom1, newAccrualTime)
	suite.storedTimeEquals(vaultDenom2, newAccrualTime)

	// Each vault gets the same ukava per second, assuming shares prices are the same.
	// The share amount determines how much is actually distributed to the vault.
	expectedIndexes := types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor: d("3.64"). // base incentive
							Add(d("360")), // staking rewards, 10% of total bkava per second
		},
	}

	suite.storedIndexesEqual(vaultDenom1, expectedIndexes)
	suite.storedIndexesEqual(vaultDenom2, expectedIndexes)
}

func (suite *AccumulateEarnRewardsTests) TestStateUpdatedWhenBlockTimeHasIncreased_bkava_partialDeposit() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	vaultDenom1Supply := i(800000)
	vaultDenom2Supply := i(200000)

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

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

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: vaultDenom2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
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
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)
	suite.keeper.AccumulateEarnRewards(suite.ctx, rewardPeriod)

	// check time and factors

	suite.storedTimeEquals(vaultDenom1, newAccrualTime)
	suite.storedTimeEquals(vaultDenom2, newAccrualTime)

	// Slightly increased rewards due to less bkava deposited
	suite.storedIndexesEqual(vaultDenom1, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("8.248571428571428571"),
		},
		{
			CollateralType: "ukava",
			RewardFactor: d("4.154285714285714286"). // base incentive
									Add(sdk.NewDecFromInt(vaultDenom1Supply). // staking rewards
															QuoInt64(10).
															MulInt64(3600).
															Quo(vault1Shares),
				),
		},
	})

	// Much higher rewards per share because only a small amount of bkava is
	// deposited. The **total** amount of incentives distributed to this vault
	// is still the same proportional amount.

	// Fixed amount total rewards distributed to the vault
	// Fewer shares deposited -> higher rewards per share

	// 7.2ukava shares per second for 1 hour (started with 0.04)
	// total rewards claimable = 7.2 * 100000 shares = 720000 ukava

	// 720000ukava distributed which is 20% of total bkava ukava rewards
	// total rewards for *all* bkava vaults for 1 hour
	// = 1000ukava per second * 3600 == 3600000ukava
	// vaultDenom2 has 20% of the total bkava amount so it should get 20% of 3600000ukava == 720000ukava

	vault2expectedIndexes := types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("14.42"),
		},
		{
			CollateralType: "ukava",
			RewardFactor: d("7.24").
				Add(sdk.NewDecFromInt(vaultDenom2Supply).
					QuoInt64(10).
					MulInt64(3600).
					Quo(vault2Shares),
				),
		},
	}
	suite.storedIndexesEqual(vaultDenom2, vault2expectedIndexes)
}

func (suite *AccumulateEarnRewardsTests) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	vaultDenom := "usdx"

	earnKeeper := newFakeEarnKeeper().addVault(vaultDenom, earntypes.NewVaultShare(vaultDenom, d("1000000")))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom, previousAccrualTime)

	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		vaultDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(vaultDenom, previousAccrualTime)
	expected, f := previousIndexes.Get(vaultDenom)
	suite.True(f)
	suite.storedIndexesEqual(vaultDenom, expected)
}

func (suite *AccumulateEarnRewardsTests) TestStateUnchangedWhenBlockTimeHasNotIncreased_bkava() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	earnKeeper := newFakeEarnKeeper().
		addVault(vaultDenom1, earntypes.NewVaultShare(vaultDenom1, d("1000000"))).
		addVault(vaultDenom2, earntypes.NewVaultShare(vaultDenom2, d("1000000")))

	liquidKeeper := newFakeLiquidKeeper().
		addDerivative(suite.ctx, vaultDenom1, i(1000000)).
		addDerivative(suite.ctx, vaultDenom2, i(1000000))

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, liquidKeeper, earnKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: vaultDenom2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(previousIndexes)

	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom1, previousAccrualTime)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom2, previousAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		"bkava",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(vaultDenom1, previousAccrualTime)
	suite.storedTimeEquals(vaultDenom2, previousAccrualTime)

	expected, f := previousIndexes.Get(vaultDenom1)
	suite.True(f)
	suite.storedIndexesEqual(vaultDenom1, expected)

	expected, f = previousIndexes.Get(vaultDenom2)
	suite.True(f)
	suite.storedIndexesEqual(vaultDenom2, expected)
}

func (suite *AccumulateEarnRewardsTests) TestNoAccumulationWhenSourceSharesAreZero() {
	vaultDenom := "usdx"

	earnKeeper := newFakeEarnKeeper() // no vault, so no source shares
	liquidKeeper := newFakeLiquidKeeper()

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, liquidKeeper, earnKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(7 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		vaultDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(vaultDenom, firstAccrualTime)
	expected, f := previousIndexes.Get(vaultDenom)
	suite.True(f)
	suite.storedIndexesEqual(vaultDenom, expected)
}

func (suite *AccumulateEarnRewardsTests) TestNoAccumulationWhenSourceSharesAreZero_bkava() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	earnKeeper := newFakeEarnKeeper() // no vault, so no source shares
	liquidKeeper := newFakeLiquidKeeper()

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, liquidKeeper, earnKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom1,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: vaultDenom2,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom1, previousAccrualTime)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom2, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(7 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		"bkava",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	// TODO: There are no bkava vaults to iterate over, so the accrual times are
	// not updated
	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(vaultDenom1, firstAccrualTime)
	suite.storedTimeEquals(vaultDenom2, firstAccrualTime)

	expected, f := previousIndexes.Get(vaultDenom1)
	suite.True(f)
	suite.storedIndexesEqual(vaultDenom1, expected)

	expected, f = previousIndexes.Get(vaultDenom2)
	suite.True(f)
	suite.storedIndexesEqual(vaultDenom2, expected)
}

func (suite *AccumulateEarnRewardsTests) TestStateAddedWhenStateDoesNotExist() {
	vaultDenom := "usdx"

	earnKeeper := newFakeEarnKeeper().addVault(vaultDenom, earntypes.NewVaultShare(vaultDenom, d("1000000")))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	period := types.NewMultiRewardPeriod(
		true,
		vaultDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)),
	)

	firstAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.storedTimeEquals(vaultDenom, firstAccrualTime)
	suite.storedIndexesEqual(vaultDenom, nil)

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(secondAccrualTime)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// After the second accumulation both current block time and indexes should be stored.
	suite.storedTimeEquals(vaultDenom, secondAccrualTime)
	suite.storedIndexesEqual(vaultDenom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("0.02"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("0.01"),
		},
	})
}

func (suite *AccumulateEarnRewardsTests) TestStateAddedWhenStateDoesNotExist_bkava() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	firstAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	earnKeeper := newFakeEarnKeeper().
		addVault(vaultDenom1, earntypes.NewVaultShare(vaultDenom1, d("1000000"))).
		addVault(vaultDenom2, earntypes.NewVaultShare(vaultDenom2, d("1000000")))

	liquidKeeper := newFakeLiquidKeeper().
		addDerivative(suite.ctx, vaultDenom1, i(1000000)).
		addDerivative(suite.ctx, vaultDenom2, i(1000000))

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, liquidKeeper, earnKeeper)

	period := types.NewMultiRewardPeriod(
		true,
		"bkava",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)),
	)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.storedTimeEquals(vaultDenom1, firstAccrualTime)
	suite.storedTimeEquals(vaultDenom2, firstAccrualTime)

	suite.storedIndexesEqual(vaultDenom1, nil)
	suite.storedIndexesEqual(vaultDenom2, nil)

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(secondAccrualTime)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// After the second accumulation both current block time and indexes should be stored.
	suite.storedTimeEquals(vaultDenom1, secondAccrualTime)
	suite.storedTimeEquals(vaultDenom2, secondAccrualTime)

	expectedIndexes := types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("0.01"),
		},
		{
			CollateralType: "ukava",
			// 10% of total bkava for rewards per second for 10 seconds
			// 1ukava per share per second + regular 0.005ukava incentive rewards
			RewardFactor: d("1.005"),
		},
	}

	suite.storedIndexesEqual(vaultDenom1, expectedIndexes)
	suite.storedIndexesEqual(vaultDenom2, expectedIndexes)
}

func (suite *AccumulateEarnRewardsTests) TestNoPanicWhenStateDoesNotExist() {
	vaultDenom := "usdx"

	earnKeeper := newFakeEarnKeeper()
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	period := types.NewMultiRewardPeriod(
		true,
		vaultDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(),
	)

	accrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(accrualTime)

	// Accumulate with no earn shares and no rewards per second will result in no increment to the indexes.
	// No increment and no previous indexes stored, results in an updated of nil. Setting this in the state panics.
	// Check there is no panic.
	suite.NotPanics(func() {
		suite.keeper.AccumulateEarnRewards(suite.ctx, period)
	})

	suite.storedTimeEquals(vaultDenom, accrualTime)
	suite.storedIndexesEqual(vaultDenom, nil)
}

func (suite *AccumulateEarnRewardsTests) TestNoPanicWhenStateDoesNotExist_bkava() {
	vaultDenom1 := "bkava-meow"
	vaultDenom2 := "bkava-woof"

	earnKeeper := newFakeEarnKeeper()
	liquidKeeper := newFakeLiquidKeeper()

	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, liquidKeeper, earnKeeper)

	period := types.NewMultiRewardPeriod(
		true,
		"bkava",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(),
	)

	accrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(accrualTime)

	// Accumulate with no earn shares and no rewards per second will result in no increment to the indexes.
	// No increment and no previous indexes stored, results in an updated of nil. Setting this in the state panics.
	// Check there is no panic.
	suite.NotPanics(func() {
		// This does not update any state, as there are no bkava vaults
		// to iterate over, denoms are unknown
		suite.keeper.AccumulateEarnRewards(suite.ctx, period)
	})

	// Times are not stored for vaults with no state
	suite.storedTimeEquals(vaultDenom1, time.Time{})
	suite.storedTimeEquals(vaultDenom2, time.Time{})
	suite.storedIndexesEqual(vaultDenom1, nil)
	suite.storedIndexesEqual(vaultDenom2, nil)
}

func (suite *AccumulateEarnRewardsTests) TestNoAccumulationWhenBeforeStartTime() {
	vaultDenom := "usdx"

	earnKeeper := newFakeEarnKeeper().addVault(vaultDenom, earntypes.NewVaultShare(vaultDenom, d("1000000")))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: vaultDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalEarnIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(10 * time.Second)

	period := types.NewMultiRewardPeriod(
		true,
		vaultDenom,
		firstAccrualTime.Add(time.Nanosecond), // start time after accrual time
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateEarnRewards(suite.ctx, period)

	// The accrual time should be updated, but the indexes unchanged
	suite.storedTimeEquals(vaultDenom, firstAccrualTime)
	expectedIndexes, f := previousIndexes.Get(vaultDenom)
	suite.True(f)
	suite.storedIndexesEqual(vaultDenom, expectedIndexes)
}

func (suite *AccumulateEarnRewardsTests) TestPanicWhenCurrentTimeLessThanPrevious() {
	vaultDenom := "usdx"

	earnKeeper := newFakeEarnKeeper().addVault(vaultDenom, earntypes.NewVaultShare(vaultDenom, d("1000000")))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, nil, nil, nil, earnKeeper)

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetEarnRewardAccrualTime(suite.ctx, vaultDenom, previousAccrualTime)

	firstAccrualTime := time.Time{}

	period := types.NewMultiRewardPeriod(
		true,
		vaultDenom,
		time.Time{}, // start time after accrual time
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.Panics(func() {
		suite.keeper.AccumulateEarnRewards(suite.ctx, period)
	})
}
