package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *KeeperTestSuite) TestAccumulateRewards() {
	type args struct {
		ctype                 string
		rewardsPerSecond      sdk.Coin
		initialTime           time.Time
		initialTotalPrincipal sdk.Coin
		timeElapsed           int
		expectedRewardFactor  sdk.Dec
	}
	type test struct {
		name string
		args args
	}
	testCases := []test{
		{
			"7 seconds",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           7,
				expectedRewardFactor:  d("0.000000856478000000"),
			},
		},
		{
			"1 day",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           86400,
				expectedRewardFactor:  d("0.0105713856"),
			},
		},
		{
			"0 seconds",
			args{
				ctype:                 "bnb-a",
				rewardsPerSecond:      c("ukava", 122354),
				initialTime:           time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialTotalPrincipal: c("usdx", 1000000000000),
				timeElapsed:           0,
				expectedRewardFactor:  d("0.0"),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithCDPGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			cdpKeeper.SetTotalPrincipal(suite.ctx, tc.args.ctype, cdptypes.DefaultStableDenom, tc.args.initialTotalPrincipal.Amount)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousAccrualTime(suite.ctx, tc.args.ctype, tc.args.initialTime)
			suite.keeper.SetRewardFactor(suite.ctx, tc.args.ctype, sdk.ZeroDec())

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			rewardPeriod, found := suite.keeper.GetRewardPeriod(suite.ctx, tc.args.ctype)
			suite.Require().True(found)
			err := suite.keeper.AccumulateRewards(suite.ctx, rewardPeriod)
			suite.Require().NoError(err)

			rewardFactor, found := suite.keeper.GetRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)
		})
	}
}

func (suite *KeeperTestSuite) TestSyncRewards() {
	type args struct {
		ctype                string
		rewardsPerSecond     sdk.Coin
		initialTime          time.Time
		initialCollateral    sdk.Coin
		initialPrincipal     sdk.Coin
		blockTimes           []int
		expectedRewardFactor sdk.Dec
		expectedRewards      sdk.Coin
	}
	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"10 blocks",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
				expectedRewardFactor: d("0.001223540000000000"),
				expectedRewards:      c("ukava", 12235400),
			},
		},
		{
			"10 blocks - long block time",
			args{
				ctype:                "bnb-a",
				rewardsPerSecond:     c("ukava", 122354),
				initialTime:          time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialCollateral:    c("bnb", 1000000000000),
				initialPrincipal:     c("usdx", 10000000000),
				blockTimes:           []int{86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400, 86400},
				expectedRewardFactor: d("10.57138560000000000"),
				expectedRewards:      c("ukava", 105713856000),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupWithCDPGenState()
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup incentive state
			params := types.NewParams(
				true,
				types.RewardPeriods{types.NewRewardPeriod(true, tc.args.ctype, tc.args.initialTime, tc.args.initialTime.Add(time.Hour*24*365*4), tc.args.rewardsPerSecond)},
				types.Multipliers{types.NewMultiplier(types.MultiplierName("small"), 1, d("0.25")), types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0"))},
				tc.args.initialTime.Add(time.Hour*24*365*5),
			)
			suite.keeper.SetParams(suite.ctx, params)
			suite.keeper.SetPreviousAccrualTime(suite.ctx, tc.args.ctype, tc.args.initialTime)
			suite.keeper.SetRewardFactor(suite.ctx, tc.args.ctype, sdk.ZeroDec())

			// setup account state
			sk := suite.app.GetSupplyKeeper()
			sk.MintCoins(suite.ctx, cdptypes.ModuleName, sdk.NewCoins(tc.args.initialCollateral))
			sk.SendCoinsFromModuleToAccount(suite.ctx, cdptypes.ModuleName, suite.addrs[0], sdk.NewCoins(tc.args.initialCollateral))

			// setup cdp state
			cdpKeeper := suite.app.GetCDPKeeper()
			err := cdpKeeper.AddCdp(suite.ctx, suite.addrs[0], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)
			suite.Require().NoError(err)

			claim, found := suite.keeper.GetClaim(suite.ctx, suite.addrs[0])
			suite.Require().True(found)
			suite.Require().Equal(sdk.ZeroDec(), claim.RewardIndexes[0].RewardFactor)

			var timeElapsed int
			previousBlockTime := suite.ctx.BlockTime()
			for _, t := range tc.args.blockTimes {
				timeElapsed += t
				updatedBlockTime := previousBlockTime.Add(time.Duration(int(time.Second) * t))
				previousBlockTime = updatedBlockTime
				blockCtx := suite.ctx.WithBlockTime(updatedBlockTime)
				rewardPeriod, found := suite.keeper.GetRewardPeriod(blockCtx, tc.args.ctype)
				suite.Require().True(found)
				err := suite.keeper.AccumulateRewards(blockCtx, rewardPeriod)
				suite.Require().NoError(err)
			}
			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			cdp, found := cdpKeeper.GetCdpByOwnerAndCollateralType(suite.ctx, suite.addrs[0], tc.args.ctype)
			suite.Require().True(found)
			suite.Require().NotPanics(func() {
				suite.keeper.SynchronizeReward(suite.ctx, cdp)
			})

			rewardFactor, found := suite.keeper.GetRewardFactor(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedRewardFactor, rewardFactor)

			claim, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0])
			fmt.Println(claim)
			suite.Require().True(found)
			suite.Require().Equal(tc.args.expectedRewardFactor, claim.RewardIndexes[0].RewardFactor)
			suite.Require().Equal(tc.args.expectedRewards, claim.Reward)
		})
	}

}

func (suite *KeeperTestSuite) SetupWithCDPGenState() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	keeper := tApp.GetIncentiveKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
}

// func (suite *KeeperTestSuite) TestExpireRewardPeriod() {
// 	rp := types.NewRewardPeriod("bnb", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})
// 	suite.keeper.SetRewardPeriod(suite.ctx, rp)
// 	suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
// 	suite.NotPanics(func() {
// 		suite.keeper.HandleRewardPeriodExpiry(suite.ctx, rp)
// 	})
// 	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
// 	suite.True(found)
// }

// func (suite *KeeperTestSuite) TestAddToClaim() {
// 	rp := types.NewRewardPeriod("bnb", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})
// 	suite.keeper.SetRewardPeriod(suite.ctx, rp)
// 	suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
// 	suite.keeper.HandleRewardPeriodExpiry(suite.ctx, rp)
// 	c1 := types.NewClaim(suite.addrs[0], c("ukava", 1000000), "bnb", 1)
// 	suite.keeper.SetClaim(suite.ctx, c1)
// 	suite.NotPanics(func() {
// 		suite.keeper.AddToClaim(suite.ctx, suite.addrs[0], "bnb", 1, c("ukava", 1000000))
// 	})
// 	testC, _ := suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
// 	suite.Equal(c("ukava", 2000000), testC.Reward)

// 	suite.NotPanics(func() {
// 		suite.keeper.AddToClaim(suite.ctx, suite.addrs[0], "xpr", 1, c("ukava", 1000000))
// 	})
// }

// func (suite *KeeperTestSuite) TestCreateRewardPeriod() {
// 	reward := types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)
// 	suite.NotPanics(func() {
// 		suite.keeper.CreateNewRewardPeriod(suite.ctx, reward)
// 	})
// 	_, found := suite.keeper.GetRewardPeriod(suite.ctx, "bnb")
// 	suite.True(found)
// }

// func (suite *KeeperTestSuite) TestCreateAndDeleteRewardsPeriods() {
// 	reward1 := types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)
// 	reward2 := types.NewReward(false, "xrp", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)
// 	reward3 := types.NewReward(false, "btc", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)
// 	// add a reward period to the store for a non-active reward
// 	suite.NotPanics(func() {
// 		suite.keeper.CreateNewRewardPeriod(suite.ctx, reward3)
// 	})
// 	params := types.NewParams(true, types.Rewards{reward1, reward2, reward3})
// 	suite.keeper.SetParams(suite.ctx, params)

// 	suite.NotPanics(func() {
// 		suite.keeper.CreateAndDeleteRewardPeriods(suite.ctx)
// 	})
// 	testCases := []struct {
// 		name        string
// 		arg         string
// 		expectFound bool
// 	}{
// 		{
// 			"active reward period",
// 			"bnb",
// 			true,
// 		},
// 		{
// 			"attempt to add inactive reward period",
// 			"xrp",
// 			false,
// 		},
// 		{
// 			"remove inactive reward period",
// 			"btc",
// 			false,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(tc.name, func() {
// 			_, found := suite.keeper.GetRewardPeriod(suite.ctx, tc.arg)
// 			if tc.expectFound {
// 				suite.True(found)
// 			} else {
// 				suite.False(found)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestApplyRewardsToCdps() {
// 	suite.setupCdpChain() // creates a test app with 3 BNB cdps and usdx incentives for bnb - each reward period is one week

// 	// move the context forward by 100 periods
// 	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 100))
// 	// apply rewards to BNB cdps
// 	suite.NotPanics(func() {
// 		suite.keeper.ApplyRewardsToCdps(suite.ctx)
// 	})
// 	// each cdp should have a claim
// 	claims := types.Claims{}
// 	suite.keeper.IterateClaims(suite.ctx, func(c types.Claim) (stop bool) {
// 		claims = append(claims, c)
// 		return false
// 	})
// 	suite.Equal(3, len(claims))
// 	// there should be no associated claim period, because the reward period has not ended yet
// 	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb-a")
// 	suite.False(found)

// 	// move ctx to the reward period expiry and check that the claim period has been created and the next claim period id has increased
// 	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour * 24 * 7))

// 	suite.NotPanics(func() {
// 		// apply rewards to cdps
// 		suite.keeper.ApplyRewardsToCdps(suite.ctx)
// 		// delete the old reward period amd create a new one
// 		suite.keeper.CreateAndDeleteRewardPeriods(suite.ctx)
// 	})
// 	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb-a")
// 	suite.True(found)
// 	testID := suite.keeper.GetNextClaimPeriodID(suite.ctx, "bnb-a")
// 	suite.Equal(uint64(2), testID)

// 	// move the context forward by 100 periods
// 	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 100))
// 	// run the begin blocker functions
// 	suite.NotPanics(func() {
// 		suite.keeper.DeleteExpiredClaimsAndClaimPeriods(suite.ctx)
// 		suite.keeper.ApplyRewardsToCdps(suite.ctx)
// 		suite.keeper.CreateAndDeleteRewardPeriods(suite.ctx)
// 	})
// 	// each cdp should now have two claims
// 	claims = types.Claims{}
// 	suite.keeper.IterateClaims(suite.ctx, func(c types.Claim) (stop bool) {
// 		claims = append(claims, c)
// 		return false
// 	})
// 	suite.Equal(6, len(claims))
// }

// func (suite *KeeperTestSuite) setupCdpChain() {
// 	// creates a new test app with bnb as the only asset the pricefeed and cdp modules
// 	// funds three addresses and creates 3 cdps, funded with 100 BNB, 1000 BNB, and 10000 BNB
// 	// each CDP draws 10, 100, and 1000 USDX respectively
// 	// adds usdx incentives for bnb - 1000 KAVA per week with a 1 year time lock

// 	tApp := app.NewTestApp()
// 	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
// 	// need pricefeed and cdp gen state with one collateral
// 	pricefeedGS := pricefeed.GenesisState{
// 		Params: pricefeed.Params{
// 			Markets: []pricefeed.Market{
// 				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
// 			},
// 		},
// 		PostedPrices: []pricefeed.PostedPrice{
// 			{
// 				MarketID:      "bnb:usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         d("12.29"),
// 				Expiry:        time.Now().Add(100000 * time.Hour),
// 			},
// 		},
// 	}
// 	// need incentive params for one collateral
// 	cdpGS := cdp.GenesisState{
// 		Params: cdp.Params{
// 			GlobalDebtLimit:              sdk.NewInt64Coin("usdx", 1000000000000),
// 			SurplusAuctionThreshold:      cdp.DefaultSurplusThreshold,
// 			SurplusAuctionLot:            cdp.DefaultSurplusLot,
// 			DebtAuctionThreshold:         cdp.DefaultDebtThreshold,
// 			DebtAuctionLot:               cdp.DefaultDebtLot,
// 			SavingsDistributionFrequency: cdp.DefaultSavingsDistributionFrequency,
// 			CollateralParams: cdp.CollateralParams{
// 				{
// 					Denom:               "bnb",
// 					Type:                "bnb-a",
// 					LiquidationRatio:    sdk.MustNewDecFromStr("2.0"),
// 					DebtLimit:           sdk.NewInt64Coin("usdx", 1000000000000),
// 					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
// 					LiquidationPenalty:  d("0.05"),
// 					AuctionSize:         i(10000000000),
// 					Prefix:              0x20,
// 					SpotMarketID:        "bnb:usd",
// 					LiquidationMarketID: "bnb:usd",
// 					ConversionFactor:    i(8),
// 				},
// 			},
// 			DebtParam: cdp.DebtParam{
// 				Denom:            "usdx",
// 				ReferenceAsset:   "usd",
// 				ConversionFactor: i(6),
// 				DebtFloor:        i(10000000),
// 				SavingsRate:      d("0.95"),
// 			},
// 		},
// 		StartingCdpID:            cdp.DefaultCdpStartingID,
// 		DebtDenom:                cdp.DefaultDebtDenom,
// 		GovDenom:                 cdp.DefaultGovDenom,
// 		CDPs:                     cdp.CDPs{},
// 		PreviousDistributionTime: cdp.DefaultPreviousDistributionTime,
// 	}
// 	incentiveGS := types.NewGenesisState(
// 		types.NewParams(
// 			true, types.Rewards{types.NewReward(true, "bnb-a", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)},
// 		),
// 		types.DefaultPreviousBlockTime,
// 		types.RewardPeriods{types.NewRewardPeriod("bnb-a", ctx.BlockTime(), ctx.BlockTime().Add(time.Hour*7*24), c("ukava", 1000), ctx.BlockTime().Add(time.Hour*7*24*2), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 		types.ClaimPeriods{},
// 		types.Claims{},
// 		types.GenesisClaimPeriodIDs{})
// 	pricefeedAppGs := app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)}
// 	cdpAppGs := app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGS)}
// 	incentiveAppGs := app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(incentiveGS)}
// 	_, addrs := app.GeneratePrivKeyAddressPairs(3)
// 	authGS := app.NewAuthGenState(
// 		addrs[0:3],
// 		[]sdk.Coins{
// 			cs(c("bnb", 10000000000)),
// 			cs(c("bnb", 100000000000)),
// 			cs(c("bnb", 1000000000000)),
// 		})
// 	tApp.InitializeFromGenesisStates(
// 		authGS,
// 		pricefeedAppGs,
// 		incentiveAppGs,
// 		cdpAppGs,
// 	)
// 	suite.app = tApp
// 	suite.keeper = tApp.GetIncentiveKeeper()
// 	suite.ctx = ctx
// 	// create 3 cdps
// 	cdpKeeper := tApp.GetCDPKeeper()
// 	err := cdpKeeper.AddCdp(suite.ctx, addrs[0], c("bnb", 10000000000), c("usdx", 10000000), "bnb-a")
// 	suite.Require().NoError(err)
// 	err = cdpKeeper.AddCdp(suite.ctx, addrs[1], c("bnb", 100000000000), c("usdx", 100000000), "bnb-a")
// 	suite.Require().NoError(err)
// 	err = cdpKeeper.AddCdp(suite.ctx, addrs[2], c("bnb", 1000000000000), c("usdx", 1000000000), "bnb-a")
// 	suite.Require().NoError(err)
// 	// total usd is 1110

// 	// set the previous block time
// 	suite.keeper.SetPreviousBlockTime(suite.ctx, suite.ctx.BlockTime())
// }
