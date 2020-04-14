package simulation

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/x/pricefeed/types"
	pricefeed "github.com/kava-labs/kava/x/pricefeed/types"
)

// RandomizedGenState generates a random GenesisState for pricefeed
func RandomizedGenState(simState *module.SimulationState) {
	// get the params with xrp, btc and bnb to usd
	// getPricefeedSimulationParams is defined to return params with xrp:usd, btc:usd, bnb:usd
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

// getPricefeedSimulationParams returns the params with xrp:usd, btc:usd, bnb:usd
func getPricefeedSimulationParams() types.Params {
	// SET UP THE PRICEFEED GENESIS STATE
	pricefeedGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				pricefeed.Market{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				pricefeed.Market{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				pricefeed.Market{MarketID: "bnb:usd", BaseAsset: "bnb", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
	}
	return pricefeedGenesis.Params
}

// getInitialPrice gets the starting price for each of the base assets
func getInitialPrice(marketId string) (price sdk.Dec) {
	switch marketId {
	case "btc:usd":
		return sdk.MustNewDecFromStr("7000")
	case "bnb:usd":
		return sdk.MustNewDecFromStr("14")
	case "xrp:usd":
		return sdk.MustNewDecFromStr("0.2")
	}
	panic(fmt.Sprintf("Invalid marketId in getInitialPrice: %s\n", marketId))
}
