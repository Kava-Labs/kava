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
		allowed       AllowedCollateralParams
		current       cdptypes.CollateralParams
		incoming      cdptypes.CollateralParams
		expectAllowed bool
	}{
		{
			name: "disallowed add CP",
			allowed: AllowedCollateralParams{
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
			current:       testCPs[:2],
			incoming:      testCPs[:3],
			expectAllowed: false,
		},
		{
			name: "disallowed remove CP",
			allowed: AllowedCollateralParams{
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
			current:       testCPs[:2],
			incoming:      testCPs[:1], // removes btc
			expectAllowed: false,
		},
		{
			name: "allowed change with different order",
			allowed: AllowedCollateralParams{
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
			current:       testCPs[:3],
			incoming:      testCPs[:3],
			expectAllowed: true,
		},
	}
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(
				tc.expectAllowed,
				tc.allowed.Allows(tc.current, tc.incoming),
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
		allowed       AllowedCollateralParam
		current       cdptypes.CollateralParam
		incoming      cdptypes.CollateralParam
		expectAllowed bool
	}{
		{
			name: "allowed change",
			allowed: AllowedCollateralParam{
				Denom:        "bnb",
				DebtLimit:    true,
				StabilityFee: true,
				AuctionSize:  true,
			},
			current:       testCP,
			incoming:      newDebtLimitCP,
			expectAllowed: true,
		},
		{
			name: "un-allowed change",
			allowed: AllowedCollateralParam{
				Denom:        "bnb",
				DebtLimit:    true,
				StabilityFee: true,
				AuctionSize:  true,
			},
			current:       testCP,
			incoming:      newMarketIDCP,
			expectAllowed: false,
		},
		{
			name: "un-allowed mismatching denom",
			allowed: AllowedCollateralParam{
				Denom:     "btc",
				DebtLimit: true,
			},
			current:       testCP,
			incoming:      newDebtLimitCP,
			expectAllowed: false,
		},

		{
			name: "allowed no change",
			allowed: AllowedCollateralParam{
				Denom:     "bnb",
				DebtLimit: true,
			},
			current:       testCP,
			incoming:      testCP, // no change
			expectAllowed: true,
		},
		{
			name: "un-allowed change with allowed change",
			allowed: AllowedCollateralParam{
				Denom:     "btc",
				DebtLimit: true,
			},
			current:       testCP,
			incoming:      newMarketIDAndDebtLimitCP,
			expectAllowed: false,
		},
		// TODO {
		// 	name: "nil Int values",
		// 	allowed: AllowedCollateralParam{
		// 		Denom:     "btc",
		// 		DebtLimit: true,
		// 	},
		// 	incoming:    cdptypes.CollateralParam{}, // nil sdk.Int types
		// 	current:     testCP,
		// 	expectAllowed: false,
		// },
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(
				tc.expectAllowed,
				tc.allowed.Allows(tc.current, tc.incoming),
			)
		})
	}
}

func (suite *PermissionsTestSuite) TestAllowedDebtParam_Allows() {
	testDP := cdptypes.DebtParam{
		Denom:            "usdx",
		ReferenceAsset:   "usd",
		ConversionFactor: i(6),
		DebtFloor:        i(10000000),
		SavingsRate:      d("0.95"),
	}
	newDenomDP := testDP
	newDenomDP.Denom = "usdz"

	newDebtFloorDP := testDP
	newDebtFloorDP.DebtFloor = i(1000)

	newDenomAndDebtFloorDP := testDP
	newDenomAndDebtFloorDP.Denom = "usdz"
	newDenomAndDebtFloorDP.DebtFloor = i(1000)

	testcases := []struct {
		name          string
		allowed       AllowedDebtParam
		current       cdptypes.DebtParam
		incoming      cdptypes.DebtParam
		expectAllowed bool
	}{
		{
			name: "allowed change",
			allowed: AllowedDebtParam{
				DebtFloor:   true,
				SavingsRate: true,
			},
			current:       testDP,
			incoming:      newDebtFloorDP,
			expectAllowed: true,
		},
		{
			name: "un-allowed change",
			allowed: AllowedDebtParam{
				DebtFloor:   true,
				SavingsRate: true,
			},
			current:       testDP,
			incoming:      newDenomDP,
			expectAllowed: false,
		},
		{
			name: "allowed no change",
			allowed: AllowedDebtParam{
				DebtFloor:   true,
				SavingsRate: true,
			},
			current:       testDP,
			incoming:      testDP, // no change
			expectAllowed: true,
		},
		{
			name: "un-allowed change with allowed change",
			allowed: AllowedDebtParam{
				DebtFloor:   true,
				SavingsRate: true,
			},
			current:       testDP,
			incoming:      newDenomAndDebtFloorDP,
			expectAllowed: false,
		},
		// TODO {
		// 	name: "nil Int values",
		// 	allowed: AllowedCollateralParam{
		// 		Denom:     "btc",
		// 		DebtLimit: true,
		// 	},
		// 	incoming:    cdptypes.CollateralParam{}, // nil sdk.Int types
		// 	current:     testCP,
		// 	expectAllowed: false,
		// },
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.Require().Equal(
				tc.expectAllowed,
				tc.allowed.Allows(tc.current, tc.incoming),
			)
		})
	}
}
