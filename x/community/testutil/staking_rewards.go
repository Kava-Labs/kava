package testutil

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// StakingRewardsTestSuite tests staking rewards per second logic
type stakingRewardsTestSuite struct {
	suite.Suite

	App    app.TestApp
	Keeper keeper.Keeper

	testFunc testFunc
}

func NewStakingRewardsTestSuite(tf testFunc) *stakingRewardsTestSuite {
	suite := &stakingRewardsTestSuite{}
	suite.testFunc = tf
	return suite
}

// The default state used by each test
func (suite *stakingRewardsTestSuite) SetupTest() {
	app.SetSDKConfig()

	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates()

	suite.App = tApp
	suite.Keeper = suite.App.GetCommunityKeeper()
}

func (suite *stakingRewardsTestSuite) TestStakingRewards() {
	testCases := []struct {
		// name of subtest
		name string

		// block time of first block
		periodStart time.Time
		// block time of last block
		periodEnd time.Time

		// block time n will be periodStart + rand(range_min...range_max)*(n-1) up to periodEnd
		blockTimeRangeMin float64
		blockTimeRangeMax float64

		// rewards per second to set in state
		rewardsPerSecond sdkmath.LegacyDec

		// the amount of ukava to mint and transfer to the community pool
		// to use to pay for rewards
		communityPoolFunds sdkmath.Int

		// how many total rewards are expected to be accumulated in ukava
		expectedRewardsTotal sdkmath.Int
	}{
		// ** These take a long time to run **
		//{
		//	name:                 "one year with 0.5 to 1 second block times",
		//	periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		//	periodEnd:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		//	blockTimeRangeMin:    0.5,
		//	blockTimeRangeMax:    1,
		//	rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("1585489.599188229325215626"),
		//	expectedRewardsTotal: sdkmath.NewInt(49999999999999), // 50 million KAVA per year
		//},
		//{
		//	name:                 "one year with 5.5 to 6.5 second blocktimes",
		//	periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		//	periodEnd:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		//	blockTimeRangeMin:    5.5,
		//	blockTimeRangeMax:    6.5,
		//	rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("1585489.599188229325215626"), // 50 million kava per year
		//	communityPoolFunds:   sdkmath.NewInt(50000000000000),
		//	expectedRewardsTotal: sdkmath.NewInt(49999999999999), // truncation results in 1 ukava error
		//},
		//
		//
		//  One Day of blocks with different block time variations
		//
		//
		{
			name:                 "one day with sub-second block times and 50 million KAVA per year",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    0.1,
			blockTimeRangeMax:    1,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("1585489.599188229325215626"), // 50 million kava per year
			communityPoolFunds:   sdkmath.NewInt(200000000000),
			expectedRewardsTotal: sdkmath.NewInt(136986301369), // 50 million / 365 days  - 1 ukava

		},
		{
			name:                 "one day with 5.5 to 6.5 second block times and 50 million KAVA per year",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    5.5,
			blockTimeRangeMax:    6.5,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("1585489.599188229325215626"), // 50 million kava per year
			communityPoolFunds:   sdkmath.NewInt(200000000000),
			expectedRewardsTotal: sdkmath.NewInt(136986301369), // 50 million / 365 days  - 1 ukava
		},
		//
		//
		// Total time span under 1 second
		//
		//
		{
			name:                 "single 6.9 second time span and 25 million KAVA per year",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 1, 0, 0, 6, 900000000, time.UTC),
			blockTimeRangeMin:    10, // forces only two blocks -- one time span
			blockTimeRangeMax:    10,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("792744.799594114662607813"), // 25 million kava per year
			communityPoolFunds:   sdkmath.NewInt(10000000),
			expectedRewardsTotal: sdkmath.NewInt(5469939), // per second rate * 6.9
		},
		{
			name:                 "multiple blocks across sub-second time span nd 10 million KAVA per year",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 1, 0, 0, 0, 800000000, time.UTC),
			blockTimeRangeMin:    0.1, // multiple blocks in a sub-second time span
			blockTimeRangeMax:    0.2,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("317097.919837645865043125"), // 10 million kava per year
			communityPoolFunds:   sdkmath.NewInt(300000),
			expectedRewardsTotal: sdkmath.NewInt(253678), // per second rate * 0.8
		},
		//
		//
		// Variations of community pool balance
		//
		//
		{
			name:                 "community pool exact funds -- should spend community to zero and not panic",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    5.5,
			blockTimeRangeMax:    6.2,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("317097.919837645865043125"), // 10 million kava per year
			communityPoolFunds:   sdkmath.NewInt(27397260273),
			expectedRewardsTotal: sdkmath.NewInt(27397260273),
		},
		{
			name:                 "community pool under funded -- should spend community pool to down to zero and not panic",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    5.5,
			blockTimeRangeMax:    6.5,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("1585489.599188229325215626"), // 25 million kava per year
			communityPoolFunds:   sdkmath.NewInt(100000000000),                                  // under funded
			expectedRewardsTotal: sdkmath.NewInt(100000000000),                                  // rewards max is the community pool balance
		},
		{
			name:                 "community pool no funds -- should pay zero rewards and not panic",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    5.5,
			blockTimeRangeMax:    6.5,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("792744.799594114662607813"), // 25 million kava per year
			communityPoolFunds:   sdkmath.NewInt(0),
			expectedRewardsTotal: sdkmath.NewInt(0),
		},
		//
		//
		// Disabled
		//
		//
		{
			name:                 "zero rewards per second results in zero rewards paid",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    5.5,
			blockTimeRangeMax:    6.5,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("0.000000000000000000"), // 25 million kava per year
			communityPoolFunds:   sdkmath.NewInt(100000000000000),
			expectedRewardsTotal: sdkmath.NewInt(0),
		},
		//
		//
		// Test underlying calculations are safe and overflow/underflow bounds are reasonable
		//
		//
		{
			name:                 "does not overflow with extremely large per second value and extremely large single block durations",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2033, 1, 1, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    315619200,                                                                         // a single 10 year long block in seconds (w/ 3 leap years)
			blockTimeRangeMax:    315619200,                                                                         // a single 10 year long block in seconds (w/ 3 leap years)
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("100000000000000000000000000.000000000000000000"), // 100 million kava per second in 18 decimal form
			communityPoolFunds:   newIntFromString("40000000000000000000000000000000000"),
			expectedRewardsTotal: newIntFromString("31561920000000000000000000000000000"), // 10 years worth of rewards (with three leap years)
		},
		{
			name:                 "able to accumulate decimal ukava units across blocks",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    5.5,
			blockTimeRangeMax:    6.5,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("0.100000000000000000"), // blocks are not long enough to accumulate a single ukava with this rate
			communityPoolFunds:   sdkmath.NewInt(10000),
			expectedRewardsTotal: sdkmath.NewInt(8640),
		},
		{
			name:                 "down to 1 ukava per year can be accumulated -- we are safe from underflow at reasonably small values",
			periodStart:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			periodEnd:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			blockTimeRangeMin:    60, // large block times speed up this test case
			blockTimeRangeMax:    120,
			rewardsPerSecond:     sdkmath.LegacyMustNewDecFromStr("0.000000031709791984"),
			communityPoolFunds:   sdkmath.NewInt(1),
			expectedRewardsTotal: sdkmath.NewInt(1),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// keepers
			keeper := suite.Keeper
			accountKeeper := suite.App.GetAccountKeeper()
			bankKeeper := suite.App.GetBankKeeper()

			// initial context at height 1
			height := int64(1)
			blockTime := tc.periodStart
			ctx := suite.App.NewContext(true, tmproto.Header{Height: height, Time: blockTime})

			// ensure community pool balance matches the test expectations
			poolAcc := accountKeeper.GetModuleAccount(ctx, types.ModuleName)
			// community pool balance should start at zero
			suite.Require().True(bankKeeper.GetBalance(ctx, poolAcc.GetAddress(), "ukava").Amount.IsZero(), "expected community pool to start with zero coins in test genesis")
			// fund withexact amount from test case
			suite.App.FundAccount(ctx, poolAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin("ukava", tc.communityPoolFunds)))

			// get starting balance of fee collector to substract later in case this is non-zero in genesis
			feeCollectorAcc := accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
			initialFeeCollectorBalance := bankKeeper.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava").Amount

			// set rewards per second in state
			params, found := keeper.GetParams(ctx)
			suite.Require().True(found)
			params.StakingRewardsPerSecond = tc.rewardsPerSecond
			keeper.SetParams(ctx, params)

			stakingRewardEvents := sdk.Events{}

			for {
				// run community begin blocker logic
				suite.testFunc(ctx, keeper)

				// accumulate event rewards from events
				stakingRewardEvents = append(stakingRewardEvents, filterStakingRewardEvents(ctx.EventManager().Events())...)

				// exit loop if we are at last block
				if blockTime.Equal(tc.periodEnd) {
					break
				}

				// create random block duration in nanoseconds
				randomBlockDurationInSeconds := tc.blockTimeRangeMin + rand.Float64()*(tc.blockTimeRangeMax-tc.blockTimeRangeMin)
				nextBlockDuration := time.Duration(randomBlockDurationInSeconds * math.Pow10(9))

				// move to next block by incrementing height, adding random duration, and settings new context
				height++
				blockTime = blockTime.Add(nextBlockDuration)
				// set last block to exact end of period if we go past
				if blockTime.After(tc.periodEnd) {
					blockTime = tc.periodEnd
				}
				ctx = suite.App.NewContext(true, tmproto.Header{Height: height, Time: blockTime})
			}

			endingFeeCollectorBalance := bankKeeper.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava").Amount
			feeCollectorBalanceAdded := endingFeeCollectorBalance.Sub(initialFeeCollectorBalance)

			// assert fee pool was payed the correct rewards
			suite.Equal(tc.expectedRewardsTotal.String(), feeCollectorBalanceAdded.String(), "expected fee collector balance to match")

			if tc.expectedRewardsTotal.IsZero() {
				suite.Equal(0, len(stakingRewardEvents), "expected no events to be emitted")
			} else {
				// we add up all reward coin events
				eventCoins := getRewardCoinsFromEvents(stakingRewardEvents)

				// assert events emitted match expected rewards
				suite.Equal(
					tc.expectedRewardsTotal.String(),
					eventCoins.AmountOf("ukava").String(),
					"expected event coins to match",
				)
			}

			// assert the community pool deducted the same amount
			expectedCommunityPoolBalance := tc.communityPoolFunds.Sub(tc.expectedRewardsTotal)
			actualCommunityPoolBalance := bankKeeper.GetBalance(ctx, poolAcc.GetAddress(), "ukava").Amount
			suite.Equal(expectedCommunityPoolBalance.String(), actualCommunityPoolBalance.String(), "expected community pool balance to match")
		})
	}

}

func (suite *stakingRewardsTestSuite) TestStakingRewardsDoNotAccumulateWhenPoolIsDrained() {
	app := suite.App
	keeper := suite.Keeper
	accountKeeper := suite.App.GetAccountKeeper()
	bankKeeper := suite.App.GetBankKeeper()

	// first block
	blockTime := time.Now()
	ctx := app.NewContext(true, tmproto.Header{Height: 1, Time: blockTime})

	poolAcc := accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	feeCollectorAcc := accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	// set state to pay staking rewards
	params, _ := keeper.GetParams(ctx)
	// we set a decimal amount that ensures after 10 seconds we overspend the community pool
	// with enough truncation error that we would have an ending balance of 20.000001 if it was
	// carried over after the pool run out of funds
	params.StakingRewardsPerSecond = sdkmath.LegacyMustNewDecFromStr("1000000.099999999999999999") // > 1 KAVA per second
	keeper.SetParams(ctx, params)

	// fund community pool account
	app.FundAccount(ctx, poolAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(10000000)))) // 10 KAVA
	initialFeeCollectorBalance := bankKeeper.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava").Amount

	// run first block (no rewards hapeen on first block)
	community.BeginBlocker(ctx, keeper)

	// run second block 10 seconds in future and spend all community pool rewards
	blockTime = blockTime.Add(10 * time.Second)
	ctx = app.NewContext(true, tmproto.Header{Height: 2, Time: blockTime})
	community.BeginBlocker(ctx, keeper)

	// run third block 10 seconds in future which no rewards will be paid
	blockTime = blockTime.Add(10 * time.Second)
	ctx = app.NewContext(true, tmproto.Header{Height: 3, Time: blockTime})
	community.BeginBlocker(ctx, keeper)

	// run fourth block 10 seconds in future which no rewards will be paid
	blockTime = blockTime.Add(10 * time.Second)
	ctx = app.NewContext(true, tmproto.Header{Height: 4, Time: blockTime})
	community.BeginBlocker(ctx, keeper)

	// refund the community pool with 100 KAVA -- plenty of funds
	app.FundAccount(ctx, poolAcc.GetAddress(), sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(100000000)))) // 100 KAVA

	// run fifth block 10 seconds in future which no rewards will be paid
	blockTime = blockTime.Add(10 * time.Second)
	ctx = app.NewContext(true, tmproto.Header{Height: 5, Time: blockTime})
	community.BeginBlocker(ctx, keeper)

	// assert that only 20 total KAVA has been distributed in rewards
	// and blocks where community pool had d
	rewards := bankKeeper.GetBalance(ctx, feeCollectorAcc.GetAddress(), "ukava").Amount.Sub(initialFeeCollectorBalance)
	suite.Require().Equal(sdkmath.NewInt(20000000).String(), rewards.String())
}

func (suite *stakingRewardsTestSuite) TestPanicsOnMissingParameters() {
	suite.SetupTest()

	ctx := suite.App.NewContext(true, tmproto.Header{Height: 1, Time: time.Now()})
	store := ctx.KVStore(suite.App.GetKVStoreKey(types.StoreKey))
	store.Delete(types.ParamsKey)

	suite.PanicsWithValue("invalid state: module parameters not found", func() {
		suite.testFunc(ctx, suite.Keeper)
	})
}

// newIntFromString returns a new sdkmath.Int from a string
func newIntFromString(str string) sdkmath.Int {
	num, ok := sdkmath.NewIntFromString(str)
	if !ok {
		panic(fmt.Sprintf("overflow creating Int from %s", str))
	}
	return num
}

func filterStakingRewardEvents(events sdk.Events) (rewardEvents sdk.Events) {
	for _, event := range events {
		if event.Type == types.EventTypeStakingRewardsPaid {
			rewardEvents = append(rewardEvents, event)
		}
	}

	return
}

func getRewardCoinsFromEvents(events sdk.Events) sdk.Coins {
	coins := sdk.NewCoins()

	for _, event := range events {
		if event.Type == types.EventTypeStakingRewardsPaid {
			rewards, err := sdk.ParseCoinNormalized(string(event.Attributes[0].Value))
			if err != nil {
				panic(err)
			}

			coins = coins.Add(rewards)
		}
	}

	return coins
}
