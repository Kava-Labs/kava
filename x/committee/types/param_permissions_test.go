package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bep3types "github.com/kava-labs/kava/x/bep3/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func (suite *PermissionsTestSuite) TestAllowedCollateralParams_Allows() {
	testCPs := cdptypes.CollateralParams{
		{
			Denom:               "bnb",
			LiquidationRatio:    d("2.0"),
			DebtLimit:           c("usdx", 1000000000000),
			StabilityFee:        d("1.000000001547125958"),
			LiquidationPenalty:  d("0.05"),
			AuctionSize:         i(100),
			Prefix:              0x20,
			ConversionFactor:    i(6),
			SpotMarketID:        "bnb:usd",
			LiquidationMarketID: "bnb:usd",
		},
		{
			Denom:               "btc",
			LiquidationRatio:    d("1.5"),
			DebtLimit:           c("usdx", 1000000000),
			StabilityFee:        d("1.000000001547125958"),
			LiquidationPenalty:  d("0.10"),
			AuctionSize:         i(1000),
			Prefix:              0x30,
			ConversionFactor:    i(8),
			SpotMarketID:        "btc:usd",
			LiquidationMarketID: "btc:usd",
		},
		{
			Denom:               "atom",
			LiquidationRatio:    d("2.0"),
			DebtLimit:           c("usdx", 1000000000),
			StabilityFee:        d("1.000000001547125958"),
			LiquidationPenalty:  d("0.07"),
			AuctionSize:         i(100),
			Prefix:              0x40,
			ConversionFactor:    i(6),
			SpotMarketID:        "atom:usd",
			LiquidationMarketID: "atom:usd",
		},
	}
	updatedTestCPs := make(cdptypes.CollateralParams, len(testCPs))
	updatedTestCPs[0] = testCPs[1]
	updatedTestCPs[1] = testCPs[0]
	updatedTestCPs[2] = testCPs[2]

	updatedTestCPs[0].DebtLimit = c("usdx", 1000)    // btc
	updatedTestCPs[1].LiquidationPenalty = d("0.15") // bnb
	updatedTestCPs[2].DebtLimit = c("usdx", 1000)    // atom
	updatedTestCPs[2].LiquidationPenalty = d("0.15") // atom

	testcases := []struct {
		name          string
		allowed       AllowedCollateralParams
		current       cdptypes.CollateralParams
		incoming      cdptypes.CollateralParams
		expectAllowed bool
	}{
		{
			name: "disallowed add",
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
					Denom:               "atom",
					LiquidationRatio:    true,
					DebtLimit:           true,
					StabilityFee:        true,
					AuctionSize:         true,
					LiquidationPenalty:  true,
					Prefix:              true,
					SpotMarketID:        true,
					LiquidationMarketID: true,
					ConversionFactor:    true,
				},
			},
			current:       testCPs[:2],
			incoming:      testCPs[:3],
			expectAllowed: false,
		},
		{
			name: "disallowed remove",
			allowed: AllowedCollateralParams{
				{
					Denom:       "bnb",
					AuctionSize: true,
				},
				{
					// allow all fields
					Denom:               "btc",
					LiquidationRatio:    true,
					DebtLimit:           true,
					StabilityFee:        true,
					AuctionSize:         true,
					LiquidationPenalty:  true,
					Prefix:              true,
					SpotMarketID:        true,
					LiquidationMarketID: true,
					ConversionFactor:    true,
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
					Denom:              "bnb",
					LiquidationPenalty: true,
				},
				{
					Denom:     "btc",
					DebtLimit: true,
				},
				{
					Denom:              "atom",
					DebtLimit:          true,
					LiquidationPenalty: true,
				},
			},
			current:       testCPs,
			incoming:      updatedTestCPs,
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

func (suite *PermissionsTestSuite) TestAllowedAssetParams_Allows() {
	testAPs := bep3types.AssetParams{
		{
			Denom:  "bnb",
			CoinID: 714,
			Limit:  i(1000000000000),
			Active: true,
		},
		{
			Denom:  "btc",
			CoinID: 0,
			Limit:  i(1000000000000),
			Active: true,
		},
		{
			Denom:  "xrp",
			CoinID: 144,
			Limit:  i(1000000000000),
			Active: true,
		},
	}
	updatedTestAPs := make(bep3types.AssetParams, len(testAPs))
	updatedTestAPs[0] = testAPs[1]
	updatedTestAPs[1] = testAPs[0]
	updatedTestAPs[2] = testAPs[2]

	updatedTestAPs[0].Limit = i(1000) // btc
	updatedTestAPs[1].Active = false  // bnb
	updatedTestAPs[2].Limit = i(1000) // xrp
	updatedTestAPs[2].Active = false  // xrp

	testcases := []struct {
		name          string
		allowed       AllowedAssetParams
		current       bep3types.AssetParams
		incoming      bep3types.AssetParams
		expectAllowed bool
	}{
		{
			name: "disallowed add",
			allowed: AllowedAssetParams{
				{
					Denom:  "bnb",
					Active: true,
				},
				{
					Denom: "btc",
					Limit: true,
				},
				{ // allow all fields
					Denom:  "xrp",
					CoinID: true,
					Limit:  true,
					Active: true,
				},
			},
			current:       testAPs[:2],
			incoming:      testAPs[:3],
			expectAllowed: false,
		},
		{
			name: "disallowed remove",
			allowed: AllowedAssetParams{
				{
					Denom:  "bnb",
					Active: true,
				},
				{ // allow all fields
					Denom:  "btc",
					CoinID: true,
					Limit:  true,
					Active: true,
				},
			},
			current:       testAPs[:2],
			incoming:      testAPs[:1], // removes btc
			expectAllowed: false,
		},
		{
			name: "allowed change with different order",
			allowed: AllowedAssetParams{
				{
					Denom:  "bnb",
					Active: true,
				},
				{
					Denom: "btc",
					Limit: true,
				},
				{
					Denom:  "xrp",
					Limit:  true,
					Active: true,
				},
			},
			current:       testAPs,
			incoming:      updatedTestAPs,
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

func (suite *PermissionsTestSuite) TestAllowedMarkets_Allows() {
	testMs := pricefeedtypes.Markets{
		{
			MarketID:   "bnb:usd",
			BaseAsset:  "bnb",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
		{
			MarketID:   "btc:usd",
			BaseAsset:  "btc",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
		{
			MarketID:   "atom:usd",
			BaseAsset:  "atom",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
	}
	updatedTestMs := make(pricefeedtypes.Markets, len(testMs))
	updatedTestMs[0] = testMs[1]
	updatedTestMs[1] = testMs[0]
	updatedTestMs[2] = testMs[2]

	updatedTestMs[0].Oracles = []sdk.AccAddress{[]byte("a test address")} // btc
	updatedTestMs[1].Active = false                                       // bnb
	updatedTestMs[2].Oracles = []sdk.AccAddress{[]byte("a test address")} // atom
	updatedTestMs[2].Active = false                                       // atom

	testcases := []struct {
		name          string
		allowed       AllowedMarkets
		current       pricefeedtypes.Markets
		incoming      pricefeedtypes.Markets
		expectAllowed bool
	}{
		{
			name: "disallowed add",
			allowed: AllowedMarkets{
				{
					MarketID: "bnb:usd",
					Active:   true,
				},
				{
					MarketID: "btc:usd",
					Oracles:  true,
				},
				{ // allow all fields
					MarketID:   "atom:usd",
					BaseAsset:  true,
					QuoteAsset: true,
					Oracles:    true,
					Active:     true,
				},
			},
			current:       testMs[:2],
			incoming:      testMs[:3],
			expectAllowed: false,
		},
		{
			name: "disallowed remove",
			allowed: AllowedMarkets{
				{
					MarketID: "bnb:usd",
					Active:   true,
				},
				{ // allow all fields
					MarketID:   "btc:usd",
					BaseAsset:  true,
					QuoteAsset: true,
					Oracles:    true,
					Active:     true,
				},
			},
			current:       testMs[:2],
			incoming:      testMs[:1], // removes btc
			expectAllowed: false,
		},
		{
			name: "allowed change with different order",
			allowed: AllowedMarkets{
				{
					MarketID: "bnb:usd",
					Active:   true,
				},
				{
					MarketID: "btc:usd",
					Oracles:  true,
				},
				{
					MarketID: "atom:usd",
					Oracles:  true,
					Active:   true,
				},
			},
			current:       testMs,
			incoming:      updatedTestMs,
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
		Denom:               "bnb",
		LiquidationRatio:    d("1.5"),
		DebtLimit:           c("usdx", 1000000000000),
		StabilityFee:        d("1.000000001547125958"), // %5 apr
		LiquidationPenalty:  d("0.05"),
		AuctionSize:         i(100),
		Prefix:              0x20,
		ConversionFactor:    i(6),
		SpotMarketID:        "bnb:usd",
		LiquidationMarketID: "bnb:usd",
	}
	newMarketIDCP := testCP
	newMarketIDCP.SpotMarketID = "btc:usd"

	newDebtLimitCP := testCP
	newDebtLimitCP.DebtLimit = c("usdx", 1000)

	newMarketIDAndDebtLimitCP := testCP
	newMarketIDCP.SpotMarketID = "btc:usd"
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

func (suite *PermissionsTestSuite) TestAllowedAssetParam_Allows() {
	testAP := bep3types.AssetParam{
		Denom:  "usdx",
		CoinID: 999,
		Limit:  i(1000000000),
		Active: true,
	}
	newCoinidAP := testAP
	newCoinidAP.CoinID = 0

	newLimitAP := testAP
	newLimitAP.Limit = i(1000)

	newCoinidAndLimitAP := testAP
	newCoinidAndLimitAP.CoinID = 0
	newCoinidAndLimitAP.Limit = i(1000)

	testcases := []struct {
		name          string
		allowed       AllowedAssetParam
		current       bep3types.AssetParam
		incoming      bep3types.AssetParam
		expectAllowed bool
	}{
		{
			name: "allowed change",
			allowed: AllowedAssetParam{
				Denom: "usdx",
				Limit: true,
			},
			current:       testAP,
			incoming:      newLimitAP,
			expectAllowed: true,
		},
		{
			name: "un-allowed change",
			allowed: AllowedAssetParam{
				Denom: "usdx",
				Limit: true,
			},
			current:       testAP,
			incoming:      newCoinidAP,
			expectAllowed: false,
		},
		{
			name: "allowed no change",
			allowed: AllowedAssetParam{
				Denom: "usdx",
				Limit: true,
			},
			current:       testAP,
			incoming:      testAP, // no change
			expectAllowed: true,
		},
		{
			name: "un-allowed change with allowed change",
			allowed: AllowedAssetParam{
				Denom: "usdx",
				Limit: true,
			},
			current:       testAP,
			incoming:      newCoinidAndLimitAP,
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

func (suite *PermissionsTestSuite) TestAllowedMarket_Allows() {
	testM := pricefeedtypes.Market{
		MarketID:   "bnb:usd",
		BaseAsset:  "bnb",
		QuoteAsset: "usd",
		Oracles:    []sdk.AccAddress{[]byte("a test address")},
		Active:     true,
	}
	newOraclesM := testM
	newOraclesM.Oracles = nil

	newActiveM := testM
	newActiveM.Active = false

	newOraclesAndActiveM := testM
	newOraclesAndActiveM.Oracles = nil
	newOraclesAndActiveM.Active = false

	testcases := []struct {
		name          string
		allowed       AllowedMarket
		current       pricefeedtypes.Market
		incoming      pricefeedtypes.Market
		expectAllowed bool
	}{
		{
			name: "allowed change",
			allowed: AllowedMarket{
				MarketID: "bnb:usd",
				Active:   true,
			},
			current:       testM,
			incoming:      newActiveM,
			expectAllowed: true,
		},
		{
			name: "un-allowed change",
			allowed: AllowedMarket{
				MarketID: "bnb:usd",
				Active:   true,
			},
			current:       testM,
			incoming:      newOraclesM,
			expectAllowed: false,
		},
		{
			name: "allowed no change",
			allowed: AllowedMarket{
				MarketID: "bnb:usd",
				Active:   true,
			},
			current:       testM,
			incoming:      testM, // no change
			expectAllowed: true,
		},
		{
			name: "un-allowed change with allowed change",
			allowed: AllowedMarket{
				MarketID: "bnb:usd",
				Active:   true,
			},
			current:       testM,
			incoming:      newOraclesAndActiveM,
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
