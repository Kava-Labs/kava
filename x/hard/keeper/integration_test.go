package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard/types"
)

func NewHARDGenState(cdc codec.JSONCodec) app.GenesisState {
	hardGenesis := types.GenesisState{
		Params: types.NewParams(
			types.MoneyMarkets{
				types.NewMoneyMarket(
					"usdx",
					types.NewBorrowLimit(
						true,
						sdk.MustNewDecFromStr("100000000000"),
						sdk.MustNewDecFromStr("1"),
					),
					"usdx:usd",
					sdk.NewInt(USDX_CF),
					types.NewInterestRateModel(
						sdk.MustNewDecFromStr("0.05"),
						sdk.MustNewDecFromStr("2"),
						sdk.MustNewDecFromStr("0.8"),
						sdk.MustNewDecFromStr("10"),
					),
					sdk.MustNewDecFromStr("0.05"),
					sdk.ZeroDec(),
				),
				types.NewMoneyMarket(
					"bnb",
					types.NewBorrowLimit(
						true,
						sdk.MustNewDecFromStr("3000000000000"),
						sdk.MustNewDecFromStr("0.5"),
					),
					"bnb:usd:30",
					sdk.NewInt(USDX_CF),
					types.NewInterestRateModel(
						sdk.MustNewDecFromStr("0"),
						sdk.MustNewDecFromStr("0.05"),
						sdk.MustNewDecFromStr("0.8"),
						sdk.MustNewDecFromStr("5.0"),
					),
					sdk.MustNewDecFromStr("0.025"),
					sdk.MustNewDecFromStr("0.02"),
				),
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
