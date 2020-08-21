package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/pricefeed"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func NewPricefeedGenState(asset string, price sdk.Dec) app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: asset + ":usd", BaseAsset: asset, QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
			{
				MarketID:      asset + ":usd",
				OracleAddress: sdk.AccAddress{},
				Price:         price,
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func NewCDPGenState(asset string, liquidationRatio sdk.Dec) app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:              sdk.NewInt64Coin("usdx", 1000000000000),
			SurplusAuctionThreshold:      cdp.DefaultSurplusThreshold,
			SurplusAuctionLot:            cdp.DefaultSurplusLot,
			DebtAuctionThreshold:         cdp.DefaultDebtThreshold,
			DebtAuctionLot:               cdp.DefaultDebtLot,
			SavingsDistributionFrequency: cdp.DefaultSavingsDistributionFrequency,
			CollateralParams: cdp.CollateralParams{
				{
					Denom:               asset,
					Type:                asset + "-a",
					LiquidationRatio:    liquidationRatio,
					DebtLimit:           sdk.NewInt64Coin("usdx", 1000000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000001547125958"), // %5 apr
					LiquidationPenalty:  d("0.05"),
					AuctionSize:         i(100),
					Prefix:              0x20,
					ConversionFactor:    i(6),
					SpotMarketID:        asset + ":usd",
					LiquidationMarketID: asset + ":usd",
				},
			},
			DebtParam: cdp.DebtParam{
				Denom:            "usdx",
				ReferenceAsset:   "usd",
				ConversionFactor: i(6),
				DebtFloor:        i(10000000),
				SavingsRate:      d("0.9"),
			},
		},
		StartingCdpID:            cdp.DefaultCdpStartingID,
		DebtDenom:                cdp.DefaultDebtDenom,
		GovDenom:                 cdp.DefaultGovDenom,
		CDPs:                     cdp.CDPs{},
		PreviousDistributionTime: cdp.DefaultPreviousDistributionTime,
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}

func NewPricefeedGenStateMulti() app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeed.PostedPrice{
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
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}
func NewCDPGenStateMulti() app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:              sdk.NewInt64Coin("usdx", 1500000000000),
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
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}

func NewCDPGenStateHighDebtLimit() app.GenesisState {
	cdpGenesis := cdp.GenesisState{
		Params: cdp.Params{
			GlobalDebtLimit:              sdk.NewInt64Coin("usdx", 100000000000000),
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
					DebtLimit:           sdk.NewInt64Coin("usdx", 50000000000000),
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
					DebtLimit:           sdk.NewInt64Coin("usdx", 50000000000000),
					StabilityFee:        sdk.MustNewDecFromStr("1.000000000782997609"), // %2.5 apr
					LiquidationPenalty:  d("0.025"),
					AuctionSize:         i(10000000),
					Prefix:              0x21,
					SpotMarketID:        "btc:usd",
					LiquidationMarketID: "btc:usd",
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
	}
	return app.GenesisState{cdp.ModuleName: cdp.ModuleCdc.MustMarshalJSON(cdpGenesis)}
}

func cdps() (cdps cdp.CDPs) {
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	c1 := cdp.NewCDP(uint64(1), addrs[0], sdk.NewCoin("xrp", sdk.NewInt(10000000)), "xrp-a", sdk.NewCoin("usdx", sdk.NewInt(8000000)), tmtime.Canonical(time.Now()))
	c2 := cdp.NewCDP(uint64(2), addrs[1], sdk.NewCoin("xrp", sdk.NewInt(100000000)), "xrp-a", sdk.NewCoin("usdx", sdk.NewInt(10000000)), tmtime.Canonical(time.Now()))
	c3 := cdp.NewCDP(uint64(3), addrs[1], sdk.NewCoin("btc", sdk.NewInt(1000000000)), "btc-a", sdk.NewCoin("usdx", sdk.NewInt(10000000)), tmtime.Canonical(time.Now()))
	c4 := cdp.NewCDP(uint64(4), addrs[2], sdk.NewCoin("xrp", sdk.NewInt(1000000000)), "xrp-a", sdk.NewCoin("usdx", sdk.NewInt(500000000)), tmtime.Canonical(time.Now()))
	cdps = append(cdps, c1, c2, c3, c4)
	return
}
