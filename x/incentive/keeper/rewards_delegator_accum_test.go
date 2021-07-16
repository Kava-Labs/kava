package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

type AccumulateDelegatorRewardsTests struct {
	unitTester
}

func (suite *AccumulateDelegatorRewardsTests) storedTimeEquals(denom string, expected time.Time) {
	storedTime, found := suite.keeper.GetPreviousDelegatorRewardAccrualTime(suite.ctx, denom)
	suite.True(found)
	suite.Equal(expected, storedTime)
}

func (suite *AccumulateDelegatorRewardsTests) storedIndexesEqual(denom string, expected types.RewardIndexes) {
	storedIndexes, found := suite.keeper.GetDelegatorRewardIndexes(suite.ctx, denom)
	suite.Equal(found, expected != nil)
	suite.Equal(expected, storedIndexes)
}

func TestAccumulateDelegatorRewards(t *testing.T) {
	suite.Run(t, new(AccumulateDelegatorRewardsTests))
}

func (suite *AccumulateDelegatorRewardsTests) TestStateUpdatedWhenBlockTimeHasIncreased() {

	stakingKeeper := newFakeStakingKeeper().addBondedTokens(1e6)
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper, nil)

	suite.storeGlobalDelegatorIndexes(types.MultiRewardIndexes{
		{
			CollateralType: types.GovDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard",
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
	suite.keeper.SetPreviousDelegatorRewardAccrualTime(suite.ctx, types.GovDenom, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(newAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		types.GovDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("hard", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(types.GovDenom, newAccrualTime)
	suite.storedIndexesEqual(types.GovDenom, types.RewardIndexes{
		{
			CollateralType: "hard",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("3.64"),
		},
	})
}

func (suite *AccumulateDelegatorRewardsTests) TestStateUnchangedWhenBlockTimeHasNotIncreased() {

	stakingKeeper := newFakeStakingKeeper().addBondedTokens(1e6)
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper, nil)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: types.GovDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalDelegatorIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousDelegatorRewardAccrualTime(suite.ctx, types.GovDenom, previousAccrualTime)

	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		types.GovDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("hard", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(types.GovDenom, previousAccrualTime)
	expected, f := previousIndexes.Get(types.GovDenom)
	suite.True(f)
	suite.storedIndexesEqual(types.GovDenom, expected)
}

func (suite *AccumulateDelegatorRewardsTests) TestNoAccumulationWhenSourceSharesAreZero() {

	stakingKeeper := newFakeStakingKeeper() // zero total bonded
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper, nil)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: types.GovDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalDelegatorIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousDelegatorRewardAccrualTime(suite.ctx, types.GovDenom, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(7 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		types.GovDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("hard", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(types.GovDenom, firstAccrualTime)
	expected, f := previousIndexes.Get(types.GovDenom)
	suite.True(f)
	suite.storedIndexesEqual(types.GovDenom, expected)
}

func (suite *AccumulateDelegatorRewardsTests) TestStateAddedWhenStateDoesNotExist() {

	stakingKeeper := newFakeStakingKeeper().addBondedTokens(1e6)
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper, nil)

	period := types.NewMultiRewardPeriod(
		true,
		types.GovDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("hard", 2000), c("ukava", 1000)),
	)

	firstAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.storedTimeEquals(types.GovDenom, firstAccrualTime)
	suite.storedIndexesEqual(types.GovDenom, nil)

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(secondAccrualTime)

	suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)

	// After the second accumulation both current block time and indexes should be stored.
	suite.storedTimeEquals(types.GovDenom, secondAccrualTime)
	suite.storedIndexesEqual(types.GovDenom, types.RewardIndexes{
		{
			CollateralType: "hard",
			RewardFactor:   d("0.02"),
		},
		{
			CollateralType: "ukava",
			RewardFactor:   d("0.01"),
		},
	})
}

func (suite *AccumulateDelegatorRewardsTests) TestNoPanicWhenStateDoesNotExist() {

	stakingKeeper := newFakeStakingKeeper()
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper, nil)

	period := types.NewMultiRewardPeriod(
		true,
		types.GovDenom,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(),
	)

	accrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(accrualTime)

	// Accumulate with no source shares and no rewards per second will result in no increment to the indexes.
	// No increment and no previous indexes stored, results in an updated of nil. Setting this in the state panics.
	// Check there is no panic.
	suite.NotPanics(func() {
		suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)
	})

	suite.storedTimeEquals(types.GovDenom, accrualTime)
	suite.storedIndexesEqual(types.GovDenom, nil)
}

func (suite *AccumulateDelegatorRewardsTests) TestNoAccumulationWhenBeforeStartTime() {

	stakingKeeper := newFakeStakingKeeper().addBondedTokens(1e6)
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper, nil)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: types.GovDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.storeGlobalDelegatorIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousDelegatorRewardAccrualTime(suite.ctx, types.GovDenom, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(10 * time.Second)

	period := types.NewMultiRewardPeriod(
		true,
		types.GovDenom,
		firstAccrualTime.Add(time.Nanosecond), // start time after accrual time
		distantFuture,
		cs(c("hard", 2000), c("ukava", 1000)),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)

	// The accrual time should be updated, but the indexes unchanged
	suite.storedTimeEquals(types.GovDenom, firstAccrualTime)
	expectedIndexes, f := previousIndexes.Get(types.GovDenom)
	suite.True(f)
	suite.storedIndexesEqual(types.GovDenom, expectedIndexes)
}

func (suite *AccumulateDelegatorRewardsTests) TestPanicWhenCurrentTimeLessThanPrevious() {

	stakingKeeper := newFakeStakingKeeper().addBondedTokens(1e6)
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, nil, nil, nil, stakingKeeper, nil)

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousDelegatorRewardAccrualTime(suite.ctx, types.GovDenom, previousAccrualTime)

	firstAccrualTime := time.Time{}

	period := types.NewMultiRewardPeriod(
		true,
		types.GovDenom,
		time.Time{}, // start time after accrual time
		distantFuture,
		cs(c("hard", 2000), c("ukava", 1000)),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.Panics(func() {
		suite.keeper.AccumulateDelegatorRewards(suite.ctx, period)
	})
}
