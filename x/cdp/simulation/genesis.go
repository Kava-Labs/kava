package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/kava-labs/kava/x/cdp/types"
)

func randomCdpGenState(selection int) types.GenesisState {
	switch selection {
	case 0:
		return types.GenesisState{
			Params: types.Params{
				GlobalDebtLimit:              sdk.NewCoins(sdk.NewInt64Coin("usdx", 100000000000000), sdk.NewInt64Coin("susd", 100000000000000)),
				SurplusAuctionThreshold:      types.DefaultSurplusThreshold,
				DebtAuctionThreshold:         types.DefaultDebtThreshold,
				SavingsDistributionFrequency: types.DefaultSavingsDistributionFrequency,
				CollateralParams: types.CollateralParams{
					{
						Denom:              "xrp",
						LiquidationRatio:   sdk.MustNewDecFromStr("2.0"),
						DebtLimit:          sdk.NewCoins(sdk.NewInt64Coin("usdx", 20000000000000), sdk.NewInt64Coin("susd", 20000000000000)),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000004431822130"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.075"),
						AuctionSize:        sdk.NewInt(10000000000),
						Prefix:             0x20,
						MarketID:           "xrp:usd",
						ConversionFactor:   sdk.NewInt(6),
					},
					{
						Denom:              "btc",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.25"),
						DebtLimit:          sdk.NewCoins(sdk.NewInt64Coin("usdx", 50000000000000), sdk.NewInt64Coin("susd", 50000000000000)),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000000782997609"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.05"),
						AuctionSize:        sdk.NewInt(50000000),
						Prefix:             0x21,
						MarketID:           "btc:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewCoins(sdk.NewInt64Coin("usdx", 30000000000000), sdk.NewInt64Coin("susd", 30000000000000)),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000002293273137"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.15"),
						AuctionSize:        sdk.NewInt(10000000000),
						Prefix:             0x22,
						MarketID:           "bnb:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				DebtParams: types.DebtParams{
					{
						Denom:            "usdx",
						ReferenceAsset:   "usd",
						ConversionFactor: sdk.NewInt(6),
						DebtFloor:        sdk.NewInt(10000000),
						SavingsRate:      sdk.MustNewDecFromStr("0.95"),
					},
					{
						Denom:            "susd",
						ReferenceAsset:   "usd",
						ConversionFactor: sdk.NewInt(6),
						DebtFloor:        sdk.NewInt(10000000),
						SavingsRate:      sdk.MustNewDecFromStr("0.5"),
					},
				},
			},
			StartingCdpID:            types.DefaultCdpStartingID,
			DebtDenom:                types.DefaultDebtDenom,
			GovDenom:                 types.DefaultGovDenom,
			CDPs:                     types.CDPs{},
			PreviousBlockTime:        types.DefaultPreviousBlockTime,
			PreviousDistributionTime: types.DefaultPreviousDistributionTime,
		}
	case 1:
		return types.GenesisState{
			Params: types.Params{
				GlobalDebtLimit:              sdk.NewCoins(sdk.NewInt64Coin("usdx", 100000000000000)),
				SurplusAuctionThreshold:      types.DefaultSurplusThreshold,
				DebtAuctionThreshold:         types.DefaultDebtThreshold,
				SavingsDistributionFrequency: types.DefaultSavingsDistributionFrequency,
				CollateralParams: types.CollateralParams{
					{
						Denom:              "bnb",
						LiquidationRatio:   sdk.MustNewDecFromStr("1.5"),
						DebtLimit:          sdk.NewCoins(sdk.NewInt64Coin("usdx", 100000000000000)),
						StabilityFee:       sdk.MustNewDecFromStr("1.000000002293273137"),
						LiquidationPenalty: sdk.MustNewDecFromStr("0.075"),
						AuctionSize:        sdk.NewInt(10000000000),
						Prefix:             0x20,
						MarketID:           "btc:usd",
						ConversionFactor:   sdk.NewInt(8),
					},
				},
				DebtParams: types.DebtParams{
					{
						Denom:            "usdx",
						ReferenceAsset:   "usd",
						ConversionFactor: sdk.NewInt(6),
						DebtFloor:        sdk.NewInt(10000000),
						SavingsRate:      sdk.MustNewDecFromStr("0.95"),
					},
				},
			},
			StartingCdpID:            types.DefaultCdpStartingID,
			DebtDenom:                types.DefaultDebtDenom,
			GovDenom:                 types.DefaultGovDenom,
			CDPs:                     types.CDPs{},
			PreviousBlockTime:        types.DefaultPreviousBlockTime,
			PreviousDistributionTime: types.DefaultPreviousDistributionTime,
		}
	default:
		panic("invalid genesis state selector")
	}
}

// RandomizedGenState generates a random GenesisState for cdp
func RandomizedGenState(simState *module.SimulationState) {

	cdpGenesis := randomCdpGenState(simState.Rand.Intn(2))

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, cdpGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(cdpGenesis)
}
