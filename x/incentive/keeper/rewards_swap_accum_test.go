package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

type AccumulateSwapRewardsTests struct {
	unitTester
}

func (suite *AccumulateSwapRewardsTests) storedTimeEquals(poolID string, expected time.Time) {
	storedTime, found := suite.keeper.GetSwapRewardAccrualTime(suite.ctx, poolID)
	suite.True(found)
	suite.Equal(expected, storedTime)
}

func (suite *AccumulateSwapRewardsTests) storedIndexesEqual(poolID string, expected types.RewardIndexes) {
	storedIndexes, found := suite.keeper.GetSwapRewardIndexes(suite.ctx, poolID)
	suite.Equal(found, expected != nil)
	suite.Equal(expected, storedIndexes)
}

func TestAccumulateSwapRewards(t *testing.T) {
	suite.Run(t, new(AccumulateSwapRewardsTests))
}

func (suite *AccumulateSwapRewardsTests) TestStateUpdatedWhenBlockTimeHasIncreased() {
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	suite.storeGlobalSwapIndexes(types.MultiRewardIndexes{
		{
			CollateralType: pool,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "swap",
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
	suite.keeper.SetSwapRewardAccrualTime(suite.ctx, pool, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(newAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(pool, newAccrualTime)
	suite.storedIndexesEqual(pool, types.RewardIndexes{
		{
			CollateralType: "swap",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("3.64"),
		},
	})
}

func (suite *AccumulateSwapRewardsTests) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: pool,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "swap",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalSwapIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetSwapRewardAccrualTime(suite.ctx, pool, previousAccrualTime)

	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(pool, previousAccrualTime)
	expected, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.storedIndexesEqual(pool, expected)
}

func (suite *AccumulateSwapRewardsTests) TestNoAccumulationWhenSourceSharesAreZero() {
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper() // no pools, so no source shares
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: pool,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "swap",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalSwapIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetSwapRewardAccrualTime(suite.ctx, pool, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(7 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(pool, firstAccrualTime)
	expected, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.storedIndexesEqual(pool, expected)
}

func (suite *AccumulateSwapRewardsTests) TestStateAddedWhenStateDoesNotExist() {
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	firstAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.storedTimeEquals(pool, firstAccrualTime)
	suite.storedIndexesEqual(pool, nil)

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(secondAccrualTime)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// After the second accumulation both current block time and indexes should be stored.
	suite.storedTimeEquals(pool, secondAccrualTime)
	suite.storedIndexesEqual(pool, types.RewardIndexes{
		{
			CollateralType: "swap",
			RewardFactor:   d("0.02"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("0.01"),
		},
	})
}

func (suite *AccumulateSwapRewardsTests) TestNoPanicWhenStateDoesNotExist() {
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper()
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(),
	)

	accrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(accrualTime)

	// Accumulate with no swap shares and no rewards per second will result in no increment to the indexes.
	// No increment and no previous indexes stored, results in an updated of nil. Setting this in the state panics.
	// Check there is no panic.
	suite.NotPanics(func() {
		suite.keeper.AccumulateSwapRewards(suite.ctx, period)
	})

	suite.storedTimeEquals(pool, accrualTime)
	suite.storedIndexesEqual(pool, nil)
}

func (suite *AccumulateSwapRewardsTests) TestNoAccumulationWhenBeforeStartTime() {
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: pool,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "swap",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalSwapIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetSwapRewardAccrualTime(suite.ctx, pool, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(10 * time.Second)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		firstAccrualTime.Add(time.Nanosecond), // start time after accrual time
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// The accrual time should be updated, but the indexes unchanged
	suite.storedTimeEquals(pool, firstAccrualTime)
	expectedIndexes, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.storedIndexesEqual(pool, expectedIndexes)
}

func (suite *AccumulateSwapRewardsTests) TestPanicWhenCurrentTimeLessThanPrevious() {
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetSwapRewardAccrualTime(suite.ctx, pool, previousAccrualTime)

	firstAccrualTime := time.Time{}

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Time{}, // start time after accrual time
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.Panics(func() {
		suite.keeper.AccumulateSwapRewards(suite.ctx, period)
	})
}
