package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/pricefeed"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func (suite *KeeperTestSuite) TestExpireRewardPeriod() {
	rp := types.NewRewardPeriod("bnb", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), time.Hour*8766)
	suite.keeper.SetRewardPeriod(suite.ctx, rp)
	suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
	suite.NotPanics(func() {
		suite.keeper.HandleRewardPeriodExpiry(suite.ctx, rp)
	})
	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.True(found)
}

func (suite *KeeperTestSuite) TestAddToClaim() {
	rp := types.NewRewardPeriod("bnb", suite.ctx.BlockTime(), suite.ctx.BlockTime().Add(time.Hour*168), c("ukava", 100000000), suite.ctx.BlockTime().Add(time.Hour*168*2), time.Hour*8766)
	suite.keeper.SetRewardPeriod(suite.ctx, rp)
	suite.keeper.SetNextClaimPeriodID(suite.ctx, "bnb", 1)
	suite.keeper.HandleRewardPeriodExpiry(suite.ctx, rp)
	c1 := types.NewClaim(suite.addrs[0], c("ukava", 1000000), "bnb", 1)
	suite.keeper.SetClaim(suite.ctx, c1)
	suite.NotPanics(func() {
		suite.keeper.AddToClaim(suite.ctx, suite.addrs[0], "bnb", 1, c("ukava", 1000000))
	})
	testC, _ := suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.Equal(c("ukava", 2000000), testC.Reward)

	suite.NotPanics(func() {
		suite.keeper.AddToClaim(suite.ctx, suite.addrs[0], "xpr", 1, c("ukava", 1000000))
	})
}

func (suite *KeeperTestSuite) TestCreateRewardPeriod() {
	reward := types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)
	suite.NotPanics(func() {
		suite.keeper.CreateNewRewardPeriod(suite.ctx, reward)
	})
	_, found := suite.keeper.GetRewardPeriod(suite.ctx, "bnb")
	suite.True(found)
}

func (suite *KeeperTestSuite) TestCreateAndDeleteRewardsPeriods() {
	reward1 := types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)
	reward2 := types.NewReward(false, "xrp", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)
	params := types.NewParams(true, types.Rewards{reward1, reward2})
	suite.keeper.SetParams(suite.ctx, params)

	suite.NotPanics(func() {
		suite.keeper.CreateAndDeleteRewardPeriods(suite.ctx)
	})
	testCases := []struct {
		name        string
		arg         string
		expectFound bool
	}{
		{
			"active reward period",
			"bnb",
			true,
		},
		{
			"inactive reward period",
			"xrp",
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, found := suite.keeper.GetRewardPeriod(suite.ctx, tc.arg)
			if tc.expectFound {
				suite.True(found)
			} else {
				suite.False(found)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestApplyRewardsToCdps() {
	suite.setupCdpChain() // creates a test app with 3 BNB cdps and usdx incentives for bnb - each reward period is one week

	// move the context forward by 100 periods
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 100))
	// apply rewards to BNB cdps
	suite.NotPanics(func() {
		suite.keeper.ApplyRewardsToCdps(suite.ctx)
	})
	// each cdp should have a claim
	claims := types.Claims{}
	suite.keeper.IterateClaims(suite.ctx, func(c types.Claim) (stop bool) {
		claims = append(claims, c)
		return false
	})
	suite.Equal(3, len(claims))
	// there should be no associated claim period, because the reward period has not ended yet
	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.False(found)

	// move ctx to the reward period expiry and check that the claim period has been created and the next claim period id has increased
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour * 24 * 7))

	suite.NotPanics(func() {
		// apply rewards to cdps
		suite.keeper.ApplyRewardsToCdps(suite.ctx)
		// delete the old reward period amd create a new one
		suite.keeper.CreateAndDeleteRewardPeriods(suite.ctx)
	})
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.True(found)
	testID := suite.keeper.GetNextClaimPeriodID(suite.ctx, "bnb")
	suite.Equal(uint64(2), testID)

	// move the context forward by 100 periods
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 100))
	// run the begin blocker functions
	suite.NotPanics(func() {
		suite.keeper.DeleteExpiredClaimsAndClaimPeriods(suite.ctx)
		suite.keeper.ApplyRewardsToCdps(suite.ctx)
		suite.keeper.CreateAndDeleteRewardPeriods(suite.ctx)
	})
	// each cdp should now have two claims
	claims = types.Claims{}
	suite.keeper.IterateClaims(suite.ctx, func(c types.Claim) (stop bool) {
		claims = append(claims, c)
		return false
	})
	suite.Equal(6, len(claims))
}

func (suite *KeeperTestSuite) setupCdpChain() {
	// creates a new test app with bnb as the only asset the pricefeed and cdp modules
	// funds three addresses and creates 3 cdps, funded with 100 BNB, 1000 BNB, and 10000 BNB
	// each CDP draws 10, 100, and 1000 USDX respectively
	// adds usdx incentives for bnb - 1000 KAVA per week with a 1 year time lock

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	// need pricefeed and cdp gen state with one collateral
	pricefeedGS := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				pricefeed.Market{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			pricefeed.PostedPrice{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         d("12.29"),
				Expiry:        time.Now().Add(100000 * time.Hour),
			},
		},
	}
	// need incentive params for one collateral
	cdpGS := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:              sdk.NewCoins(sdk.NewInt64Coin("usdx", 1000000000000)),
			SurplusAuctionThreshold:      cdp.DefaultSurplusThreshold,
			DebtAuctionThreshold:         cdp.DefaultDebtThreshold,
			SavingsDistributionFrequency: cdp.DefaultSavingsDistributionFrequency,
			CollateralParams: cdp.CollateralParams{
				{
					Denom:              "bnb",
					LiquidationRatio:   sdk.MustNewDecFromStr("2.0"),
					DebtLimit:          sdk.NewCoins(sdk.NewInt64Coin("usdx", 1000000000000)),
					StabilityFee:       sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty: d("0.05"),
					AuctionSize:        i(10000000000),
					Prefix:             0x20,
					MarketID:           "bnb:usd",
					ConversionFactor:   i(8),
				},
			},
			DebtParams: cdp.DebtParams{
				{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: i(6),
					DebtFloor:        i(10000000),
					SavingsRate:      d("0.95"),
				},
			},
		},
		StartingCdpID:            cdp.DefaultCdpStartingID,
		DebtDenom:                cdp.DefaultDebtDenom,
		GovDenom:                 cdp.DefaultGovDenom,
		CDPs:                     cdp.CDPs{},
		PreviousBlockTime:        cdp.DefaultPreviousBlockTime,
		PreviousDistributionTime: cdp.DefaultPreviousDistributionTime,
	}
	incentiveGS := types.NewGenesisState(
		types.NewParams(
			true, types.Rewards{types.NewReward(true, "bnb", c("ukava", 1000000000), time.Hour*7*24, time.Hour*24*365, time.Hour*7*24)},
		),
		types.DefaultPreviousBlockTime,
		types.RewardPeriods{types.NewRewardPeriod("bnb", ctx.BlockTime(), ctx.BlockTime().Add(time.Hour*7*24), c("ukava", 1000), ctx.BlockTime().Add(time.Hour*7*24*2), time.Hour*365*24)},
		types.ClaimPeriods{},
		types.Claims{},
		types.GenesisClaimPeriodIDs{})
	pricefeedAppGs := app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)}
	cdpAppGs := app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGS)}
	incentiveAppGs := app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(incentiveGS)}
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	authGS := app.NewAuthGenState(
		addrs[0:3],
		[]sdk.Coins{
			cs(c("bnb", 10000000000)),
			cs(c("bnb", 100000000000)),
			cs(c("bnb", 1000000000000)),
		})
	tApp.InitializeFromGenesisStates(
		authGS,
		pricefeedAppGs,
		incentiveAppGs,
		cdpAppGs,
	)
	suite.app = tApp
	suite.keeper = tApp.GetIncentiveKeeper()
	suite.ctx = ctx
	// create 3 cdps
	cdpKeeper := tApp.GetCDPKeeper()
	err := cdpKeeper.AddCdp(suite.ctx, addrs[0], cs(c("bnb", 10000000000)), cs(c("usdx", 10000000)))
	suite.NoError(err)
	err = cdpKeeper.AddCdp(suite.ctx, addrs[1], cs(c("bnb", 100000000000)), cs(c("usdx", 100000000)))
	suite.NoError(err)
	err = cdpKeeper.AddCdp(suite.ctx, addrs[2], cs(c("bnb", 1000000000000)), cs(c("usdx", 1000000000)))
	suite.NoError(err)
	// total usd is 1110

	// set the previous block time
	suite.keeper.SetPreviousBlockTime(suite.ctx, suite.ctx.BlockTime())
}

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
