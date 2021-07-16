package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

type AccumulateUSDXRewardsTests struct {
	usdxRewardsUnitTester
}

func (suite *AccumulateUSDXRewardsTests) storedTimeEquals(cType string, expected time.Time) {
	storedTime, found := suite.keeper.GetPreviousUSDXMintingAccrualTime(suite.ctx, cType)
	suite.True(found)
	suite.Equal(expected, storedTime)
}

func (suite *AccumulateUSDXRewardsTests) storedIndexesEqual(cType string, expected sdk.Dec) {
	storedIndexes, found := suite.keeper.GetUSDXMintingRewardFactor(suite.ctx, cType)
	suite.True(found)
	suite.Equal(expected, storedIndexes)
}

func TestAccumulateUSDXRewards(t *testing.T) {
	suite.Run(t, new(AccumulateUSDXRewardsTests))
}

func (suite *AccumulateUSDXRewardsTests) TestStateUpdatedWhenBlockTimeHasIncreased() {
	cType := "bnb-a"

	cdpKeeper := newFakeCDPKeeper().addTotalPrincipal(i(1e6)).addInterestFactor(d("1"))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, cdpKeeper, nil, nil, nil, nil)

	suite.storeGlobalUSDXIndexes(types.RewardIndexes{
		{
			CollateralType: cType,
			RewardFactor:   d("0.04"),
		},
	})
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, cType, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(newAccrualTime)

	period := types.NewRewardPeriod(
		true,
		cType,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		c("ukava", 1000),
	)

	suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(cType, newAccrualTime)
	suite.storedIndexesEqual(cType, d("3.64"))
}

func (suite *AccumulateUSDXRewardsTests) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	cType := "bnb-a"

	cdpKeeper := newFakeCDPKeeper().addTotalPrincipal(i(1e6)).addInterestFactor(d("1"))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, cdpKeeper, nil, nil, nil, nil)

	previousIndexes := types.RewardIndexes{
		{
			CollateralType: cType,
			RewardFactor:   d("0.04"),
		},
	}
	suite.storeGlobalUSDXIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, cType, previousAccrualTime)

	suite.ctx = suite.ctx.WithBlockTime(previousAccrualTime)

	period := types.NewRewardPeriod(
		true,
		cType,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		c("ukava", 2000),
	)

	suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(cType, previousAccrualTime)
	expected, f := previousIndexes.Get(cType)
	suite.True(f)
	suite.storedIndexesEqual(cType, expected)
}

func (suite *AccumulateUSDXRewardsTests) TestNoAccumulationWhenSourceSharesAreZero() {
	cType := "bnb-a"

	cdpKeeper := newFakeCDPKeeper() // zero total borrows
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, cdpKeeper, nil, nil, nil, nil)

	previousIndexes := types.RewardIndexes{
		{
			CollateralType: cType,
			RewardFactor:   d("0.04"),
		},
	}
	suite.storeGlobalUSDXIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, cType, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(7 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	period := types.NewRewardPeriod(
		true,
		cType,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		c("ukava", 1000),
	)

	suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, period)

	// check time and factors

	suite.storedTimeEquals(cType, firstAccrualTime)
	expected, f := previousIndexes.Get(cType)
	suite.True(f)
	suite.storedIndexesEqual(cType, expected)
}

func (suite *AccumulateUSDXRewardsTests) TestStateAddedWhenStateDoesNotExist() {
	cType := "bnb-a"

	cdpKeeper := newFakeCDPKeeper().addTotalPrincipal(i(1e6)).addInterestFactor(d("1"))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, cdpKeeper, nil, nil, nil, nil)

	period := types.NewRewardPeriod(
		true,
		cType,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		c("ukava", 1000),
	)

	firstAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, period)

	// After the first accumulation the current block time should be stored and the factor will be zero.
	suite.storedTimeEquals(cType, firstAccrualTime)
	suite.storedIndexesEqual(cType, sdk.ZeroDec())

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(secondAccrualTime)

	suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, period)

	// After the second accumulation both current block time and indexes should be stored.
	suite.storedTimeEquals(cType, secondAccrualTime)
	suite.storedIndexesEqual(cType, d("0.01"))
}

func (suite *AccumulateUSDXRewardsTests) TestNoAccumulationWhenBeforeStartTime() {
	cType := "bnb-a"

	cdpKeeper := newFakeCDPKeeper().addTotalPrincipal(i(1e6)).addInterestFactor(d("1"))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, cdpKeeper, nil, nil, nil, nil)

	previousIndexes := types.RewardIndexes{
		{
			CollateralType: cType,
			RewardFactor:   d("0.04"),
		},
	}
	suite.storeGlobalUSDXIndexes(previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, cType, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(10 * time.Second)

	period := types.NewRewardPeriod(
		true,
		cType,
		firstAccrualTime.Add(time.Nanosecond), // start time after accrual time
		distantFuture,
		c("ukava", 1000),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, period)

	// The accrual time should be updated, but the indexes unchanged
	suite.storedTimeEquals(cType, firstAccrualTime)
	expected, f := previousIndexes.Get(cType)
	suite.True(f)
	suite.storedIndexesEqual(cType, expected)
}

func (suite *AccumulateUSDXRewardsTests) TestPanicWhenCurrentTimeLessThanPrevious() {
	cType := "bnb-a"

	cdpKeeper := newFakeCDPKeeper().addTotalPrincipal(i(1e6)).addInterestFactor(d("1"))
	suite.keeper = suite.NewKeeper(&fakeParamSubspace{}, nil, cdpKeeper, nil, nil, nil, nil)

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.SetPreviousUSDXMintingAccrualTime(suite.ctx, cType, previousAccrualTime)

	firstAccrualTime := time.Time{}

	period := types.NewRewardPeriod(
		true,
		cType,
		time.Time{}, // start time after accrual time
		distantFuture,
		c("ukava", 1000),
	)

	suite.ctx = suite.ctx.WithBlockTime(firstAccrualTime)

	suite.Panics(func() {
		suite.keeper.AccumulateUSDXMintingRewards(suite.ctx, period)
	})
}
