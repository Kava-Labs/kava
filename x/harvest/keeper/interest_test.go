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
			"normal no jump",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("3000"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("0.513333333333333334"),
			},
		},
		{
			"normal with jump",
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
		{
			"high reserves",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("1000000000000"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("0.180000000000000000"),
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
		interestFactor := harvest.CalculateInterestFactor(tc.args.perSecondInterestRate, tc.args.timeElapsed)
		suite.Require().Equal(tc.args.expectedValue, interestFactor)
	}
}

func (suite *InterestTestSuite) TestAPYToSPY() {
	type args struct {
		apy           sdk.Dec
		expectedValue sdk.Dec
	}

	type test struct {
		name        string
		args        args
		expectError bool
	}

	testCases := []test{
		{
			"lowest apy",
			args{
				apy:           sdk.MustNewDecFromStr("0.005"),
				expectedValue: sdk.MustNewDecFromStr("0.999999831991472557"),
			},
			false,
		},
		{
			"lower apy",
			args{
				apy:           sdk.MustNewDecFromStr("0.05"),
				expectedValue: sdk.MustNewDecFromStr("0.999999905005957279"),
			},
			false,
		},
		{
			"medium-low apy",
			args{
				apy:           sdk.MustNewDecFromStr("0.5"),
				expectedValue: sdk.MustNewDecFromStr("0.999999978020447332"),
			},
			false,
		},
		{
			"medium-high apy",
			args{
				apy:           sdk.MustNewDecFromStr("5"),
				expectedValue: sdk.MustNewDecFromStr("1.000000051034942717"),
			},
			false,
		},
		{
			"high apy",
			args{
				apy:           sdk.MustNewDecFromStr("50"),
				expectedValue: sdk.MustNewDecFromStr("1.000000124049443433"),
			},
			false,
		},
		{
			"highest apy",
			args{
				apy:           sdk.MustNewDecFromStr("170"),
				expectedValue: sdk.MustNewDecFromStr("1.000000162855113371"),
			},
			false,
		},
		{
			"out of bounds error at 178",
			args{
				apy:           sdk.MustNewDecFromStr("178"),
				expectedValue: sdk.ZeroDec(),
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			spy, err := harvest.APYToSPY(tc.args.apy)
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.args.expectedValue, spy)
			}
		})
	}
}

type ExpectedInterest struct {
	elapsedTime int64
}

func (suite *KeeperTestSuite) TestInterest() {
	type args struct {
		user                     sdk.AccAddress
		borrowCoinDenom          string
		borrowCoins              sdk.Coins
		interestRateModel        types.InterestRateModel
		reserveFactor            sdk.Dec
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

	normalModel := types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("0.5"))

	// oneDayInSeconds := int64(86400)
	// oneWeekInSeconds := int64(604800)
	// oneMonthInSeconds := int64(2592000)
	oneYearInSeconds := int64(31536000)

	testCases := []interestTest{
		{
			"valid",
			args{
				user:              sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				borrowCoinDenom:   "ukava",
				borrowCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel: normalModel,
				reserveFactor:     sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedInterest{
					ExpectedInterest{
						elapsedTime: oneYearInSeconds,
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		// {
		// 	"multiple snapshots",
		// 	args{
		// 		user:              sdk.AccAddress(crypto.AddressHash([]byte("test"))),
		// 		borrowCoinDenom:   "ukava",
		// 		borrowCoins:       sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
		// 		interestRateModel: normalModel,
		// 		reserveFactor:     sdk.MustNewDecFromStr("0.05"),
		// 		expectedInterestSnaphots: []ExpectedInterest{
		// 			ExpectedInterest{
		// 				elapsedTime:       oneMonthInSeconds,
		// 			},
		// 			ExpectedInterest{
		// 				elapsedTime:       oneYearInSeconds,
		// 			},
		// 		},
		// 	},
		// 	errArgs{
		// 		expectPass: true,
		// 		contains:   "",
		// 	},
		// },
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
						tc.args.interestRateModel,
						tc.args.reserveFactor), // Reserve Factor
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

			// Run begin blocker and store initial block time
			harvest.BeginBlocker(suite.ctx, suite.keeper)

			// Deposit 2x as many coins for each coin we intend to borrow
			for _, coin := range tc.args.borrowCoins {
				err = suite.keeper.Deposit(suite.ctx, tc.args.user, sdk.NewCoin(coin.Denom, coin.Amount.Mul(sdk.NewInt(2))))
				suite.Require().NoError(err)
			}

			// Borrow coins
			err = suite.keeper.Borrow(suite.ctx, tc.args.user, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Check that the initial module-level borrow balance is correct and store it
			initialBorrowedCoins, _ := suite.keeper.GetBorrowedCoins(suite.ctx)
			suite.Require().Equal(tc.args.borrowCoins, initialBorrowedCoins)

			// Check interest levels for each snapshot
			for _, snapshot := range tc.args.expectedInterestSnaphots {
				// ---------------------------- Calculate expected interest ----------------------------
				// 1. Get cash, borrows, reserves, and borrow index
				cashPrior := suite.getModuleAccountAtCtx(types.ModuleName, suite.ctx).GetCoins().AmountOf(tc.args.borrowCoinDenom)

				borrowCoinsPrior, borrowCoinsPriorFound := suite.keeper.GetBorrowedCoins(suite.ctx)
				suite.Require().True(borrowCoinsPriorFound)
				borrowCoinPriorAmount := borrowCoinsPrior.AmountOf(tc.args.borrowCoinDenom)

				reservesPrior, foundReservesPrior := suite.keeper.GetTotalReserves(suite.ctx, tc.args.borrowCoinDenom)
				if !foundReservesPrior {
					reservesPrior = sdk.NewCoin(tc.args.borrowCoinDenom, sdk.ZeroInt())
				}

				borrowIndexPrior, foundBorrowIndexPrior := suite.keeper.GetBorrowIndex(suite.ctx, tc.args.borrowCoinDenom)
				suite.Require().True(foundBorrowIndexPrior)

				// 2. Calculate expected interest owed
				borrowRateApy, err := harvest.CalculateBorrowRate(tc.args.interestRateModel, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowCoinPriorAmount), sdk.NewDecFromInt(reservesPrior.Amount))
				suite.Require().NoError(err)

				// Convert from APY to SPY, expressed as (1 + borrow rate)
				borrowRateSpy, err := harvest.APYToSPY(sdk.OneDec().Add(borrowRateApy))
				suite.Require().NoError(err)

				interestFactor := harvest.CalculateInterestFactor(borrowRateSpy, sdk.NewInt(snapshot.elapsedTime))
				expectedInterest := (interestFactor.Mul(sdk.NewDecFromInt(borrowCoinPriorAmount)).TruncateInt()).Sub(borrowCoinPriorAmount)
				expectedReserves := reservesPrior.Add(sdk.NewCoin(tc.args.borrowCoinDenom, sdk.NewDecFromInt(expectedInterest).Mul(tc.args.reserveFactor).TruncateInt()))
				expectedBorrowIndex := borrowIndexPrior.Mul(interestFactor)
				// -------------------------------------------------------------------------------------

				// Set up snapshot chain context and run begin blocker
				runAtTime := time.Unix(suite.ctx.BlockTime().Unix()+snapshot.elapsedTime, 0)
				snapshotCtx := suite.ctx.WithBlockTime(runAtTime)
				harvest.BeginBlocker(snapshotCtx, suite.keeper)

				// Check that the total amount of borrowed coins has increased by expected interest amount
				expectedBorrowedCoins := initialBorrowedCoins.AmountOf(tc.args.borrowCoinDenom).Add(expectedInterest)
				currBorrowedCoins, _ := suite.keeper.GetBorrowedCoins(snapshotCtx)
				suite.Require().Equal(expectedBorrowedCoins, currBorrowedCoins.AmountOf(tc.args.borrowCoinDenom))

				// Check that the total reserves have changed as expected
				currTotalReserves, _ := suite.keeper.GetTotalReserves(snapshotCtx, tc.args.borrowCoinDenom)
				diffTotalReserves := currTotalReserves.Sub(reservesPrior)
				suite.Require().Equal(expectedReserves, diffTotalReserves)

				// Check that the borrow index has increased as expected
				currIndexPrior, _ := suite.keeper.GetBorrowIndex(snapshotCtx, tc.args.borrowCoinDenom)
				suite.Require().Equal(expectedBorrowIndex, currIndexPrior)

				// TODO: Add a clause for successive borrows, then check that the user's
				// 		 borrow balance has been correctly updated.
			}
		})
	}
}

func TestInterestTestSuite(t *testing.T) {
	suite.Run(t, new(InterestTestSuite))
}
