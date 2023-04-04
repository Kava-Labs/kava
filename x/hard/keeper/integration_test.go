package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

func NewHARDGenState(cdc codec.JSONCodec) app.GenesisState {
	hardGenesis := types.GenesisState{
		Params: types.NewParams(
			types.MoneyMarkets{
				types.MoneyMarket{
					Denom: "usdx",
					BorrowLimit: types.BorrowLimit{
						HasMaxLimit:  true,
						MaximumLimit: sdk.MustNewDecFromStr("100000000000"),
						LoanToValue:  sdk.MustNewDecFromStr("1"),
					},
					SpotMarketID:     "usdx:usd",
					ConversionFactor: sdk.NewInt(USDX_CF),
					InterestRateModel: types.InterestRateModel{
						BaseRateAPY:    sdk.MustNewDecFromStr("0.05"),
						BaseMultiplier: sdk.MustNewDecFromStr("2"),
						Kink:           sdk.MustNewDecFromStr("0.8"),
						JumpMultiplier: sdk.MustNewDecFromStr("10"),
					},
					ReserveFactor:          sdk.MustNewDecFromStr("0.05"),
					KeeperRewardPercentage: sdk.ZeroDec(),
				},
				types.MoneyMarket{
					Denom: "bnb",
					BorrowLimit: types.BorrowLimit{
						HasMaxLimit:  true,
						MaximumLimit: sdk.MustNewDecFromStr("3000000000000"),
						LoanToValue:  sdk.MustNewDecFromStr("0.5"),
					},
					SpotMarketID:     "bnb:usd",
					ConversionFactor: sdk.NewInt(USDX_CF),
					InterestRateModel: types.InterestRateModel{
						BaseRateAPY:    sdk.MustNewDecFromStr("0"),
						BaseMultiplier: sdk.MustNewDecFromStr("0.05"),
						Kink:           sdk.MustNewDecFromStr("0.8"),
						JumpMultiplier: sdk.MustNewDecFromStr("5.0"),
					},
					ReserveFactor:          sdk.MustNewDecFromStr("0.025"),
					KeeperRewardPercentage: sdk.MustNewDecFromStr("0.02"),
				},
				types.MoneyMarket{
					Denom: "busd",
					BorrowLimit: types.BorrowLimit{
						HasMaxLimit:  true,
						MaximumLimit: sdk.MustNewDecFromStr("1000000000000000"),
						LoanToValue:  sdk.MustNewDecFromStr("0.5"),
					},
					SpotMarketID:     "busd:usd",
					ConversionFactor: sdk.NewInt(100000000),
					InterestRateModel: types.InterestRateModel{
						BaseRateAPY:    sdk.MustNewDecFromStr("0"),
						BaseMultiplier: sdk.MustNewDecFromStr("0.5"),
						Kink:           sdk.MustNewDecFromStr("0.8"),
						JumpMultiplier: sdk.MustNewDecFromStr("5"),
					},
					ReserveFactor:          sdk.MustNewDecFromStr("0.025"),
					KeeperRewardPercentage: sdk.MustNewDecFromStr("0.02"),
				},
			},
			sdk.MustNewDecFromStr("10"),
		),
		PreviousAccumulationTimes: types.GenesisAccumulationTimes{
			types.NewGenesisAccumulationTime(
				"usdx",
				time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				sdk.OneDec(),
				sdk.OneDec(),
			),
		},
		Deposits:      types.DefaultDeposits,
		Borrows:       types.DefaultBorrows,
		TotalSupplied: sdk.NewCoins(),
		TotalBorrowed: sdk.NewCoins(),
		TotalReserves: sdk.NewCoins(),
	}
	return app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(&hardGenesis)}
}

func NewPricefeedGenStateMulti(cdc codec.JSONCodec) app.GenesisState {
	pfGenesis := pricefeedtypes.GenesisState{
		Params: pricefeedtypes.Params{
			Markets: []pricefeedtypes.Market{
				{MarketID: "usdx:usd", BaseAsset: "usdx", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				{MarketID: "busd:usd", BaseAsset: "busd", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
		PostedPrices: []pricefeedtypes.PostedPrice{
			{
				MarketID:      "usdx:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.OneDec(),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("618.13"),
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
	return app.GenesisState{pricefeedtypes.ModuleName: cdc.MustMarshalJSON(&pfGenesis)}
}
