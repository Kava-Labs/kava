package keeper_test

import (
	// "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	// codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	// "github.com/kava-labs/kava/app"
	// cdptypes "github.com/kava-labs/kava/x/cdp/types"
	// committeetypes "github.com/kava-labs/kava/x/committee/types"
	// "github.com/kava-labs/kava/x/incentive/testutil"
	// pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

// func NewCDPGenStateMulti() app.GenesisState {
// 	cdpGenesis := cdptypes.GenesisState{
// 		Params: cdptypes.Params{
// 			GlobalDebtLimit:         sdk.NewInt64Coin("usdx", 2000000000000),
// 			SurplusAuctionThreshold: cdptypes.DefaultSurplusThreshold,
// 			SurplusAuctionLot:       cdptypes.DefaultSurplusLot,
// 			DebtAuctionThreshold:    cdptypes.DefaultDebtThreshold,
// 			DebtAuctionLot:          cdptypes.DefaultDebtLot,
// 			CollateralParams: cdptypes.CollateralParams{
// 				{
// 					Denom:               "xrp",
// 					Type:                "xrp-a",
// 					LiquidationRatio:    sdk.MustNewDecFromStr("2.0"),
// 					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
// 					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
// 					LiquidationPenalty:  d("0.05"),
// 					AuctionSize:         i(7000000000),
// 					SpotMarketID:        "xrp:usd",
// 					LiquidationMarketID: "xrp:usd",
// 					ConversionFactor:    i(6),
// 				},
// 				{
// 					Denom:               "btc",
// 					Type:                "btc-a",
// 					LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
// 					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
// 					StabilityFee:        sdk.MustNewDecFromStr("1.000000000782997609"), // %2.5 apr
// 					LiquidationPenalty:  d("0.025"),
// 					AuctionSize:         i(10000000),
// 					SpotMarketID:        "btc:usd",
// 					LiquidationMarketID: "btc:usd",
// 					ConversionFactor:    i(8),
// 				},
// 				{
// 					Denom:               "bnb",
// 					Type:                "bnb-a",
// 					LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
// 					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
// 					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
// 					LiquidationPenalty:  d("0.05"),
// 					AuctionSize:         i(50000000000),
// 					SpotMarketID:        "bnb:usd",
// 					LiquidationMarketID: "bnb:usd",
// 					ConversionFactor:    i(8),
// 				},
// 				{
// 					Denom:               "busd",
// 					Type:                "busd-a",
// 					LiquidationRatio:    d("1.01"),
// 					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
// 					StabilityFee:        sdk.OneDec(), // %0 apr
// 					LiquidationPenalty:  d("0.05"),
// 					AuctionSize:         i(10000000000),
// 					SpotMarketID:        "busd:usd",
// 					LiquidationMarketID: "busd:usd",
// 					ConversionFactor:    i(8),
// 				},
// 			},
// 			DebtParam: cdptypes.DebtParam{
// 				Denom:            "usdx",
// 				ReferenceAsset:   "usd",
// 				ConversionFactor: i(6),
// 				DebtFloor:        i(10000000),
// 			},
// 		},
// 		StartingCdpID: cdptypes.DefaultCdpStartingID,
// 		DebtDenom:     cdptypes.DefaultDebtDenom,
// 		GovDenom:      cdptypes.DefaultGovDenom,
// 		CDPs:          cdptypes.CDPs{},
// 		PreviousAccumulationTimes: cdptypes.GenesisAccumulationTimes{
// 			cdptypes.NewGenesisAccumulationTime("btc-a", time.Time{}, sdk.OneDec()),
// 			cdptypes.NewGenesisAccumulationTime("xrp-a", time.Time{}, sdk.OneDec()),
// 			cdptypes.NewGenesisAccumulationTime("busd-a", time.Time{}, sdk.OneDec()),
// 			cdptypes.NewGenesisAccumulationTime("bnb-a", time.Time{}, sdk.OneDec()),
// 		},
// 		TotalPrincipals: cdptypes.GenesisTotalPrincipals{
// 			cdptypes.NewGenesisTotalPrincipal("btc-a", sdk.ZeroInt()),
// 			cdptypes.NewGenesisTotalPrincipal("xrp-a", sdk.ZeroInt()),
// 			cdptypes.NewGenesisTotalPrincipal("busd-a", sdk.ZeroInt()),
// 			cdptypes.NewGenesisTotalPrincipal("bnb-a", sdk.ZeroInt()),
// 		},
// 	}
// 	return app.GenesisState{cdptypes.ModuleName: cdptypes.ModuleCdc.MustMarshalJSON(cdpGenesis)}
// }

// func NewPricefeedGenStateMultiFromTime(t time.Time) app.GenesisState {
// 	expiry := 100 * 365 * 24 * time.Hour // 100 years

// 	pfGenesis := pricefeedtypes.GenesisState{
// 		Params: pricefeedtypes.Params{
// 			Markets: []pricefeedtypes.Market{
// 				{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
// 				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
// 				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
// 				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
// 				{MarketID: "busd:usd", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
// 				{MarketID: "zzz:usd", BaseAsset: "zzz", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
// 			},
// 		},
// 		PostedPrices: []pricefeedtypes.PostedPrice{
// 			{
// 				MarketID:      "kava:usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         sdk.MustNewDecFromStr("2.00"),
// 				Expiry:        t.Add(expiry),
// 			},
// 			{
// 				MarketID:      "btc:usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         sdk.MustNewDecFromStr("8000.00"),
// 				Expiry:        t.Add(expiry),
// 			},
// 			{
// 				MarketID:      "xrp:usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         sdk.MustNewDecFromStr("0.25"),
// 				Expiry:        t.Add(expiry),
// 			},
// 			{
// 				MarketID:      "bnb:usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         sdk.MustNewDecFromStr("17.25"),
// 				Expiry:        t.Add(expiry),
// 			},
// 			{
// 				MarketID:      "busd:usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         sdk.OneDec(),
// 				Expiry:        t.Add(expiry),
// 			},
// 			{
// 				MarketID:      "zzz:usd",
// 				OracleAddress: sdk.AccAddress{},
// 				Price:         sdk.MustNewDecFromStr("2.00"),
// 				Expiry:        t.Add(expiry),
// 			},
// 		},
// 	}
// 	return app.GenesisState{pricefeedtypes.ModuleName: pricefeedtypes.ModuleCdc.MustMarshalJSON(pfGenesis)}
// }

// func NewHardGenStateMulti(genTime time.Time) testutil.HardGenesisBuilder {
// 	kavaMM := testutil.NewStandardMoneyMarket("ukava")
// 	kavaMM.SpotMarketID = "kava:usd"
// 	btcMM := testutil.NewStandardMoneyMarket("btcb")
// 	btcMM.SpotMarketID = "btc:usd"

// 	builder := testutil.NewHardGenesisBuilder().WithGenesisTime(genTime).
// 		WithInitializedMoneyMarket(testutil.NewStandardMoneyMarket("usdx")).
// 		WithInitializedMoneyMarket(kavaMM).
// 		WithInitializedMoneyMarket(testutil.NewStandardMoneyMarket("bnb")).
// 		WithInitializedMoneyMarket(btcMM).
// 		WithInitializedMoneyMarket(testutil.NewStandardMoneyMarket("xrp")).
// 		WithInitializedMoneyMarket(testutil.NewStandardMoneyMarket("zzz"))
// 	return builder
// }

// func NewStakingGenesisState() app.GenesisState {
// 	genState := stakingtypes.DefaultGenesisState()
// 	genState.Params.BondDenom = "ukava"
// 	return app.GenesisState{
// 		stakingtypes.ModuleName: stakingtypes.ModuleCdc.MustMarshalJSON(genState),
// 	}
// }

// func NewCommitteeGenesisState(committeeID uint64, members ...sdk.AccAddress) app.GenesisState {
// 	genState := committeetypes.DefaultGenesisState()

// 	com, err := committeetypes.NewMemberCommittee(
// 		committeeID,
// 		"This committee is for testing.",
// 		members,
// 		[]committeetypes.Permission{committeetypes.GodPermission{}},
// 		sdk.MustNewDecFromStr("0.666666667"),
// 		time.Hour*24*7,
// 		committeetypes.TALLY_OPTION_FIRST_PAST_THE_POST,
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	genesisComms := committeetypes.Committees{com}

// 	var anyUnpacker codectypes.AnyUnpacker
// 	err = genesisComms.UnpackInterfaces(anyUnpacker)
// 	if err != nil {
// 		panic(err)
// 	}

// 	genState.Committees = genesisComms

// 	return app.GenesisState{
// 		committeetypes.ModuleName: committeetypes.ModuleCdc.MustMarshalJSON(genState),
// 	}
// }
