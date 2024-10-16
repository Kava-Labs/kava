package testutil

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// lendGenesisBuilder builds the Hard and Pricefeed genesis states for setting up Kava Lend
type lendGenesisBuilder struct {
	hardMarkets []hardtypes.MoneyMarket
	pfMarkets   []pricefeedtypes.Market
	prices      []pricefeedtypes.PostedPrice
}

func NewLendGenesisBuilder() lendGenesisBuilder {
	return lendGenesisBuilder{}
}

func (b lendGenesisBuilder) Build() (hardtypes.GenesisState, pricefeedtypes.GenesisState) {
	hardGS := hardtypes.DefaultGenesisState()
	hardGS.Params.MoneyMarkets = b.hardMarkets

	pricefeedGS := pricefeedtypes.DefaultGenesisState()
	pricefeedGS.Params.Markets = b.pfMarkets
	pricefeedGS.PostedPrices = b.prices
	return hardGS, pricefeedGS
}

func (b lendGenesisBuilder) WithMarket(denom, spotMarketId string, price sdkmath.LegacyDec) lendGenesisBuilder {
	// add hard money market
	b.hardMarkets = append(b.hardMarkets,
		hardtypes.NewMoneyMarket(
			denom,
			hardtypes.NewBorrowLimit(false, sdkmath.LegacyNewDec(1e15), sdkmath.LegacyMustNewDecFromStr("0.6")),
			spotMarketId,
			sdkmath.NewInt(1e6),
			hardtypes.NewInterestRateModel(sdkmath.LegacyMustNewDecFromStr("0.05"), sdkmath.LegacyMustNewDecFromStr("2"), sdkmath.LegacyMustNewDecFromStr("0.8"), sdkmath.LegacyMustNewDecFromStr("10")),
			sdkmath.LegacyMustNewDecFromStr("0.05"),
			sdkmath.LegacyZeroDec(),
		),
	)

	// add pricefeed
	b.pfMarkets = append(b.pfMarkets,
		pricefeedtypes.Market{MarketID: spotMarketId, BaseAsset: denom, QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
	)
	b.prices = append(b.prices,
		pricefeedtypes.PostedPrice{
			MarketID:      spotMarketId,
			OracleAddress: sdk.AccAddress{},
			Price:         price,
			Expiry:        time.Now().Add(100 * time.Hour),
		},
	)

	return b
}
