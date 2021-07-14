package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

type AccumulateSwapRewardsTests struct {
	unitTester
}

func (suite *AccumulateSwapRewardsTests) checkStoredTimeEquals(poolID string, expected time.Time) {
	storedTime, found := suite.keeper.GetSwapRewardAccrualTime(suite.ctx, poolID)
	suite.True(found)
	suite.Equal(expected, storedTime)
}

func (suite *AccumulateSwapRewardsTests) checkStoredIndexesEqual(poolID string, expected types.RewardIndexes) {
	storedIndexes, found := suite.keeper.GetSwapRewardIndexes(suite.ctx, poolID)
	suite.True(found)
	suite.Equal(expected, storedIndexes)
}

func TestAccumulateSwapRewards(t *testing.T) {
	suite.Run(t, new(AccumulateSwapRewardsTests))
}

func (suite *AccumulateSwapRewardsTests) TestStateUpdatedWhenBlockTimeHasIncreased() {
	swapKeeper := &fakeSwapKeeper{i(1e6)}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	pool := "btc/usdx"
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

	suite.checkStoredTimeEquals(pool, newAccrualTime)

	expectedIndexes := types.RewardIndexes{
		{
			CollateralType: "swap",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("3.64"),
		},
	}
	suite.checkStoredIndexesEqual(pool, expectedIndexes)
}

func (suite *AccumulateSwapRewardsTests) TestLimitsOfAccumulationPrecision() {
	swapKeeper := &fakeSwapKeeper{i(1e17)} // approximate shares in a $1B pool of 10^8 precision ~$1 asset
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	pool := "btc/usdx"
	suite.storeGlobalSwapIndexes(types.MultiRewardIndexes{
		{
			CollateralType: pool,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "swap",
					RewardFactor:   d("0.0"),
				},
			},
		},
	})
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetSwapRewardAccrualTime(suite.ctx, pool, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Second) // 1 second is the smallest increment accrual happens over
	suite.ctx = suite.ctx.WithBlockTime(newAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0),
		distantFuture,
		cs(c("swap", 1)), // single unit of any denom is the smallest reward amount
	)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// check time and factors

	suite.checkStoredTimeEquals(pool, newAccrualTime)

	expectedIndexes := types.RewardIndexes{
		{
			CollateralType: "swap",
			// smallest reward amount over smallest accumulation duration does not go past 10^-18 decimal precision
			RewardFactor: d("0.000000000000000010"),
		},
	}
	suite.checkStoredIndexesEqual(pool, expectedIndexes)
}

func (suite *AccumulateSwapRewardsTests) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	swapKeeper := &fakeSwapKeeper{i(1e6)}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	pool := "btc/usdx"
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

	suite.checkStoredTimeEquals(pool, previousAccrualTime)

	expectedIndexes := types.RewardIndexes{
		{
			CollateralType: "swap",
			RewardFactor:   d("0.02"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("0.04"),
		},
	}
	suite.checkStoredIndexesEqual(pool, expectedIndexes)
}

func (suite *AccumulateSwapRewardsTests) TestStateAddedWhenStateDoesNotExist() {
	swapKeeper := &fakeSwapKeeper{i(1e6)}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	pool := "btc/usdx"

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
	// This indexes will be zero as no time has passed since the previous block because it didn't exist.
	suite.checkStoredTimeEquals(pool, firstAccrualTime)

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(secondAccrualTime)

	suite.keeper.AccumulateSwapRewards(suite.ctx, period)

	// After the second accumulation both current block time and indexes should be stored.
	suite.checkStoredTimeEquals(pool, secondAccrualTime)

	expectedIndexes := types.RewardIndexes{
		{
			CollateralType: "swap",
			RewardFactor:   d("0.02"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("0.01"),
		},
	}
	suite.checkStoredIndexesEqual(pool, expectedIndexes)
}
func (suite *AccumulateSwapRewardsTests) TestNoPanicWhenStateDoesNotExist() {
	swapKeeper := &fakeSwapKeeper{i(0)}
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, nil, swapKeeper)

	pool := "btc/usdx"

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

	suite.checkStoredTimeEquals(pool, accrualTime)
}

type fakeSwapKeeper struct {
	poolShares sdk.Int
}

func (k fakeSwapKeeper) GetPoolShares(ctx sdk.Context, poolID string) (sdk.Int, bool) {
	return k.poolShares, true
}
func (k fakeSwapKeeper) GetDepositorSharesAmount(ctx sdk.Context, depositor sdk.AccAddress, poolID string) (sdk.Int, bool) {
	// This is just to implement the swap keeper interface.
	return sdk.Int{}, false
}

// note: amino panics when encoding times ≥ the start of year 10000.
var distantFuture = time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)
