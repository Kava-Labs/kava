package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/pricefeed"
)

func NewCDPGenStateMulti() app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:              sdk.NewInt64Coin("usdx", 2000000000000),
			SurplusAuctionThreshold:      cdp.DefaultSurplusThreshold,
			SurplusAuctionLot:            cdp.DefaultSurplusLot,
			DebtAuctionThreshold:         cdp.DefaultDebtThreshold,
			DebtAuctionLot:               cdp.DefaultDebtLot,
			SavingsDistributionFrequency: cdp.DefaultSavingsDistributionFrequency,
			CollateralParams: cdp.CollateralParams{
				{
					Denom:               "xrp",
					Type:                "xrp-a",
					LiquidationRatio:    sdk.MustNewDecFromStr("2.0"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(7000000000),
					Prefix:              0x20,
					SpotMarketID:        "xrp:usd",
					LiquidationMarketID: "xrp:usd",
					ConversionFactor:    i(6),
				},
				{
					Denom:               "btc",
					Type:                "btc-a",
					LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000000782997609"), // %2.5 apr
					LiquidationPenalty:  d("0.025"),
					AuctionSize:         i(10000000),
					Prefix:              0x21,
					SpotMarketID:        "btc:usd",
					LiquidationMarketID: "btc:usd",
					ConversionFactor:    i(8),
				},
				{
					Denom:               "bnb",
					Type:                "bnb-a",
					LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(50000000000),
					Prefix:              0x22,
					SpotMarketID:        "bnb:usd",
					LiquidationMarketID: "bnb:usd",
					ConversionFactor:    i(8),
				},
				{
					Denom:               "busd",
					Type:                "busd-a",
					LiquidationRatio:    d("1.01"),
					DebtLimit:           sdk.NewInt64Coin("usdx", 500000000000),
					StabilityFee:        sdk.OneDec(), // %0 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(10000000000),
					Prefix:              0x23,
					SpotMarketID:        "busd:usd",
					LiquidationMarketID: "busd:usd",
					ConversionFactor:    i(8),
				},
			},
			DebtParam: cdp.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: i(6),
				DebtFloor:        i(10000000),
				SavingsRate:      d("0.95"),
			},
		},
		StartingCdpID:            cdp.DefaultCdpStartingID,
		DebtDenom:                cdp.DefaultDebtDenom,
		GovDenom:                 cdp.DefaultGovDenom,
		CDPs:                     cdp.CDPs{},
		PreviousDistributionTime: cdp.DefaultPreviousDistributionTime,
		PreviousAccumulationTimes: cdp.GenesisAccumulationTimes{
			cdp.NewGenesisAccumulationTime("btc-a", time.Time{}, sdk.OneDec()),
			cdp.NewGenesisAccumulationTime("xrp-a", time.Time{}, sdk.OneDec()),
			cdp.NewGenesisAccumulationTime("busd-a", time.Time{}, sdk.OneDec()),
			cdp.NewGenesisAccumulationTime("bnb-a", time.Time{}, sdk.OneDec()),
		},
		TotalPrincipals: cdp.GenesisTotalPrincipals{
			cdp.NewGenesisTotalPrincipal("btc-a", sdk.ZeroInt()),
			cdp.NewGenesisTotalPrincipal("xrp-a", sdk.ZeroInt()),
			cdp.NewGenesisTotalPrincipal("busd-a", sdk.ZeroInt()),
			cdp.NewGenesisTotalPrincipal("bnb-a", sdk.ZeroInt()),
		},
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}

func NewPricefeedGenStateMulti() app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: "kava:usd", BaseAsset: "kava", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "busd:usd", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			{
				MarketID:      "kava:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("2.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "btc:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("8000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "xrp:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("17.25"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "busd:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.OneDec(),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func NewHardGenStateMulti() app.GenesisState {
	KAVA_CF := int64(1000000)
	USDX_CF := int64(1000000)
	BNB_CF := int64(100000000)
	BTCB_CF := int64(100000000)

	loanToValue, _ := sdk.NewDecFromStr("0.6")
	borrowLimit := sdk.NewDec(1000000000000000)

	hardGS := hard.NewGenesisState(hard.NewParams(
		true,
		hard.MoneyMarkets{
			hard.NewMoneyMarket("usdx", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "usdx:usd", sdk.NewInt(1000000), sdk.NewInt(USDX_CF*1000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("ukava", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "kava:usd", sdk.NewInt(1000000), sdk.NewInt(KAVA_CF*1000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("bnb", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "bnb:usd", sdk.NewInt(1000000), sdk.NewInt(BNB_CF*1000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			hard.NewMoneyMarket("btcb", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "btcb:usd", sdk.NewInt(1000000), sdk.NewInt(BTCB_CF*1000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
		},
		0, // LTV counter
	), hard.DefaultPreviousBlockTime)

	return app.GenesisState{hard.ModuleName: hard.ModuleCdc.MustMarshalJSON(hardGS)}
}
