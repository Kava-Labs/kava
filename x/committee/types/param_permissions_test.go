package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func (suite *PermissionsTestSuite) TestAllowedCollateralParams_Allows() {
	testCPs := cdptypes.CollateralParams{
		{
			Denom:              "bnb",
			LiquidationRatio:   d("2.0"),
			DebtLimit:          c("usdx", 1000000000000),
			StabilityFee:       d("1.000000001547125958"),
			LiquidationPenalty: d("0.05"),
			AuctionSize:        i(100),
			Prefix:             0x20,
			ConversionFactor:   i(6),
			MarketID:           "bnb:usd",
		},
		{
			Denom:              "btc",
			LiquidationRatio:   d("1.5"),
			DebtLimit:          c("usdx", 1000000000),
			StabilityFee:       d("1.000000001547125958"),
			LiquidationPenalty: d("0.10"),
			AuctionSize:        i(1000),
			Prefix:             0x30,
			ConversionFactor:   i(8),
			MarketID:           "btc:usd",
		},
		{
			Denom:              "atom",
			LiquidationRatio:   d("2.0"),
			DebtLimit:          c("usdx", 1000000000),
			StabilityFee:       d("1.000000001547125958"),
			LiquidationPenalty: d("0.07"),
			AuctionSize:        i(100),
			Prefix:             0x40,
			ConversionFactor:   i(6),
			MarketID:           "atom:usd",
		},
	}
	updatedTestCPs := make(cdptypes.CollateralParams, len(testCPs))
	updatedTestCPs[0] = testCPs[1]
	updatedTestCPs[1] = testCPs[0]
	updatedTestCPs[2] = testCPs[2]

	updatedTestCPs[0].DebtLimit = c("usdx", 1000)
	updatedTestCPs[1].LiquidationPenalty = d("0.15")
	updatedTestCPs[2].DebtLimit = c("usdx", 1000)
	updatedTestCPs[2].LiquidationPenalty = d("0.15")

	testcases := []struct {
		name          string
		allowedCPs    AllowedCollateralParams
		currentCPs    cdptypes.CollateralParams
		incomingCPs   cdptypes.CollateralParams
		expectAllowed bool
	}{
		{
			name: "disallowed add CP",
			allowedCPs: AllowedCollateralParams{
				{
					Denom:       "bnb",
					AuctionSize: true,
				},
				{
					Denom:        "btc",
					StabilityFee: true,
				},
				{ // allow all fields
					Denom:              "atom",
					LiquidationRatio:   true,
					DebtLimit:          true,
					StabilityFee:       true,
					AuctionSize:        true,
					LiquidationPenalty: true,
					Prefix:             true,
					MarketID:           true,
					ConversionFactor:   true,
				},
			},
			currentCPs:    testCPs[:2],
			incomingCPs:   testCPs[:3],
			expectAllowed: false,
		},
		{
			name: "disallowed remove CP",
			allowedCPs: AllowedCollateralParams{
				{
					Denom:       "bnb",
					AuctionSize: true,
				},
				{
					// allow all fields
					Denom:              "btc",
					LiquidationRatio:   true,
					DebtLimit:          true,
					StabilityFee:       true,
					AuctionSize:        true,
					LiquidationPenalty: true,
					Prefix:             true,
					MarketID:           true,
					ConversionFactor:   true,
				},
			},
			currentCPs:    testCPs[:2],
			incomingCPs:   testCPs[:1], // removes btc
			expectAllowed: false,
		},
		{
			name: "allowed change with different order",
			allowedCPs: AllowedCollateralParams{
				{
					Denom:     "bnb",
					DebtLimit: true,
				},
				{
					Denom:              "btc",
					LiquidationPenalty: true,
				},
				{
					Denom:              "atom",
					DebtLimit:          true,
					LiquidationPenalty: true,
				},
			},
			currentCPs:    testCPs[:3],
			incomingCPs:   testCPs[:3],
			expectAllowed: true,
		},
	}
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(
				tc.expectAllowed,
				tc.allowedCPs.Allows(tc.currentCPs, tc.incomingCPs),
			)
		})
	}

}

func (suite *PermissionsTestSuite) TestAllowedCollateralParam_Allows() {
	testCP := cdptypes.CollateralParam{
		Denom:              "bnb",
		LiquidationRatio:   d("1.5"),
		DebtLimit:          c("usdx", 1000000000000),
		StabilityFee:       d("1.000000001547125958"), // %5 apr
		LiquidationPenalty: d("0.05"),
		AuctionSize:        i(100),
		Prefix:             0x20,
		ConversionFactor:   i(6),
		MarketID:           "bnb:usd",
	}
	newMarketIDCP := testCP
	newMarketIDCP.MarketID = "btc:usd"

	newDebtLimitCP := testCP
	newDebtLimitCP.DebtLimit = c("usdx", 1000)

	newMarketIDAndDebtLimitCP := testCP
	newMarketIDCP.MarketID = "btc:usd"
	newDebtLimitCP.DebtLimit = c("usdx", 1000)

	testcases := []struct {
		name          string
		allowedCP     AllowedCollateralParam
		currentCP     cdptypes.CollateralParam
		incomingCP    cdptypes.CollateralParam
		expectAllowed bool
	}{
		{
			name: "allowed change",
			allowedCP: AllowedCollateralParam{
				Denom:        "bnb",
				DebtLimit:    true,
				StabilityFee: true,
				AuctionSize:  true,
			},
			currentCP:     testCP,
			incomingCP:    newDebtLimitCP,
			expectAllowed: true,
		},
		{
			name: "un-allowed change",
			allowedCP: AllowedCollateralParam{
				Denom:        "bnb",
				DebtLimit:    true,
				StabilityFee: true,
				AuctionSize:  true,
			},
			currentCP:     testCP,
			incomingCP:    newMarketIDCP,
			expectAllowed: false,
		},
		{
			name: "un-allowed mismatching denom",
			allowedCP: AllowedCollateralParam{
				Denom:     "btc",
				DebtLimit: true,
			},
			currentCP:     testCP,
			incomingCP:    newDebtLimitCP,
			expectAllowed: false,
		},

		{
			name: "allowed no change",
			allowedCP: AllowedCollateralParam{
				Denom:     "bnb",
				DebtLimit: true,
			},
			currentCP:     testCP,
			incomingCP:    testCP, // no change
			expectAllowed: true,
		},
		{
			name: "un-allowed change with allowed change",
			allowedCP: AllowedCollateralParam{
				Denom:     "btc",
				DebtLimit: true,
			},
			currentCP:     testCP,
			incomingCP:    newMarketIDAndDebtLimitCP,
			expectAllowed: false,
		},
		// TODO {
		// 	name: "nil Int values",
		// 	allowedCP: AllowedCollateralParam{
		// 		Denom:     "btc",
		// 		DebtLimit: true,
		// 	},
		// 	incomingCP:    cdptypes.CollateralParam{}, // nil sdk.Int types
		// 	currentCP:     testCP,
		// 	expectAllowed: false,
		// },
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(
				tc.expectAllowed,
				tc.allowedCP.Allows(tc.currentCP, tc.incomingCP),
			)
		})
	}
}
