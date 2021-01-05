package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/hard/types"
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
			utilRatio := hard.CalculateUtilizationRatio(tc.args.cash, tc.args.borrows, tc.args.reserves)
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
				cash:          sdk.MustNewDecFromStr("5000"),
				borrows:       sdk.MustNewDecFromStr("1000"),
				reserves:      sdk.MustNewDecFromStr("1000"),
				model:         normalModel,
				expectedValue: sdk.MustNewDecFromStr("0.020000000000000000"),
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
				model:         types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("1.0")),
				expectedValue: sdk.MustNewDecFromStr("0.447457627118644068"),
			},
		},
		{
			"decreased kink",
			args{
				cash:          sdk.MustNewDecFromStr("1000"),
				borrows:       sdk.MustNewDecFromStr("5000"),
				reserves:      sdk.MustNewDecFromStr("100"),
				model:         types.NewInterestRateModel(sdk.MustNewDecFromStr("0"), sdk.MustNewDecFromStr("0.5"), sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("1.0")),
				expectedValue: sdk.MustNewDecFromStr("0.797457627118644068"),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			borrowRate, err := hard.CalculateBorrowRate(tc.args.model, tc.args.cash, tc.args.borrows, tc.args.reserves)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.args.expectedValue, borrowRate)
		})
	}
}

func (suite *InterestTestSuite) TestCalculateBorrowInterestFactor() {
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
		interestFactor := hard.CalculateBorrowInterestFactor(tc.args.perSecondInterestRate, tc.args.timeElapsed)
		suite.Require().Equal(tc.args.expectedValue, interestFactor)
	}
}

func (suite *InterestTestSuite) TestCalculateSupplyInterestFactor() {
	type args struct {
		borrowInterestFactorIncrement sdk.Dec
		cash                          sdk.Dec
		borrows                       sdk.Dec
		reserves                      sdk.Dec
		reserveFactor                 sdk.Dec
		expectedValue                 sdk.Dec
	}

	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"low borrow interest factor increment",
			args{
				borrowInterestFactorIncrement: sdk.MustNewDecFromStr("0.0000000000005"),
				cash:                          sdk.MustNewDecFromStr("100.0"),
				borrows:                       sdk.MustNewDecFromStr("1000.0"),
				reserves:                      sdk.MustNewDecFromStr("10.0"),
				reserveFactor:                 sdk.MustNewDecFromStr("0.05"),
				expectedValue:                 sdk.MustNewDecFromStr("1.000000000000435780"),
			},
		},
		{
			"medium borrow interest factor increment",
			args{
				borrowInterestFactorIncrement: sdk.MustNewDecFromStr("0.000005"),
				cash:                          sdk.MustNewDecFromStr("100.0"),
				borrows:                       sdk.MustNewDecFromStr("1000.0"),
				reserves:                      sdk.MustNewDecFromStr("10.0"),
				reserveFactor:                 sdk.MustNewDecFromStr("0.05"),
				expectedValue:                 sdk.MustNewDecFromStr("1.000004357798165138"),
			},
		},
		{
			"high borrow interest factor increment",
			args{
				borrowInterestFactorIncrement: sdk.MustNewDecFromStr("0.05"),
				cash:                          sdk.MustNewDecFromStr("100.0"),
				borrows:                       sdk.MustNewDecFromStr("1000.0"),
				reserves:                      sdk.MustNewDecFromStr("10.0"),
				reserveFactor:                 sdk.MustNewDecFromStr("0.05"),
				expectedValue:                 sdk.MustNewDecFromStr("1.043577981651376147"),
			},
		},
	}

	for _, tc := range testCases {
		interestFactor := hard.CalculateSupplyInterestFactor(tc.args.borrowInterestFactorIncrement,
			tc.args.cash, tc.args.borrows, tc.args.reserves, tc.args.reserveFactor)
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
				apy:           sdk.MustNewDecFromStr("177"),
				expectedValue: sdk.MustNewDecFromStr("1.000002441641340532"),
			},
			false,
		},
		{
			"out of bounds error after 178",
			args{
				apy:           sdk.MustNewDecFromStr("178"),
				expectedValue: sdk.ZeroDec(),
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			spy, err := hard.APYToSPY(tc.args.apy)
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.args.expectedValue, spy)
			}
		})
	}
}

type ExpectedBorrowInterest struct {
	elapsedTime  int64
	shouldBorrow bool
	borrowCoin   sdk.Coin
}

func (suite *KeeperTestSuite) TestBorrowInterest() {
	type args struct {
		user                     sdk.AccAddress
		initialBorrowerCoins     sdk.Coins
		initialModuleCoins       sdk.Coins
		borrowCoinDenom          string
		borrowCoins              sdk.Coins
		interestRateModel        types.InterestRateModel
		reserveFactor            sdk.Dec
		expectedInterestSnaphots []ExpectedBorrowInterest
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

	oneDayInSeconds := int64(86400)
	oneWeekInSeconds := int64(604800)
	oneMonthInSeconds := int64(2592000)
	oneYearInSeconds := int64(31536000)

	testCases := []interestTest{
		{
			"one day",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneDayInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"one week",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneWeekInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"one month",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"one year",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneYearInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"0 reserve factor",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneYearInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"borrow during snapshot",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneYearInSeconds,
						shouldBorrow: true,
						borrowCoin:   sdk.NewCoin("ukava", sdk.NewInt(1*KAVA_CF)),
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"multiple snapshots",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneMonthInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"varied snapshots",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedBorrowInterest{
					{
						elapsedTime:  oneDayInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneWeekInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneMonthInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneYearInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
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
				[]sdk.Coins{tc.args.initialBorrowerCoins},
			)

			// Hard module genesis state
			hardGS := types.NewGenesisState(types.NewParams(
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
						"kava:usd",                // Market ID
						sdk.NewInt(KAVA_CF),       // Conversion Factor
						sdk.NewInt(USDX_CF*1000),  // Auction Size
						tc.args.interestRateModel, // Interest Rate Model
						tc.args.reserveFactor,     // Reserve Factor
						sdk.ZeroDec()),            // Keeper Reward Percentage
				},
				0, // LTV counter
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
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(hardGS)})

			// Mint coins to Hard module account
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, tc.args.initialModuleCoins)

			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			var err error

			// Run begin blocker and store initial block time
			hard.BeginBlocker(suite.ctx, suite.keeper)

			// Deposit 2x as many coins for each coin we intend to borrow
			depositCoins := sdk.NewCoins()
			for _, borrowCoin := range tc.args.borrowCoins {
				depositCoins = depositCoins.Add(sdk.NewCoin(borrowCoin.Denom, borrowCoin.Amount.Mul(sdk.NewInt(2))))
			}
			err = suite.keeper.Deposit(suite.ctx, tc.args.user, depositCoins)
			suite.Require().NoError(err)

			// Borrow coins
			err = suite.keeper.Borrow(suite.ctx, tc.args.user, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Check that the initial module-level borrow balance is correct and store it
			initialBorrowedCoins, _ := suite.keeper.GetBorrowedCoins(suite.ctx)
			suite.Require().Equal(tc.args.borrowCoins, initialBorrowedCoins)

			// Check interest levels for each snapshot
			prevCtx := suite.ctx
			for _, snapshot := range tc.args.expectedInterestSnaphots {
				// ---------------------------- Calculate expected interest ----------------------------
				// 1. Get cash, borrows, reserves, and borrow index
				cashPrior := suite.getModuleAccountAtCtx(types.ModuleName, prevCtx).GetCoins().AmountOf(tc.args.borrowCoinDenom)

				borrowCoinsPrior, borrowCoinsPriorFound := suite.keeper.GetBorrowedCoins(prevCtx)
				suite.Require().True(borrowCoinsPriorFound)
				borrowCoinPriorAmount := borrowCoinsPrior.AmountOf(tc.args.borrowCoinDenom)

				reservesPrior, foundReservesPrior := suite.keeper.GetTotalReserves(prevCtx, tc.args.borrowCoinDenom)
				if !foundReservesPrior {
					reservesPrior = sdk.NewCoin(tc.args.borrowCoinDenom, sdk.ZeroInt())
				}

				interestFactorPrior, foundInterestFactorPrior := suite.keeper.GetBorrowInterestFactor(prevCtx, tc.args.borrowCoinDenom)
				suite.Require().True(foundInterestFactorPrior)

				// 2. Calculate expected interest owed
				borrowRateApy, err := hard.CalculateBorrowRate(tc.args.interestRateModel, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowCoinPriorAmount), sdk.NewDecFromInt(reservesPrior.Amount))
				suite.Require().NoError(err)

				// Convert from APY to SPY, expressed as (1 + borrow rate)
				borrowRateSpy, err := hard.APYToSPY(sdk.OneDec().Add(borrowRateApy))
				suite.Require().NoError(err)

				interestFactor := hard.CalculateBorrowInterestFactor(borrowRateSpy, sdk.NewInt(snapshot.elapsedTime))
				expectedInterest := (interestFactor.Mul(sdk.NewDecFromInt(borrowCoinPriorAmount)).TruncateInt()).Sub(borrowCoinPriorAmount)
				expectedReserves := reservesPrior.Add(sdk.NewCoin(tc.args.borrowCoinDenom, sdk.NewDecFromInt(expectedInterest).Mul(tc.args.reserveFactor).TruncateInt()))
				expectedInterestFactor := interestFactorPrior.Mul(interestFactor)
				// -------------------------------------------------------------------------------------

				// Set up snapshot chain context and run begin blocker
				runAtTime := time.Unix(prevCtx.BlockTime().Unix()+(snapshot.elapsedTime), 0)
				snapshotCtx := prevCtx.WithBlockTime(runAtTime)
				hard.BeginBlocker(snapshotCtx, suite.keeper)

				// Check that the total amount of borrowed coins has increased by expected interest amount
				expectedBorrowedCoins := borrowCoinsPrior.AmountOf(tc.args.borrowCoinDenom).Add(expectedInterest)
				currBorrowedCoins, _ := suite.keeper.GetBorrowedCoins(snapshotCtx)
				suite.Require().Equal(expectedBorrowedCoins, currBorrowedCoins.AmountOf(tc.args.borrowCoinDenom))

				// Check that the total reserves have changed as expected
				currTotalReserves, _ := suite.keeper.GetTotalReserves(snapshotCtx, tc.args.borrowCoinDenom)
				suite.Require().Equal(expectedReserves, currTotalReserves)

				// Check that the borrow index has increased as expected
				currIndexPrior, _ := suite.keeper.GetBorrowInterestFactor(snapshotCtx, tc.args.borrowCoinDenom)
				suite.Require().Equal(expectedInterestFactor, currIndexPrior)

				// After borrowing again user's borrow balance should have any outstanding interest applied
				if snapshot.shouldBorrow {
					borrowCoinsBefore, _ := suite.keeper.GetBorrow(snapshotCtx, tc.args.user)
					expectedInterestCoins := sdk.NewCoin(tc.args.borrowCoinDenom, expectedInterest)
					expectedBorrowCoinsAfter := borrowCoinsBefore.Amount.Add(snapshot.borrowCoin).Add(expectedInterestCoins)

					err = suite.keeper.Borrow(snapshotCtx, tc.args.user, sdk.NewCoins(snapshot.borrowCoin))
					suite.Require().NoError(err)

					borrowCoinsAfter, _ := suite.keeper.GetBorrow(snapshotCtx, tc.args.user)
					suite.Require().Equal(expectedBorrowCoinsAfter, borrowCoinsAfter.Amount)
				}
				// Update previous context to this snapshot's context, segmenting time periods between snapshots
				prevCtx = snapshotCtx
			}
		})
	}
}

type ExpectedSupplyInterest struct {
	elapsedTime  int64
	shouldBorrow bool
	borrowCoin   sdk.Coin
	supplyCoin   sdk.Coin
}

func (suite *KeeperTestSuite) TestSupplyInterest() {
	type args struct {
		user                     sdk.AccAddress
		initialSupplierCoins     sdk.Coins
		initialBorrowerCoins     sdk.Coins
		initialModuleCoins       sdk.Coins
		depositCoins             sdk.Coins
		borrowCoinDenom          string
		borrowCoins              sdk.Coins
		interestRateModel        types.InterestRateModel
		reserveFactor            sdk.Dec
		expectedInterestSnaphots []ExpectedSupplyInterest
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

	oneMonthInSeconds := int64(2592000)

	testCases := []interestTest{
		{
			"one month",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				borrowCoinDenom:      "ukava",
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(80*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldBorrow: false,
						borrowCoin:   sdk.Coin{},
						supplyCoin:   sdk.Coin{},
					},
					// {
					// 	elapsedTime:  oneMonthInSeconds,
					// 	shouldBorrow: false,
					// 	borrowCoin:   sdk.Coin{},
					// 	supplyCoin:   sdk.Coin{},
					// },
					// {
					// 	elapsedTime:  oneMonthInSeconds,
					// 	shouldBorrow: false,
					// 	borrowCoin:   sdk.Coin{},
					// 	supplyCoin:   sdk.Coin{},
					// },
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
				[]sdk.Coins{tc.args.initialBorrowerCoins},
			)

			// Hard module genesis state
			hardGS := types.NewGenesisState(types.NewParams(
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
						"kava:usd",                // Market ID
						sdk.NewInt(KAVA_CF),       // Conversion Factor
						sdk.NewInt(USDX_CF*1000),  // Auction Size
						tc.args.interestRateModel, // Interest Rate Model
						tc.args.reserveFactor,     // Reserve Factor
						sdk.ZeroDec()),            // Keeper Reward Percentage
				},
				0, // LTV counter
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
				app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(hardGS)})

			// Mint coins to Hard module account
			supplyKeeper := tApp.GetSupplyKeeper()
			supplyKeeper.MintCoins(ctx, types.ModuleAccountName, tc.args.initialModuleCoins)

			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper
			suite.keeper.SetSuppliedCoins(ctx, tc.args.initialModuleCoins)

			var err error

			// Run begin blocker
			hard.BeginBlocker(suite.ctx, suite.keeper)

			// // Deposit coins
			err = suite.keeper.Deposit(suite.ctx, tc.args.user, tc.args.depositCoins)
			suite.Require().NoError(err)

			// Borrow coins
			err = suite.keeper.Borrow(suite.ctx, tc.args.user, tc.args.borrowCoins)
			suite.Require().NoError(err)

			// Check interest levels for each snapshot
			prevCtx := suite.ctx
			for i, snapshot := range tc.args.expectedInterestSnaphots {
				fmt.Println()
				fmt.Println("RUNNING ITERATION:", i)
				// ---------------------------- Calculate expected supply interest ----------------------------
				// 1. Get cash, borrows, reserves, and borrow index
				cashPrior := suite.getModuleAccountAtCtx(types.ModuleName, prevCtx).GetCoins().AmountOf(tc.args.borrowCoinDenom)
				fmt.Printf("[TEST] cashPrior: %s\n", cashPrior)

				var borrowCoinPriorAmount sdk.Int
				borrowCoinsPrior, borrowCoinsPriorFound := suite.keeper.GetBorrowedCoins(prevCtx)
				suite.Require().True(borrowCoinsPriorFound)
				borrowCoinPriorAmount = borrowCoinsPrior.AmountOf(tc.args.borrowCoinDenom)

				var supplyCoinPriorAmount sdk.Int
				supplyCoinsPrior, supplyCoinsPriorFound := suite.keeper.GetSuppliedCoins(prevCtx)
				suite.Require().True(supplyCoinsPriorFound)
				supplyCoinPriorAmount = supplyCoinsPrior.AmountOf(tc.args.borrowCoinDenom)

				reservesPrior, foundReservesPrior := suite.keeper.GetTotalReserves(prevCtx, tc.args.borrowCoinDenom)
				if !foundReservesPrior {
					reservesPrior = sdk.NewCoin(tc.args.borrowCoinDenom, sdk.ZeroInt())
				}
				fmt.Printf("- [TEST] reservesPrior: %s\n", reservesPrior)

				borrowInterestFactorPrior, foundBorrowInterestFactorPrior := suite.keeper.GetBorrowInterestFactor(prevCtx, tc.args.borrowCoinDenom)
				suite.Require().True(foundBorrowInterestFactorPrior)
				fmt.Printf("[TEST] Borrow Interest Factor Prior: %s\n", borrowInterestFactorPrior)

				supplyInterestFactorPrior, foundSupplyInterestFactorPrior := suite.keeper.GetSupplyInterestFactor(prevCtx, tc.args.borrowCoinDenom)
				suite.Require().True(foundSupplyInterestFactorPrior)
				fmt.Printf("[TEST] Supply Interest Factor Prior: %s\n", supplyInterestFactorPrior)

				// 2. Calculate expected borrow interest owed
				borrowRateApy, err := hard.CalculateBorrowRate(tc.args.interestRateModel, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowCoinPriorAmount), sdk.NewDecFromInt(reservesPrior.Amount))
				suite.Require().NoError(err)
				fmt.Printf("[TEST] Borrow Rate APY: %s\n", borrowRateApy)

				// Convert from APY to SPY, expressed as (1 + borrow rate)
				borrowRateSpy, err := hard.APYToSPY(sdk.OneDec().Add(borrowRateApy))
				fmt.Printf("[TEST] Borrow Rate SPY: %s\n", borrowRateSpy)
				suite.Require().NoError(err)

				borrowInterestFactor := hard.CalculateBorrowInterestFactor(borrowRateSpy, sdk.NewInt(snapshot.elapsedTime))
				fmt.Printf("[TEST] Borrow Interest Factor: %s\n", borrowInterestFactor)
				expectedBorrowInterest := (borrowInterestFactor.Mul(sdk.NewDecFromInt(borrowCoinPriorAmount)).TruncateInt()).Sub(borrowCoinPriorAmount)
				fmt.Printf("[TEST] Expected Borrow Interest: %s\n", expectedBorrowInterest)
				fmt.Printf("- [TEST] Brand New Reserves: %s\n", sdk.NewCoin(tc.args.borrowCoinDenom, sdk.NewDecFromInt(expectedBorrowInterest).Mul(tc.args.reserveFactor).TruncateInt()))
				expectedReserves := reservesPrior.Add(sdk.NewCoin(tc.args.borrowCoinDenom, sdk.NewDecFromInt(expectedBorrowInterest).Mul(tc.args.reserveFactor).TruncateInt()))
				fmt.Printf("- [TEST] Total Expected Reserves (prior + new): %s\n", expectedReserves)
				expectedBorrowInterestFactor := borrowInterestFactorPrior.Mul(borrowInterestFactor)

				// 3. Calculate expected supply interest owed
				borrowInterestFactorDiff := expectedBorrowInterestFactor.Sub(borrowInterestFactorPrior)
				fmt.Printf("[TEST] Borrow Interest Factor Diff: %s\n", borrowInterestFactorDiff)
				supplyInterestFactor := hard.CalculateSupplyInterestFactor(borrowInterestFactorDiff, cashPrior.ToDec(), sdk.NewDecFromInt(borrowCoinPriorAmount), sdk.NewDecFromInt(reservesPrior.Amount), tc.args.reserveFactor)
				fmt.Printf("[TEST] Supply Interest Factor: %s\n", supplyInterestFactor)
				expectedSupplyInterest := (supplyInterestFactor.Mul(sdk.NewDecFromInt(supplyCoinPriorAmount)).TruncateInt()).Sub(supplyCoinPriorAmount)
				fmt.Printf("[TEST] Expected Supply Interest: %s\n", expectedSupplyInterest)
				expectedSupplyInterestFactor := supplyInterestFactorPrior.Mul(supplyInterestFactor)
				fmt.Printf("[TEST] Expected Supply Interest Factor: %s\n", expectedSupplyInterestFactor)
				// -------------------------------------------------------------------------------------

				// Set up snapshot chain context and run begin blocker
				runAtTime := time.Unix(prevCtx.BlockTime().Unix()+(snapshot.elapsedTime), 0)
				snapshotCtx := prevCtx.WithBlockTime(runAtTime)
				hard.BeginBlocker(snapshotCtx, suite.keeper)

				reserveFactorAmount := tc.args.reserveFactor.MulInt(expectedBorrowInterest).TruncateInt()
				fmt.Printf("- [TEST] Reserve Factor Amount: %s\n", reserveFactorAmount)
				fmt.Printf("- [TEST] Expected Supply Interest: %s\n", expectedSupplyInterest)
				fmt.Printf("- [TEST] Expected Supply + Reserve Factor Amount: %s\n", expectedSupplyInterest.Add(reserveFactorAmount))
				fmt.Printf("- [TEST] Expected Borrow Interest: %s\n", expectedBorrowInterest)
				suite.Require().Equal(expectedBorrowInterest, expectedSupplyInterest.Add(reserveFactorAmount))

				// Check that the total amount of borrowed coins has increased by expected borrow interest amount
				borrowCoinsPost, _ := suite.keeper.GetBorrowedCoins(snapshotCtx)
				borrowCoinPostAmount := borrowCoinsPost.AmountOf(tc.args.borrowCoinDenom)
				suite.Require().Equal(borrowCoinPostAmount, borrowCoinPriorAmount.Add(expectedBorrowInterest))

				// Check that the total amount of supplied coins has increased by expected supply interest amount
				supplyCoinsPost, _ := suite.keeper.GetSuppliedCoins(snapshotCtx)
				supplyCoinPostAmount := supplyCoinsPost.AmountOf(tc.args.borrowCoinDenom)
				suite.Require().Equal(supplyCoinPostAmount, supplyCoinPriorAmount.Add(expectedSupplyInterest))

				// Check current total reserves
				currTotalReserves, _ := suite.keeper.GetTotalReserves(snapshotCtx, tc.args.borrowCoinDenom)
				suite.Require().Equal(expectedReserves, currTotalReserves)

				// Check that the supply index has increased as expected
				currSupplyIndexPrior, _ := suite.keeper.GetSupplyInterestFactor(snapshotCtx, tc.args.borrowCoinDenom)
				suite.Require().Equal(expectedSupplyInterestFactor, currSupplyIndexPrior)

				// Check that the borrow index has increased as expected
				currBorrowIndexPrior, _ := suite.keeper.GetBorrowInterestFactor(snapshotCtx, tc.args.borrowCoinDenom)
				suite.Require().Equal(expectedBorrowInterestFactor, currBorrowIndexPrior)

				prevCtx = snapshotCtx
			}
		})
	}
}

func TestInterestTestSuite(t *testing.T) {
	suite.Run(t, new(InterestTestSuite))
}
