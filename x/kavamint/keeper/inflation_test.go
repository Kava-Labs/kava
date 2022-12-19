package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavamint/keeper"
	"github.com/kava-labs/kava/x/kavamint/types"
	"github.com/stretchr/testify/suite"
)

type InflationTestSuite struct {
	suite.Suite
}

func TestInflationTestSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) TestCalculateInflationFactor() {
	testCases := []struct {
		name          string
		apy           sdk.Dec
		secondsPassed uint64
		expectedRate  sdk.Dec
		// preciseToDecimal is the decimal place after which error is present. if >=18, number is exact.
		// ex. precise to 3 decimals means accurate to +/-0.0005
		preciseToDecimal int
	}{
		{
			name:             "any apy over 0 seconds is 0",
			apy:              sdk.OneDec(),
			secondsPassed:    0,
			expectedRate:     sdk.ZeroDec(),
			preciseToDecimal: 19,
		},
		{
			name:             "zero apy for any number of seconds is 0",
			apy:              sdk.ZeroDec(),
			secondsPassed:    100,
			expectedRate:     sdk.ZeroDec(),
			preciseToDecimal: 19,
		},
		{
			name:             "an apy over a year's worth of seconds should be the apy",
			apy:              sdk.NewDecWithPrec(10, 2),
			secondsPassed:    keeper.SecondsPerYear,
			expectedRate:     sdk.NewDecWithPrec(10, 2),
			preciseToDecimal: 10,
		},
		{
			name:             "example: 22 percent for 6 seconds, precise to 17 decimals",
			apy:              sdk.NewDecWithPrec(22, 2),
			secondsPassed:    6,
			expectedRate:     sdk.MustNewDecFromStr("0.000000037833116915"),
			preciseToDecimal: 17,
		},
		{
			name:             "example: 3 percent for 10 seconds, precise to 17 decimals",
			apy:              sdk.NewDecWithPrec(3, 2),
			secondsPassed:    10,
			expectedRate:     sdk.MustNewDecFromStr("0.000000009373034748"),
			preciseToDecimal: 17,
		},
		{
			name:             "example: 150 percent for 10 seconds, precise to 17 decimals",
			apy:              sdk.NewDecWithPrec(150, 2),
			secondsPassed:    10,
			expectedRate:     sdk.MustNewDecFromStr("0.000000290553927255"),
			preciseToDecimal: 17,
		},
		{
			name:             "example: 10 percent for 10 seconds, precise to 16 decimals",
			apy:              sdk.NewDecWithPrec(10, 2),
			secondsPassed:    10,
			expectedRate:     sdk.MustNewDecFromStr("0.000000030222660212"),
			preciseToDecimal: 16,
		},
		{
			name:             "example: 1,000 percent for 10 seconds, precise to 16 decimals",
			apy:              sdk.NewDecWithPrec(1000, 2),
			secondsPassed:    10,
			expectedRate:     sdk.MustNewDecFromStr("0.000000760367892072"),
			preciseToDecimal: 16,
		},
		{
			name:             "example: 10 percent for 30 seconds, precise to 16 decimals",
			apy:              sdk.NewDecWithPrec(10, 2),
			secondsPassed:    30,
			expectedRate:     sdk.MustNewDecFromStr("0.00000009066798338"),
			preciseToDecimal: 16,
		},
		{
			name:             "example: 95 percent for 15 seconds, precise to 16 decimals",
			apy:              sdk.NewDecWithPrec(95, 2),
			secondsPassed:    15,
			expectedRate:     sdk.MustNewDecFromStr("0.000000317651007726"),
			preciseToDecimal: 16,
		},
		{
			name:             "example: 10,000 percent for 10 seconds, precise to 13 decimals",
			apy:              sdk.NewDecWithPrec(10000, 2),
			secondsPassed:    10,
			expectedRate:     sdk.MustNewDecFromStr("0.000001463446186527"),
			preciseToDecimal: 13,
		},
		{
			name:             "can handle upper bound of APY (but w/ large error)",
			apy:              types.MaxMintingRate,
			secondsPassed:    10,
			expectedRate:     sdk.MustNewDecFromStr("0.000001642242155465"),
			preciseToDecimal: 4, // NOTE: error is really large here
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			minter := keeper.NewMinter(tc.name, tc.apy, sdk.ZeroInt(), "ukava", "ignored")
			actualRate, err := minter.CalculateInflationRate(tc.secondsPassed)
			suite.Require().NoError(err)

			marginOfError := sdk.ZeroDec()
			if tc.preciseToDecimal < 18 {
				marginOfError = sdk.NewDecWithPrec(5, int64(tc.preciseToDecimal+1))
			}
			suite.requireWithinError(tc.expectedRate, actualRate, marginOfError)
		})
	}

	suite.Run("errors with out-of-bounds when rate is too high", func() {
		// Dec.ApproxRoot will error w/ out-of-bounds when rate is >176.5
		oob := sdk.NewDec(177)
		minter := keeper.NewMinter(
			"out-of-bounds-minter",
			oob,
			sdk.OneInt(),
			"uakva",
			"ignored",
		)
		_, err := minter.CalculateInflationRate(100)
		suite.Error(err)

		// ensure max mint rate is less than out-of-bounds value
		suite.True(types.MaxMintingRate.LT(oob))
	})
}

func (suite *InflationTestSuite) Test_AccumulateInflation_FailsWithInvalidRate() {
	oob := sdk.NewDec(177)
	minter := keeper.NewMinter(
		"out-of-bounds-minter",
		oob,
		sdk.OneInt(),
		"uakva",
		"ignored",
	)
	_, err := minter.AccumulateInflation(1)
	suite.ErrorContains(err, "out of bounds")
}

func (suite *InflationTestSuite) requireWithinError(expected, actual, margin sdk.Dec) {
	suite.Require().Truef(
		actual.Sub(expected).Abs().LTE(margin),
		fmt.Sprintf("precision is outside desired margin of error %s\nexpected: %s\nactual  : %s", margin, expected, actual),
	)
}
