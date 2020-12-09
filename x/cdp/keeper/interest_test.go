package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
)

type InterestTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *InterestTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
}

// createCdps is a helper function to create two CDPs each with zero fees
func (suite *InterestTestSuite) createCdps() {
	// create 2 accounts in the state and give them some coins
	// create two private key pair addresses
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	ak := suite.app.GetAccountKeeper()
	// setup the first account
	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc.SetCoins(cs(c("xrp", 200000000), c("btc", 500000000)))

	ak.SetAccount(suite.ctx, acc)
	// now setup the second account
	acc2 := ak.NewAccountWithAddress(suite.ctx, addrs[1])
	acc2.SetCoins(cs(c("xrp", 200000000), c("btc", 500000000)))
	ak.SetAccount(suite.ctx, acc2)

	// now create two cdps with the addresses we just created
	// use the created account to create a cdp that SHOULD have fees updated
	// to get a ratio between 100 - 110% of liquidation ratio we can use 200xrp ($50) and 24 usdx (208% collateralization with liquidation ratio of 200%)
	// create CDP for the first address
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 200000000), c("usdx", 24000000), "xrp-a")
	suite.NoError(err) // check that no error was thrown

	// use the other account to create a cdp that SHOULD NOT have fees updated - 500% collateralization
	// create CDP for the second address
	err = suite.keeper.AddCdp(suite.ctx, addrs[1], c("xrp", 200000000), c("usdx", 10000000), "xrp-a")
	suite.NoError(err) // check that no error was thrown

}

func (suite *InterestTestSuite) TestCalculateInterestFactor() {
	type args struct {
		perSecondInterestRate sdk.Dec
		timeElapsed           sdk.Int
		expectedValue         sdk.Dec
	}

	type test struct {
		name string
		args args
	}

	oneYearInSeconds := int64(31536000)

	testCases := []test{
		{
			"1 year",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000005555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("1.191463614477847370"),
			},
		},
		{
			"10 year",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000005555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds * 10),
				expectedValue:         sdk.MustNewDecFromStr("5.765113233897391189"),
			},
		},
		{
			"1 month",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000005555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds / 12),
				expectedValue:         sdk.MustNewDecFromStr("1.014705619075717373"),
			},
		},
		{
			"1 day",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000005555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds / 365),
				expectedValue:         sdk.MustNewDecFromStr("1.000480067194057924"),
			},
		},
		{
			"1 year: low interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000000555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("1.017656545925063632"),
			},
		},
		{
			"1 year, lower interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000000055"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("1.001735985079841390"),
			},
		},
		{
			"1 year, lowest interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000000005"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("1.000157692432076670"),
			},
		},
		{
			"1 year: high interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000055555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("5.766022095987868825"),
			},
		},
		{
			"1 year: higher interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000000555555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("40628388.864535408465693310"),
			},
		},
		// If we raise the per second interest rate too much we'll cause an integer overflow.
		// For example, perSecondInterestRate: '1.000005555555' will cause a panic.
		{
			"1 year: highest interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000001555555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("2017093013158200407564.613502861572552603"),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			interestFactor := keeper.CalculateInterestFactor(tc.args.perSecondInterestRate, tc.args.timeElapsed)
			suite.Require().Equal(tc.args.expectedValue, interestFactor)
		})
	}
}

func (suite *InterestTestSuite) TestAccumulateInterest() {

	type args struct {
		ctype                   string
		initialTime             time.Time
		totalPrincipal          sdk.Int
		timeElapsed             int
		expectedTotalPrincipal  sdk.Int
		expectedLastAccrualTime time.Time
	}

	type test struct {
		name string
		args args
	}
	oneYearInSeconds := 31536000

	testCases := []test{
		{
			"1 year",
			args{
				ctype:                   "bnb-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				totalPrincipal:          sdk.NewInt(100000000000000),
				timeElapsed:             oneYearInSeconds,
				expectedTotalPrincipal:  sdk.NewInt(105000000000012),
				expectedLastAccrualTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Add(time.Duration(int(time.Second) * oneYearInSeconds)),
			},
		},
		{
			"1 year - zero principal",
			args{
				ctype:                   "bnb-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				totalPrincipal:          sdk.ZeroInt(),
				timeElapsed:             oneYearInSeconds,
				expectedTotalPrincipal:  sdk.ZeroInt(),
				expectedLastAccrualTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Add(time.Duration(int(time.Second) * oneYearInSeconds)),
			},
		},
		{
			"1 month",
			args{
				ctype:                   "bnb-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				totalPrincipal:          sdk.NewInt(100000000000000),
				timeElapsed:             86400 * 30,
				expectedTotalPrincipal:  sdk.NewInt(100401820189198),
				expectedLastAccrualTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Add(time.Duration(int(time.Second) * 86400 * 30)),
			},
		},
		{
			"1 month - interest rounds to zero",
			args{
				ctype:                   "bnb-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				totalPrincipal:          sdk.NewInt(10),
				timeElapsed:             86400 * 30,
				expectedTotalPrincipal:  sdk.NewInt(10),
				expectedLastAccrualTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			"7 seconds",
			args{
				ctype:                   "bnb-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				totalPrincipal:          sdk.NewInt(100000000000000),
				timeElapsed:             7,
				expectedTotalPrincipal:  sdk.NewInt(100000001082988),
				expectedLastAccrualTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Add(time.Duration(int(time.Second) * 7)),
			},
		},
		{
			"7 seconds - interest rounds to zero",
			args{
				ctype:                   "bnb-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				totalPrincipal:          sdk.NewInt(30000000),
				timeElapsed:             7,
				expectedTotalPrincipal:  sdk.NewInt(30000000),
				expectedLastAccrualTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			"7 seconds - zero interest",
			args{
				ctype:                   "busd-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				totalPrincipal:          sdk.NewInt(100000000000000),
				timeElapsed:             7,
				expectedTotalPrincipal:  sdk.NewInt(100000000000000),
				expectedLastAccrualTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Add(time.Duration(int(time.Second) * 7)),
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)
			suite.keeper.SetTotalPrincipal(suite.ctx, tc.args.ctype, types.DefaultStableDenom, tc.args.totalPrincipal)
			suite.keeper.SetPreviousAccrualTime(suite.ctx, tc.args.ctype, suite.ctx.BlockTime())
			suite.keeper.SetInterestFactor(suite.ctx, tc.args.ctype, sdk.OneDec())

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			err := suite.keeper.AccumulateInterest(suite.ctx, tc.args.ctype)
			suite.Require().NoError(err)

			actualTotalPrincipal := suite.keeper.GetTotalPrincipal(suite.ctx, tc.args.ctype, types.DefaultStableDenom)
			suite.Require().Equal(tc.args.expectedTotalPrincipal, actualTotalPrincipal)
			actualAccrualTime, _ := suite.keeper.GetPreviousAccrualTime(suite.ctx, tc.args.ctype)
			suite.Require().Equal(tc.args.expectedLastAccrualTime, actualAccrualTime)
		})
	}
}

// TestSynchronizeInterest tests the functionality of synchronizing the accumulated interest for CDPs
func (suite *InterestTestSuite) TestSynchronizeInterest() {
	type args struct {
		ctype                   string
		initialTime             time.Time
		initialCollateral       sdk.Coin
		initialPrincipal        sdk.Coin
		timeElapsed             int
		expectedFees            sdk.Coin
		expectedFeesUpdatedTime time.Time
	}

	type test struct {
		name string
		args args
	}

	oneYearInSeconds := 31536000
	testCases := []test{
		{
			"1 year",
			args{
				ctype:                   "bnb-a",
				initialTime:             time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				initialCollateral:       c("bnb", 1000000000000),
				initialPrincipal:        c("usdx", 100000000000),
				timeElapsed:             oneYearInSeconds,
				expectedFees:            c("usdx", 5000000000),
				expectedFeesUpdatedTime: time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Add(time.Duration(int(time.Second) * oneYearInSeconds)),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.ctx = suite.ctx.WithBlockTime(tc.args.initialTime)

			// setup account state
			_, addrs := app.GeneratePrivKeyAddressPairs(1)
			ak := suite.app.GetAccountKeeper()
			// setup the first account
			acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
			ak.SetAccount(suite.ctx, acc)
			sk := suite.app.GetSupplyKeeper()
			err := sk.MintCoins(suite.ctx, types.ModuleName, cs(tc.args.initialCollateral))
			suite.Require().NoError(err)
			err = sk.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, addrs[0], cs(tc.args.initialCollateral))
			suite.Require().NoError(err)

			// setup pricefeed
			pk := suite.app.GetPriceFeedKeeper()
			pk.SetPrice(suite.ctx, sdk.AccAddress{}, "bnb:usd", d("17.25"), tc.args.expectedFeesUpdatedTime.Add(time.Second))

			// setup cdp state
			suite.keeper.SetPreviousAccrualTime(suite.ctx, tc.args.ctype, suite.ctx.BlockTime())
			suite.keeper.SetInterestFactor(suite.ctx, tc.args.ctype, sdk.OneDec())
			err = suite.keeper.AddCdp(suite.ctx, addrs[0], tc.args.initialCollateral, tc.args.initialPrincipal, tc.args.ctype)

			updatedBlockTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			suite.ctx = suite.ctx.WithBlockTime(updatedBlockTime)
			err = suite.keeper.AccumulateInterest(suite.ctx, tc.args.ctype)
			suite.Require().NoError(err)

			cdp, found := suite.keeper.GetCDP(suite.ctx, tc.args.ctype, 1)
			suite.Require().True(found)

			cdp = suite.keeper.SynchronizeInterest(suite.ctx, cdp)

			suite.Require().Equal(tc.args.expectedFees, cdp.AccumulatedFees)
			suite.Require().Equal(tc.args.expectedFeesUpdatedTime, cdp.FeesUpdated)

		})
	}
}

func TestInterestTestSuite(t *testing.T) {
	suite.Run(t, new(InterestTestSuite))
}
