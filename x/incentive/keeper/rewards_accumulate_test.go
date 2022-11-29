package keeper_test

import (
	"testing"
	"time"

	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
)

type AccumulateTestSuite struct {
	unitTester
}

func TestAccumulateTestSuite(t *testing.T) {
	suite.Run(t, new(AccumulateTestSuite))
}

func (suite *AccumulateTestSuite) storedTimeEquals(
	claimType types.ClaimType,
	poolID string,
	expected time.Time,
) {
	storedTime, found := suite.keeper.Store.GetRewardAccrualTime(suite.ctx, claimType, poolID)
	suite.True(found)
	suite.Equal(expected, storedTime)
}

func (suite *AccumulateTestSuite) storedIndexesEquals(
	claimType types.ClaimType,
	poolID string,
	expected types.RewardIndexes,
) {
	storedIndexes, found := suite.keeper.Store.GetRewardIndexesOfClaimType(suite.ctx, claimType, poolID)
	suite.Equal(found, expected != nil)
	if found {
		suite.Equal(expected, storedIndexes)
	} else {
		suite.Empty(storedIndexes)
	}
}

func (suite *AccumulateTestSuite) TestStateUpdatedWhenBlockTimeHasIncreased() {
	claimType := types.CLAIM_TYPE_SWAP
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper, nil, nil, nil)

	suite.storeGlobalIndexes(claimType, types.MultiRewardIndexes{
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
	suite.keeper.Store.SetRewardAccrualTime(suite.ctx, claimType, pool, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(newAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateRewards(suite.ctx, claimType, period)

	// check time and factors

	suite.storedTimeEquals(claimType, pool, newAccrualTime)
	suite.storedIndexesEquals(claimType, pool, types.RewardIndexes{
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

func (suite *AccumulateTestSuite) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	claimType := types.CLAIM_TYPE_SWAP
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper, nil, nil, nil)

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
	suite.storeGlobalIndexes(claimType, previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.ctx, claimType, pool, previousAccrualTime)

	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateRewards(suite.ctx, claimType, period)

	// check time and factors

	suite.storedTimeEquals(claimType, pool, previousAccrualTime)
	expected, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.storedIndexesEquals(claimType, pool, expected)
}

func (suite *AccumulateTestSuite) TestNoAccumulationWhenSourceSharesAreZero() {
	claimType := types.CLAIM_TYPE_SWAP
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper() // no pools, so no source shares
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper, nil, nil, nil)

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
	suite.storeGlobalIndexes(claimType, previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.ctx, claimType, pool, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(7 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateRewards(suite.ctx, claimType, period)

	// check time and factors

	suite.storedTimeEquals(claimType, pool, firstAccrualTime)
	expected, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.storedIndexesEquals(claimType, pool, expected)
}

func (suite *AccumulateTestSuite) TestStateAddedWhenStateDoesNotExist() {
	claimType := types.CLAIM_TYPE_SWAP
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper, nil, nil, nil)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	firstAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateRewards(suite.ctx, claimType, period)

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.storedTimeEquals(claimType, pool, firstAccrualTime)
	suite.storedIndexesEquals(claimType, pool, nil)

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(secondAccrualTime)

	suite.keeper.AccumulateRewards(suite.ctx, claimType, period)

	// After the second accumulation both current block time and indexes should be stored.
	suite.storedTimeEquals(claimType, pool, secondAccrualTime)
	suite.storedIndexesEquals(claimType, pool, types.RewardIndexes{
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

func (suite *AccumulateTestSuite) TestNoPanicWhenStateDoesNotExist() {
	claimType := types.CLAIM_TYPE_SWAP
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper()
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper, nil, nil, nil)

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
		suite.keeper.AccumulateRewards(suite.ctx, claimType, period)
	})

	suite.storedTimeEquals(claimType, pool, accrualTime)
	suite.storedIndexesEquals(claimType, pool, nil)
}

func (suite *AccumulateTestSuite) TestNoAccumulationWhenBeforeStartTime() {
	claimType := types.CLAIM_TYPE_SWAP
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper, nil, nil, nil)

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
	suite.storeGlobalIndexes(claimType, previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.ctx, claimType, pool, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(10 * time.Second)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		firstAccrualTime.Add(time.Nanosecond), // start time after accrual time
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateRewards(suite.ctx, claimType, period)

	// The accrual time should be updated, but the indexes unchanged
	suite.storedTimeEquals(claimType, pool, firstAccrualTime)
	expectedIndexes, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.storedIndexesEquals(claimType, pool, expectedIndexes)
}

func (suite *AccumulateTestSuite) TestPanicWhenCurrentTimeLessThanPrevious() {
	claimType := types.CLAIM_TYPE_SWAP
	pool := "btc:usdx"

	swapKeeper := newFakeSwapKeeper().addPool(pool, i(1e6))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper, nil, nil, nil)

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.ctx, claimType, pool, previousAccrualTime)

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
		suite.keeper.AccumulateRewards(suite.ctx, claimType, period)
	})
}
