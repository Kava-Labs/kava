package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/kava-labs/kava/x/cdp/types"
)

// RandomizedGenState generates a random GenesisState for cdp
func RandomizedGenState(simState *module.SimulationState) {

	cdpGenesis := randomCdpGenState(simState.Rand.Intn(2))

	// hacky way to give accounts coins so they can create cdps (coins  includes usdx so it's possible to have sufficient balance to close a cdp)
	var authGenesis auth.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[auth.ModuleName], &authGenesis)
	totalCdpCoins := sdk.NewCoins()
	for _, acc := range authGenesis.Accounts {
		_, ok := acc.(supplyexported.ModuleAccountI)
		if ok {
			continue
		}
		coinsToAdd := sdk.NewCoins(
			sdk.NewCoin("bnb", sdk.NewInt(int64(simState.Rand.Intn(100000000000)))),
			sdk.NewCoin("xrp", sdk.NewInt(int64(simState.Rand.Intn(100000000000)))),
			sdk.NewCoin("btc", sdk.NewInt(int64(simState.Rand.Intn(500000000)))),
			sdk.NewCoin("usdx", sdk.NewInt(int64(simState.Rand.Intn(1000000000)))),
			sdk.NewCoin("ukava", sdk.NewInt(int64(simState.Rand.Intn(500000000000)))),
		)
		err := acc.SetCoins(acc.GetCoins().Add(coinsToAdd...))
		if err != nil {
			panic(err)
		}
		totalCdpCoins = totalCdpCoins.Add(coinsToAdd...)
		authGenesis.Accounts = replaceOrAppendAccount(authGenesis.Accounts, acc)
	}
	simState.GenState[auth.ModuleName] = simState.Cdc.MustMarshalJSON(authGenesis)

	var supplyGenesis supply.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[supply.ModuleName], &supplyGenesis)
	supplyGenesis.Supply = supplyGenesis.Supply.Add(totalCdpCoins...)
	simState.GenState[supply.ModuleName] = simState.Cdc.MustMarshalJSON(supplyGenesis)

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, cdpGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(cdpGenesis)
}

// In a list of accounts, replace the first account found with the same address. If not found, append the account.
func replaceOrAppendAccount(accounts []authexported.GenesisAccount, acc authexported.GenesisAccount) []authexported.GenesisAccount {
	newAccounts := accounts
	for i, a := range accounts {
		if a.GetAddress().Equals(acc.GetAddress()) {
			newAccounts[i] = acc
			return newAccounts
		}
	}
	return append(newAccounts, acc)
}

func randomCdpGenState(selection int) types.GenesisState {
	switch selection {
	case 0:
		return types.GenesisState{
			Params: types.Params{
				GlobalDebtLimit:              sdk.NewInt64Coin("usdx", 100000000000000),
				SurplusAuctionThreshold:      types.DefaultSurplusThreshold,
				DebtAuctionThreshold:         types.DefaultDebtThreshold,
				SavingsDistributionFrequency: types.DefaultSavingsDistributionFrequency,
				CollateralParams: types.CollateralParams{
					{
						Denom:               "xrp",
						LiquidationRatio:    sdk.MustNewDecFromStr("2.0"),
						DebtLimit:           sdk.NewInt64Coin("usdx", 20000000000000),
						StabilityFee:        sdk.MustNewDecFromStr("1.000000004431822130"),
						LiquidationPenalty:  sdk.MustNewDecFromStr("0.075"),
						AuctionSize:         sdk.NewInt(100000000000),
						Prefix:              0x20,
						SpotMarketID:        "xrp:usd",
						LiquidationMarketID: "xrp:usd",
						ConversionFactor:    sdk.NewInt(6),
					},
					{
						Denom:               "btc",
						LiquidationRatio:    sdk.MustNewDecFromStr("1.25"),
						DebtLimit:           sdk.NewInt64Coin("usdx", 50000000000000),
						StabilityFee:        sdk.MustNewDecFromStr("1.000000000782997609"),
						LiquidationPenalty:  sdk.MustNewDecFromStr("0.05"),
						AuctionSize:         sdk.NewInt(1000000000),
						Prefix:              0x21,
						SpotMarketID:        "btc:usd",
						LiquidationMarketID: "btc:usd",
						ConversionFactor:    sdk.NewInt(8),
					},
					{
						Denom:               "bnb",
						LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
						DebtLimit:           sdk.NewInt64Coin("usdx", 30000000000000),
						StabilityFee:        sdk.MustNewDecFromStr("1.000000002293273137"),
						LiquidationPenalty:  sdk.MustNewDecFromStr("0.15"),
						AuctionSize:         sdk.NewInt(1000000000000),
						Prefix:              0x22,
						SpotMarketID:        "bnb:usd",
						LiquidationMarketID: "bnb:usd",
						ConversionFactor:    sdk.NewInt(8),
					},
				},
				DebtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
			},
			StartingCdpID:            types.DefaultCdpStartingID,
			DebtDenom:                types.DefaultDebtDenom,
			GovDenom:                 types.DefaultGovDenom,
			CDPs:                     types.CDPs{},
			PreviousDistributionTime: types.DefaultPreviousDistributionTime,
		}
	case 1:
		return types.GenesisState{
			Params: types.Params{
				GlobalDebtLimit:              sdk.NewInt64Coin("usdx", 100000000000000),
				SurplusAuctionThreshold:      types.DefaultSurplusThreshold,
				DebtAuctionThreshold:         types.DefaultDebtThreshold,
				SavingsDistributionFrequency: types.DefaultSavingsDistributionFrequency,
				CollateralParams: types.CollateralParams{
					{
						Denom:               "bnb",
						LiquidationRatio:    sdk.MustNewDecFromStr("1.5"),
						DebtLimit:           sdk.NewInt64Coin("usdx", 100000000000000),
						StabilityFee:        sdk.MustNewDecFromStr("1.000000002293273137"),
						LiquidationPenalty:  sdk.MustNewDecFromStr("0.075"),
						AuctionSize:         sdk.NewInt(10000000000),
						Prefix:              0x20,
						SpotMarketID:        "bnb:usd",
						LiquidationMarketID: "bnb:usd",
						ConversionFactor:    sdk.NewInt(8),
					},
				},
				DebtParam: types.DebtParam{
					Denom:            "usdx",
					ReferenceAsset:   "usd",
					ConversionFactor: sdk.NewInt(6),
					DebtFloor:        sdk.NewInt(10000000),
					SavingsRate:      sdk.MustNewDecFromStr("0.95"),
				},
			},
			StartingCdpID:            types.DefaultCdpStartingID,
			DebtDenom:                types.DefaultDebtDenom,
			GovDenom:                 types.DefaultGovDenom,
			CDPs:                     types.CDPs{},
			PreviousDistributionTime: types.DefaultPreviousDistributionTime,
		}
	default:
		panic("invalid genesis state selector")
	}
}
