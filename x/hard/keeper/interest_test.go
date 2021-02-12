package keeper_test

import (
	"strconv"
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
		{
			"1 year: highest interest rate",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("1.000001555555"),
				timeElapsed:           sdk.NewInt(oneYearInSeconds),
				expectedValue:         sdk.MustNewDecFromStr("2017093013158200407564.613502861572552603"),
			},
		},
		{
			"largest per second interest rate with practical elapsed time",
			args{
				perSecondInterestRate: sdk.MustNewDecFromStr("18.445"), // Begins to panic at ~18.45 (1845%/second interest rate)
				timeElapsed:           sdk.NewInt(30),                  // Assume a 30 second period, longer than any expected individual block
				expectedValue:         sdk.MustNewDecFromStr("94702138679846565921082258202543002089.215969366091911769"),
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
		newInterest   sdk.Dec
		cash          sdk.Dec
		borrows       sdk.Dec
		reserves      sdk.Dec
		reserveFactor sdk.Dec
		expectedValue sdk.Dec
	}

	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"low new interest",
			args{
				newInterest:   sdk.MustNewDecFromStr("1"),
				cash:          sdk.MustNewDecFromStr("100.0"),
				borrows:       sdk.MustNewDecFromStr("1000.0"),
				reserves:      sdk.MustNewDecFromStr("10.0"),
				reserveFactor: sdk.MustNewDecFromStr("0.05"),
				expectedValue: sdk.MustNewDecFromStr("1.000917431192660550"),
			},
		},
		{
			"medium new interest",
			args{
				newInterest:   sdk.MustNewDecFromStr("5"),
				cash:          sdk.MustNewDecFromStr("100.0"),
				borrows:       sdk.MustNewDecFromStr("1000.0"),
				reserves:      sdk.MustNewDecFromStr("10.0"),
				reserveFactor: sdk.MustNewDecFromStr("0.05"),
				expectedValue: sdk.MustNewDecFromStr("1.004587155963302752"),
			},
		},
		{
			"high new interest",
			args{
				newInterest:   sdk.MustNewDecFromStr("10"),
				cash:          sdk.MustNewDecFromStr("100.0"),
				borrows:       sdk.MustNewDecFromStr("1000.0"),
				reserves:      sdk.MustNewDecFromStr("10.0"),
				reserveFactor: sdk.MustNewDecFromStr("0.05"),
				expectedValue: sdk.MustNewDecFromStr("1.009174311926605505"),
			},
		},
	}

	for _, tc := range testCases {
		interestFactor := hard.CalculateSupplyInterestFactor(tc.args.newInterest,
			tc.args.cash, tc.args.borrows, tc.args.reserves)
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

func (suite *InterestTestSuite) TestSPYToEstimatedAPY() {
	type args struct {
		spy             sdk.Dec
		expectedAPY     float64
		acceptableRange float64
	}

	type test struct {
		name string
		args args
	}

	testCases := []test{
		{
			"lowest apy",
			args{
				spy:             sdk.MustNewDecFromStr("0.999999831991472557"),
				expectedAPY:     0.005,   // Returned value: 0.004999999888241291
				acceptableRange: 0.00001, // +/- 1/10000th of a precent
			},
		},
		{
			"lower apy",
			args{
				spy:             sdk.MustNewDecFromStr("0.999999905005957279"),
				expectedAPY:     0.05,    // Returned value: 0.05000000074505806
				acceptableRange: 0.00001, // +/- 1/10000th of a precent
			},
		},
		{
			"medium-low apy",
			args{
				spy:             sdk.MustNewDecFromStr("0.999999978020447332"),
				expectedAPY:     0.5,     // Returned value: 0.5
				acceptableRange: 0.00001, // +/- 1/10000th of a precent
			},
		},
		{
			"medium-high apy",
			args{
				spy:             sdk.MustNewDecFromStr("1.000000051034942717"),
				expectedAPY:     5,       // Returned value: 5
				acceptableRange: 0.00001, // +/- 1/10000th of a precent
			},
		},
		{
			"high apy",
			args{
				spy:             sdk.MustNewDecFromStr("1.000000124049443433"),
				expectedAPY:     50,      // Returned value: 50
				acceptableRange: 0.00001, // +/- 1/10000th of a precent
			},
		},
		{
			"highest apy",
			args{
				spy:             sdk.MustNewDecFromStr("1.000000146028999310"),
				expectedAPY:     100,     // 100
				acceptableRange: 0.00001, // +/- 1/10000th of a precent
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// From SPY calculate APY and parse result from sdk.Dec to float64
			calculatedAPY := hard.SPYToEstimatedAPY(tc.args.spy)
			calculatedAPYFloat, err := strconv.ParseFloat(calculatedAPY.String(), 32)
			suite.Require().NoError(err)

			// Check that the calculated value is within an acceptable percentage range
			suite.Require().InEpsilon(tc.args.expectedAPY, calculatedAPYFloat, tc.args.acceptableRange)
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
				types.MoneyMarkets{
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",                // Market ID
						sdk.NewInt(KAVA_CF),       // Conversion Factor
						tc.args.interestRateModel, // Interest Rate Model
						tc.args.reserveFactor,     // Reserve Factor
						sdk.ZeroDec()),            // Keeper Reward Percentage
				},
				sdk.NewDec(10),
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)

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

				reservesPrior, foundReservesPrior := suite.keeper.GetTotalReserves(prevCtx)
				if !foundReservesPrior {
					reservesPrior = sdk.NewCoins(sdk.NewCoin(tc.args.borrowCoinDenom, sdk.ZeroInt()))
				}

				interestFactorPrior, foundInterestFactorPrior := suite.keeper.GetBorrowInterestFactor(prevCtx, tc.args.borrowCoinDenom)
				suite.Require().True(foundInterestFactorPrior)

				// 2. Calculate expected interest owed
				borrowRateApy, err := hard.CalculateBorrowRate(tc.args.interestRateModel, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowCoinPriorAmount), sdk.NewDecFromInt(reservesPrior.AmountOf(tc.args.borrowCoinDenom)))
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
				currTotalReserves, _ := suite.keeper.GetTotalReserves(snapshotCtx)
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
	shouldSupply bool
	supplyCoin   sdk.Coin
}

func (suite *KeeperTestSuite) TestSupplyInterest() {
	type args struct {
		user                     sdk.AccAddress
		initialSupplierCoins     sdk.Coins
		initialBorrowerCoins     sdk.Coins
		initialModuleCoins       sdk.Coins
		depositCoins             sdk.Coins
		coinDenoms               []string
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
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneDayInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
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
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneWeekInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
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
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
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
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneYearInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"supply/borrow multiple coins",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(100*BNB_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(100*BNB_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF)), sdk.NewCoin("bnb", sdk.NewInt(20*BNB_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
					},
				},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"supply during snapshot",
			args{
				user:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialBorrowerCoins: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				initialModuleCoins:   sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000*KAVA_CF))),
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldSupply: true,
						supplyCoin:   sdk.NewCoin("ukava", sdk.NewInt(20*KAVA_CF)),
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
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(80*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneMonthInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneMonthInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
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
				depositCoins:         sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100*KAVA_CF))),
				coinDenoms:           []string{"ukava"},
				borrowCoins:          sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(50*KAVA_CF))),
				interestRateModel:    normalModel,
				reserveFactor:        sdk.MustNewDecFromStr("0.05"),
				expectedInterestSnaphots: []ExpectedSupplyInterest{
					{
						elapsedTime:  oneMonthInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneDayInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneYearInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
					},
					{
						elapsedTime:  oneWeekInSeconds,
						shouldSupply: false,
						supplyCoin:   sdk.Coin{},
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
				types.MoneyMarkets{
					types.NewMoneyMarket("ukava",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"kava:usd",                // Market ID
						sdk.NewInt(KAVA_CF),       // Conversion Factor
						tc.args.interestRateModel, // Interest Rate Model
						tc.args.reserveFactor,     // Reserve Factor
						sdk.ZeroDec()),            // Keeper Reward Percentage
					types.NewMoneyMarket("bnb",
						types.NewBorrowLimit(false, sdk.NewDec(100000000*BNB_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
						"bnb:usd",                 // Market ID
						sdk.NewInt(BNB_CF),        // Conversion Factor
						tc.args.interestRateModel, // Interest Rate Model
						tc.args.reserveFactor,     // Reserve Factor
						sdk.ZeroDec()),            // Keeper Reward Percentage
				},
				sdk.NewDec(10),
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)

			// Pricefeed module genesis state
			pricefeedGS := pricefeed.GenesisState{
				Params: pricefeed.Params{
					Markets: []pricefeed.Market{
						{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
						{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
					},
				},
				PostedPrices: []pricefeed.PostedPrice{
					{
						MarketID:      "kava:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("2.00"),
						Expiry:        time.Now().Add(100 * time.Hour),
					},
					{
						MarketID:      "bnb:usd",
						OracleAddress: sdk.AccAddress{},
						Price:         sdk.MustNewDecFromStr("20.00"),
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
			for _, snapshot := range tc.args.expectedInterestSnaphots {
				for _, coinDenom := range tc.args.coinDenoms {
					// ---------------------------- Calculate expected supply interest ----------------------------
					// 1. Get cash, borrows, reserves, and borrow index
					cashPrior := suite.getModuleAccountAtCtx(types.ModuleName, prevCtx).GetCoins().AmountOf(coinDenom)

					var borrowCoinPriorAmount sdk.Int
					borrowCoinsPrior, borrowCoinsPriorFound := suite.keeper.GetBorrowedCoins(prevCtx)
					suite.Require().True(borrowCoinsPriorFound)
					borrowCoinPriorAmount = borrowCoinsPrior.AmountOf(coinDenom)

					var supplyCoinPriorAmount sdk.Int
					supplyCoinsPrior, supplyCoinsPriorFound := suite.keeper.GetSuppliedCoins(prevCtx)
					suite.Require().True(supplyCoinsPriorFound)
					supplyCoinPriorAmount = supplyCoinsPrior.AmountOf(coinDenom)

					reservesPrior, foundReservesPrior := suite.keeper.GetTotalReserves(prevCtx)
					if !foundReservesPrior {
						reservesPrior = sdk.NewCoins(sdk.NewCoin(coinDenom, sdk.ZeroInt()))
					}

					borrowInterestFactorPrior, foundBorrowInterestFactorPrior := suite.keeper.GetBorrowInterestFactor(prevCtx, coinDenom)
					suite.Require().True(foundBorrowInterestFactorPrior)

					supplyInterestFactorPrior, foundSupplyInterestFactorPrior := suite.keeper.GetSupplyInterestFactor(prevCtx, coinDenom)
					suite.Require().True(foundSupplyInterestFactorPrior)

					// 2. Calculate expected borrow interest owed
					borrowRateApy, err := hard.CalculateBorrowRate(tc.args.interestRateModel, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowCoinPriorAmount), sdk.NewDecFromInt(reservesPrior.AmountOf(coinDenom)))
					suite.Require().NoError(err)

					// Convert from APY to SPY, expressed as (1 + borrow rate)
					borrowRateSpy, err := hard.APYToSPY(sdk.OneDec().Add(borrowRateApy))
					suite.Require().NoError(err)

					newBorrowInterestFactor := hard.CalculateBorrowInterestFactor(borrowRateSpy, sdk.NewInt(snapshot.elapsedTime))
					expectedBorrowInterest := (newBorrowInterestFactor.Mul(sdk.NewDecFromInt(borrowCoinPriorAmount)).TruncateInt()).Sub(borrowCoinPriorAmount)
					expectedReserves := reservesPrior.Add(sdk.NewCoin(coinDenom, sdk.NewDecFromInt(expectedBorrowInterest).Mul(tc.args.reserveFactor).TruncateInt())).Sub(reservesPrior)
					expectedTotalReserves := expectedReserves.Add(reservesPrior...)

					expectedBorrowInterestFactor := borrowInterestFactorPrior.Mul(newBorrowInterestFactor)
					expectedSupplyInterest := expectedBorrowInterest.Sub(expectedReserves.AmountOf(coinDenom))

					newSupplyInterestFactor := hard.CalculateSupplyInterestFactor(expectedSupplyInterest.ToDec(), sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowCoinPriorAmount), sdk.NewDecFromInt(reservesPrior.AmountOf(coinDenom)))
					expectedSupplyInterestFactor := supplyInterestFactorPrior.Mul(newSupplyInterestFactor)
					// -------------------------------------------------------------------------------------

					// Set up snapshot chain context and run begin blocker
					runAtTime := time.Unix(prevCtx.BlockTime().Unix()+(snapshot.elapsedTime), 0)
					snapshotCtx := prevCtx.WithBlockTime(runAtTime)
					hard.BeginBlocker(snapshotCtx, suite.keeper)

					borrowInterestFactor, _ := suite.keeper.GetBorrowInterestFactor(ctx, coinDenom)
					suite.Require().Equal(expectedBorrowInterestFactor, borrowInterestFactor)
					suite.Require().Equal(expectedBorrowInterest, expectedSupplyInterest.Add(expectedReserves.AmountOf(coinDenom)))

					// Check that the total amount of borrowed coins has increased by expected borrow interest amount
					borrowCoinsPost, _ := suite.keeper.GetBorrowedCoins(snapshotCtx)
					borrowCoinPostAmount := borrowCoinsPost.AmountOf(coinDenom)
					suite.Require().Equal(borrowCoinPostAmount, borrowCoinPriorAmount.Add(expectedBorrowInterest))

					// Check that the total amount of supplied coins has increased by expected supply interest amount
					supplyCoinsPost, _ := suite.keeper.GetSuppliedCoins(prevCtx)
					supplyCoinPostAmount := supplyCoinsPost.AmountOf(coinDenom)
					suite.Require().Equal(supplyCoinPostAmount, supplyCoinPriorAmount.Add(expectedSupplyInterest))

					// Check current total reserves
					totalReserves, _ := suite.keeper.GetTotalReserves(snapshotCtx)
					suite.Require().Equal(
						sdk.NewCoin(coinDenom, expectedTotalReserves.AmountOf(coinDenom)),
						sdk.NewCoin(coinDenom, totalReserves.AmountOf(coinDenom)),
					)

					// Check that the supply index has increased as expected
					currSupplyIndexPrior, _ := suite.keeper.GetSupplyInterestFactor(snapshotCtx, coinDenom)
					suite.Require().Equal(expectedSupplyInterestFactor, currSupplyIndexPrior)

					// // Check that the borrow index has increased as expected
					currBorrowIndexPrior, _ := suite.keeper.GetBorrowInterestFactor(snapshotCtx, coinDenom)
					suite.Require().Equal(expectedBorrowInterestFactor, currBorrowIndexPrior)

					// After supplying again user's supplied balance should have owed supply interest applied
					if snapshot.shouldSupply {
						// Calculate percentage of supply interest profits owed to user
						userSupplyBefore, _ := suite.keeper.GetDeposit(snapshotCtx, tc.args.user)
						userSupplyCoinAmount := userSupplyBefore.Amount.AmountOf(coinDenom)
						userPercentOfTotalSupplied := userSupplyCoinAmount.ToDec().Quo(supplyCoinPriorAmount.ToDec())
						userExpectedSupplyInterestCoin := sdk.NewCoin(coinDenom, userPercentOfTotalSupplied.MulInt(expectedSupplyInterest).TruncateInt())

						// Calculate percentage of borrow interest profits owed to user
						userBorrowBefore, _ := suite.keeper.GetBorrow(snapshotCtx, tc.args.user)
						userBorrowCoinAmount := userBorrowBefore.Amount.AmountOf(coinDenom)
						userPercentOfTotalBorrowed := userBorrowCoinAmount.ToDec().Quo(borrowCoinPriorAmount.ToDec())
						userExpectedBorrowInterestCoin := sdk.NewCoin(coinDenom, userPercentOfTotalBorrowed.MulInt(expectedBorrowInterest).TruncateInt())
						expectedBorrowCoinsAfter := userBorrowBefore.Amount.Add(userExpectedBorrowInterestCoin)

						// Supplying syncs user's owed supply and borrow interest
						err = suite.keeper.Deposit(snapshotCtx, tc.args.user, sdk.NewCoins(snapshot.supplyCoin))
						suite.Require().NoError(err)

						// Fetch user's new borrow and supply balance post-interaction
						userSupplyAfter, _ := suite.keeper.GetDeposit(snapshotCtx, tc.args.user)
						userBorrowAfter, _ := suite.keeper.GetBorrow(snapshotCtx, tc.args.user)

						// Confirm that user's supply index for the denom has increased as expected
						var userSupplyAfterIndexFactor sdk.Dec
						for _, indexFactor := range userSupplyAfter.Index {
							if indexFactor.Denom == coinDenom {
								userSupplyAfterIndexFactor = indexFactor.Value
							}
						}
						suite.Require().Equal(userSupplyAfterIndexFactor, currSupplyIndexPrior)

						// Check user's supplied amount increased by supply interest owed + the newly supplied coins
						expectedSupplyCoinsAfter := userSupplyBefore.Amount.Add(snapshot.supplyCoin).Add(userExpectedSupplyInterestCoin)
						suite.Require().Equal(expectedSupplyCoinsAfter, userSupplyAfter.Amount)

						// Confirm that user's borrow index for the denom has increased as expected
						var userBorrowAfterIndexFactor sdk.Dec
						for _, indexFactor := range userBorrowAfter.Index {
							if indexFactor.Denom == coinDenom {
								userBorrowAfterIndexFactor = indexFactor.Value
							}
						}
						suite.Require().Equal(userBorrowAfterIndexFactor, currBorrowIndexPrior)

						// Check user's borrowed amount increased by borrow interest owed
						suite.Require().Equal(expectedBorrowCoinsAfter, userBorrowAfter.Amount)
					}
					prevCtx = snapshotCtx
				}
			}
		})
	}
}

func TestInterestTestSuite(t *testing.T) {
	suite.Run(t, new(InterestTestSuite))
}
