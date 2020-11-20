package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/harvest"
	"github.com/kava-labs/kava/x/harvest/types"
	"github.com/kava-labs/kava/x/pricefeed"
)

type InterestTestSuite struct {
	suite.Suite
}

func (suite *InterestTestSuite) TestCalculateUtilizationRatio() {
	type args struct {
		cash          sdk.Dec
		borrows       sdk.Dec
		reserves      sdk.Dec
		expectedValue sdk.Dec
	}

	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"normal",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				expectedValue: sdk.MustNewDecFromStr("0.847457627118644068"),
			},
		},
		{
			"high util ratio",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("250000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				expectedValue: sdk.MustNewDecFromStr("0.996412913511359107"),
			},
		},
		{
			"very high util ratio",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("250000000000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				expectedValue: sdk.MustNewDecFromStr("0.999999996400000013"),
			},
		},
		{
			"low util ratio",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("50"),
				reserves:      sdk.MustNewDecFromStr("100"),
				expectedValue: sdk.MustNewDecFromStr("0.052631578947368421"),
			},
		},
		{
			"very low util ratio",
			args{
				cash:          sdk.MustNewDecFromStr("10000000"),
				borrows:       sdk.MustNewDecFromStr("50"),
				reserves:      sdk.MustNewDecFromStr("100"),
				expectedValue: sdk.MustNewDecFromStr("0.000005000025000125"),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			utilRatio := harvest.CalculateUtilizationRatio(tc.args.cash, tc.args.borrows, tc.args.reserves)
			suite.Require().Equal(tc.args.expectedValue, utilRatio)
		})
	}
}

func (suite *InterestTestSuite) TestCalculateBorrowRate() {
	type args struct {
		cash          sdk.Dec
		borrows       sdk.Dec
		reserves      sdk.Dec
		model         types.InterestRateModel
		expectedValue sdk.Dec
	}

	type test struct {
		name string
		args args
	}

	// Normal model has:
	// 	- BaseRateAPY:      0.0
	// 	- BaseMultiplier:   0.1
	// 	- Kink:             0.8
	// 	- JumpMultiplier:   0.5
	normalModel := types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.5"))

	testCases := []test{
		{
			"normal",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("0.103728813559322034"),
			},
		},
		{
			"high cash",
			args{
				cash:          sdk.MustNewDecFromStr("10000000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("0.000049975511999120"),
			},
		},
		{
			"high borrows",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000000000000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("0.179999999910000000"),
			},
		},
		// TODO: high reserves results in a negative value
		{
			"high reserves",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("1000000000000"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("-0.000000000500000003"),
			},
		},
		{
			"random numbers",
			args{
				cash:          sdk.MustNewDecFromStr("125"),
				borrows:       sdk.MustNewDecFromStr("11"),
				reserves:      sdk.MustNewDecFromStr("82"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("0.020370370370370370"),
			},
		},
		{
			"increased base multiplier",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				model:         types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.5")),
				expectedValue: sdk.MustNewDecFromStr("0.423728813559322034"),
			},
		},
		{
			"decreased kink",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				model:         types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.5")),
				expectedValue: sdk.MustNewDecFromStr("0.423728813559322034"),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			borrowRate, err := harvest.CalculateBorrowRate(tc.args.model, tc.args.cash, tc.args.borrows, tc.args.reserves)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.args.expectedValue, borrowRate)
		})
	}
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
				perSecondInterestRate: sdk.MustNewDecFromStr("0.999999999"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("0.968956073391928722"),
			},
		},
		{
			"10 year",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("0.999999999"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds * 10),
				expectedValue:         sdk.MustNewDecFromStr("0.729526197443942502"),
			},
		},
		{
			"1 month",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("0.999999999"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds / 12),
				expectedValue:         sdk.MustNewDecFromStr("0.997375450167679744"),
			},
		},
		{
			"1 day",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("0.999999999"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds / 365),
				expectedValue:         sdk.MustNewDecFromStr("0.999913603732329314"),
			},
		},
		{
			"1 year: low interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("0.999999998"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("0.938875872133496909"),
			},
		},
		{
			"1 year, lower interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("0.999999995"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("0.854123057283825426"),
			},
		},
		{
			"1 year, lowest interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("0.99999995"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("0.206635266092017207"),
			},
		},
	}

	for _, tc := range testCases {
		interestFactor := harvest.CalculateInterestFactor(tc.args.perSecondInterestRate, tc.args.timeElapsed)
		suite.Require().Equal(tc.args.expectedValue, interestFactor)
	}
}

type ExpectedInterest struct {
	height    int64
	coinsUser sdk.Coins
}

func (suite *KeeperTestSuite) TestInterest() {

	type args struct {
		user                     sdk.AccAddress
		borrowCoins              sdk.Coins
		expectedInterestSnaphots []ExpectedInterest
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type interestTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []interestTest{
		{
			"valid",
			args{
				user:        sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				borrowCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				expectedInterestSnaphots: []ExpectedInterest{
					ExpectedInterest{
						height:    10,
						coinsUser: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
					},
					ExpectedInterest{
						height:    500,
						coinsUser: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

			// Auth module genesis state
			authGS := app.NewAuthGenState(
				[]sdk.AccAddress{tc.args.user},
				[]sdk.Coins{sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)))},
			)

			// Harvest module genesis state
			harvestGS := types.NewGenesisState(types.NewParams(
				true,
				types.DistributionSchedules{
					types.NewDistributionSchedule(true, "ukava", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 22, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(5000)), time.Date(2021, 11, 22, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
				},
				types.DelegatorDistributionSchedules{types.NewDelegatorDistributionSchedule(
					types.NewDistributionSchedule(true, "usdx", time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC), time.Date(2025, 10, 8, 14, 0, 0, 0, time.UTC), sdk.NewCoin("hard", sdk.NewInt(500)), time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Medium, 6, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Medium, 24, sdk.OneDec())}),
					time.Hour*24,
				),
				},
				types.MoneyMarkets{
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",          // Market ID
						sdk.NewInt(KAVA_CF), // Conversion Factor
						types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.1")),
						sdk.MustNewDecFromStr("0.05")), // Reserve Factor
				},
			), types.DefaultPreviousBlockTime, types.DefaultDistributionTimes)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeed.PostedPrice{
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("2.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
				},
			}

			// Initialize test application
			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pricefeedGS)},
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(harvestGS)})

			// Mint coins to Harvest module account
			supplyKeeper := tApp.GetSupplyKeeper()
			harvestMaccCoins := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF)))
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, harvestMaccCoins)

			keeper := tApp.GetHarvestKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			var err error

			// Deposit 2x as many coins for each coin we intend to borrow
			for _, coin := range tc.args.borrowCoins {
				err = suite.keeper.Deposit(suite.ctx, tc.args.user, sdk.NewCoin(coin.Denom, coin.Amount.Mul(sdk.NewInt(2))))
				suite.Require().NoError(err)
			}

			// Borrow coins
			err = suite.keeper.Borrow(suite.ctx, tc.args.user, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Check interest levels for each snapshot
			for _, snapshot := range tc.args.expectedInterestSnaphots {
				// Set up snapshot chain context
				snapshotCtx := suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + snapshot.height)
				harvest.BeginBlocker(snapshotCtx, suite.keeper)

				// TODO: borrowedCoins in store
				// borrowedCoins, _ := suite.keeper.GetBorrowedCoins(snapshotCtx)

				// TODO: borrow in store
				// borrow, _ := suite.keeper.GetBorrow(snapshotCtx, tc.args.user)
				// fmt.Println("borrow:", borrow)

				// TODO: check reserves

				// TODO: should we check the module account balance?
				// mAcc := suite.getModuleAccount(types.ModuleAccountName)
				// suite.Require().Equal(tc.args.expectedModAccountBalance.Add(depositedCoins...), mAcc.GetCoins())
			}
		})
	}
}

func TestInterestTestSuite(t *testing.T) {
	suite.Run(t, new(InterestTestSuite))
}
