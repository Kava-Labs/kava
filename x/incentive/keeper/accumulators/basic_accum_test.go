package accumulators_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	swaptypes "github.com/kava-labs/kava/x/swap/types"
)

type BasicAccumulatorTestSuite struct {
	testutil.IntegrationTester

	keeper    testutil.TestKeeper
	userAddrs []sdk.AccAddress
	valAddrs  []sdk.ValAddress

	pool string
}

func TestBasicAccumulatorTestSuite(t *testing.T) {
	suite.Run(t, new(BasicAccumulatorTestSuite))
}

func (suite *BasicAccumulatorTestSuite) SetupTest() {
	suite.IntegrationTester.SetupTest()

	suite.keeper = testutil.TestKeeper{
		Keeper: suite.App.GetIncentiveKeeper(),
	}

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	suite.userAddrs = addrs[0:2]
	suite.valAddrs = []sdk.ValAddress{
		sdk.ValAddress(addrs[2]),
		sdk.ValAddress(addrs[3]),
	}

	poolDenomA := "btc"
	poolDenomB := "usdx"

	// Setup app with test state
	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleAccount(addrs[0], cs(
			c("ukava", 1e12),
			c(poolDenomA, 1e12),
			c(poolDenomB, 1e12),
		)).
		WithSimpleAccount(addrs[1], cs(c("ukava", 1e12))).
		WithSimpleAccount(addrs[2], cs(c("ukava", 1e12))).
		WithSimpleAccount(addrs[3], cs(c("ukava", 1e12)))

	incentiveBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.GenesisTime).
		WithSimpleRewardPeriod(types.CLAIM_TYPE_EARN, "bkava", cs())

	savingsBuilder := testutil.NewSavingsGenesisBuilder().
		WithSupportedDenoms("bkava")

	earnBuilder := testutil.NewEarnGenesisBuilder().
		WithAllowedVaults(earntypes.AllowedVault{
			Denom:             "bkava",
			Strategies:        earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
			IsPrivateVault:    false,
			AllowedDepositors: nil,
		})

	stakingBuilder := testutil.NewStakingGenesisBuilder()

	kavamintBuilder := testutil.NewKavamintGenesisBuilder().
		WithStakingRewardsApy(sdk.MustNewDecFromStr("0.2")).
		WithPreviousBlockTime(suite.GenesisTime)

	suite.StartChainWithBuilders(
		authBuilder,
		incentiveBuilder,
		savingsBuilder,
		earnBuilder,
		stakingBuilder,
		kavamintBuilder,
	)

	suite.pool = swaptypes.PoolID(poolDenomA, poolDenomB)

	swapKeeper := suite.App.GetSwapKeeper()
	swapKeeper.SetParams(suite.Ctx, swaptypes.NewParams(
		swaptypes.NewAllowedPools(
			swaptypes.NewAllowedPool(poolDenomA, poolDenomB),
		),
		sdk.ZeroDec(),
	))

}

func TestAccumulateSwapRewards(t *testing.T) {
	suite.Run(t, new(BasicAccumulatorTestSuite))
}

func (suite *BasicAccumulatorTestSuite) TestStateUpdatedWhenBlockTimeHasIncreased() {
	pool := "btc:usdx"

	err := suite.DeliverSwapMsgDeposit(suite.userAddrs[0], c("btc", 1e6), c("usdx", 1e6), d("1.0"))
	suite.Require().NoError(err)

	suite.keeper.StoreGlobalIndexes(
		suite.Ctx,
		types.CLAIM_TYPE_SWAP,
		types.MultiRewardIndexes{
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
		},
	)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_SWAP, pool, previousAccrualTime)

	newAccrualTime := previousAccrualTime.Add(1 * time.Hour)
	suite.Ctx = suite.Ctx.WithBlockTime(newAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	err = suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
	suite.Require().NoError(err)

	// check time and factors

	suite.StoredTimeEquals(types.CLAIM_TYPE_SWAP, pool, newAccrualTime)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_SWAP, pool, types.RewardIndexes{
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

func (suite *BasicAccumulatorTestSuite) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	pool := "btc:usdx"

	err := suite.DeliverSwapMsgDeposit(suite.userAddrs[0], c("btc", 1e6), c("usdx", 1e6), d("1.0"))
	suite.Require().NoError(err)

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
	suite.keeper.StoreGlobalIndexes(
		suite.Ctx,
		types.CLAIM_TYPE_SWAP,
		previousIndexes,
	)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_SWAP, pool, previousAccrualTime)

	suite.Ctx = suite.Ctx.WithBlockTime(previousAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	err = suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
	suite.Require().NoError(err)

	// check time and factors

	suite.StoredTimeEquals(types.CLAIM_TYPE_SWAP, pool, previousAccrualTime)
	expected, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_SWAP, pool, expected)
}

func (suite *BasicAccumulatorTestSuite) TestNoAccumulationWhenSourceSharesAreZero() {
	pool := "btc:usdx"

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
	suite.keeper.StoreGlobalIndexes(
		suite.Ctx,
		types.CLAIM_TYPE_SWAP, previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_SWAP, pool, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(7 * time.Second)
	suite.Ctx = suite.Ctx.WithBlockTime(firstAccrualTime)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	err := suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
	suite.Require().NoError(err)

	// check time and factors

	suite.StoredTimeEquals(types.CLAIM_TYPE_SWAP, pool, firstAccrualTime)
	expected, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_SWAP, pool, expected)
}

func (suite *BasicAccumulatorTestSuite) TestStateAddedWhenStateDoesNotExist() {
	pool := "btc:usdx"

	err := suite.DeliverSwapMsgDeposit(suite.userAddrs[0], c("btc", 1e6), c("usdx", 1e6), d("1.0"))
	suite.Require().NoError(err)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	firstAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.Ctx = suite.Ctx.WithBlockTime(firstAccrualTime)

	err = suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
	suite.Require().NoError(err)

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.StoredTimeEquals(types.CLAIM_TYPE_SWAP, pool, firstAccrualTime)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_SWAP, pool, nil)

	secondAccrualTime := firstAccrualTime.Add(10 * time.Second)
	suite.Ctx = suite.Ctx.WithBlockTime(secondAccrualTime)

	err = suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
	suite.Require().NoError(err)

	// After the second accumulation both current block time and indexes should be stored.
	suite.StoredTimeEquals(types.CLAIM_TYPE_SWAP, pool, secondAccrualTime)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_SWAP, pool, types.RewardIndexes{
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

func (suite *BasicAccumulatorTestSuite) TestNoPanicWhenStateDoesNotExist() {
	pool := "btc:usdx"

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(),
	)

	accrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.Ctx = suite.Ctx.WithBlockTime(accrualTime)

	// Accumulate with no swap shares and no rewards per second will result in no increment to the indexes.
	// No increment and no previous indexes stored, results in an updated of nil. Setting this in the state panics.
	// Check there is no panic.
	suite.NotPanics(func() {
		err := suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
		suite.Require().NoError(err)
	})

	suite.StoredTimeEquals(types.CLAIM_TYPE_SWAP, pool, accrualTime)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_SWAP, pool, nil)
}

func (suite *BasicAccumulatorTestSuite) TestNoAccumulationWhenBeforeStartTime() {
	pool := "btc:usdx"

	err := suite.DeliverSwapMsgDeposit(suite.userAddrs[0], c("btc", 1e6), c("usdx", 1e6), d("1.0"))
	suite.Require().NoError(err)

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
	suite.keeper.StoreGlobalIndexes(
		suite.Ctx,
		types.CLAIM_TYPE_SWAP, previousIndexes)
	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_SWAP, pool, previousAccrualTime)

	firstAccrualTime := previousAccrualTime.Add(10 * time.Second)

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		firstAccrualTime.Add(time.Nanosecond), // start time after accrual time
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	suite.Ctx = suite.Ctx.WithBlockTime(firstAccrualTime)

	err = suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
	suite.Require().NoError(err)

	// The accrual time should be updated, but the indexes unchanged
	suite.StoredTimeEquals(types.CLAIM_TYPE_SWAP, pool, firstAccrualTime)
	expectedIndexes, f := previousIndexes.Get(pool)
	suite.True(f)
	suite.StoredIndexesEqual(types.CLAIM_TYPE_SWAP, pool, expectedIndexes)
}

func (suite *BasicAccumulatorTestSuite) TestPanicWhenCurrentTimeLessThanPrevious() {
	pool := "btc:usdx"

	err := suite.DeliverSwapMsgDeposit(suite.userAddrs[0], c("btc", 1e6), c("usdx", 1e6), d("1.0"))
	suite.Require().NoError(err)

	previousAccrualTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	suite.keeper.Store.SetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_SWAP, pool, previousAccrualTime)

	firstAccrualTime := time.Time{}

	period := types.NewMultiRewardPeriod(
		true,
		pool,
		time.Time{}, // start time after accrual time
		distantFuture,
		cs(c("swap", 2000), c("ukava", 1000)),
	)

	suite.Ctx = suite.Ctx.WithBlockTime(firstAccrualTime)

	suite.Panics(func() {
		suite.keeper.AccumulateRewards(suite.Ctx, types.CLAIM_TYPE_SWAP, period)
	})
}
