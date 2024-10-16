package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/hard/types"
)

type ParamTestSuite struct {
	suite.Suite
}

func (suite *ParamTestSuite) TestParamValidation() {
	type args struct {
		minBorrowVal sdkmath.LegacyDec
		mms          types.MoneyMarkets
	}
	testCases := []struct {
		name        string
		args        args
		expectPass  bool
		expectedErr string
	}{
		{
			name: "default",
			args: args{
				minBorrowVal: types.DefaultMinimumBorrowUSDValue,
				mms:          types.DefaultMoneyMarkets,
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "invalid: conversion factor < one",
			args: args{
				minBorrowVal: types.DefaultMinimumBorrowUSDValue,
				mms: types.MoneyMarkets{
					{
						Denom: "btcb",
						BorrowLimit: types.NewBorrowLimit(
							false,
							sdkmath.LegacyMustNewDecFromStr("100000000000"),
							sdkmath.LegacyMustNewDecFromStr("0.5"),
						),
						SpotMarketID:           "btc:usd",
						ConversionFactor:       sdkmath.NewInt(0),
						InterestRateModel:      types.InterestRateModel{},
						ReserveFactor:          sdkmath.LegacyMustNewDecFromStr("0.05"),
						KeeperRewardPercentage: sdkmath.LegacyMustNewDecFromStr("0.05"),
					},
				},
			},
			expectPass:  false,
			expectedErr: "conversion '0' factor must be ≥ one",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := types.NewParams(tc.args.mms, tc.args.minBorrowVal)
			err := params.Validate()
			if tc.expectPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
				suite.Require().Contains(err.Error(), tc.expectedErr)
			}
		})
	}
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}
