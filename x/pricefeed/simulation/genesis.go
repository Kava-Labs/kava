package simulation

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

// RandomizedGenState generates a random GenesisState for pricefeed
func RandomizedGenState(simState *module.SimulationState) {

	params := getPricefeedSimulationParams()
	markets := []types.Market{}
	genPrices := []types.PostedPrice{}
	// chose one account to be the oracle
	oracle := simState.Accounts[simulation.RandIntBetween(simState.Rand, 0, len(simState.Accounts))]
	for _, market := range params.Markets {
		updatedMarket := types.Market{market.MarketID, market.BaseAsset, market.QuoteAsset, []sdk.AccAddress{oracle.Address}, true}
		markets = append(markets, updatedMarket)
		genPrice := types.PostedPrice{market.MarketID, oracle.Address, getInitialPrice(market.MarketID), simState.GenTimestamp.Add(time.Hour * 24)}
		genPrices = append(genPrices, genPrice)
	}
	params = types.NewParams(markets)
	pricefeedGenesis := types.NewGenesisState(params, genPrices)
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, pricefeedGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(pricefeedGenesis)
}

func getPricefeedSimulationParams() types.Params {
	return types.Params{
		Markets: []types.Market{
			types.Market{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			types.Market{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			types.Market{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
		},
	}
}

func getInitialPrice(marketID string) sdk.Dec {
	switch marketID {
	case "xrp:usd":
		return sdk.MustNewDecFromStr("0.20")
	case "bnb:usd":
		return sdk.MustNewDecFromStr("14.0")
	case "btc:usd":
		return sdk.MustNewDecFromStr("7000.0")
	default:
		panic("invalid market id")
	}

}
