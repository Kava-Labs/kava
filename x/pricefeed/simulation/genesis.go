package simulation

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kava-labs/kava/x/pricefeed/types"
	pricefeed "github.com/kava-labs/kava/x/pricefeed/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandomizedGenState generates a random GenesisState for pricefeed
func RandomizedGenState(simState *module.SimulationState) {

	// get the params with xrp, btc and bnb to usd
	// define getPricefeedSimulationParams to return params with xrp:usd, btc:usd, bnb:usd
	params := getPricefeedSimulationParams() // TODO QUESTION? IS THIS CORRECT? NEED TO CALL ANOTHER METHOD EG getPricefeedSimulationParams() ??

	// now go through and modify the params adding the accounts as an oracle to each of the markets
	genPrices := []types.PostedPrice{}
	newMarkets := []types.Market{}
	for _, market := range params.Markets {
		oracles := []sdk.AccAddress{}
		for _, acc := range simState.Accounts {
			oracles = []sdk.AccAddress{acc.Address}
			// define getInitialPrice to set the initial price for each market (ie a switch statement where btc:usd is ~7000, bnb:usd is 14, and xrp:usd is 0.2(
			genPrice := types.PostedPrice{market.MarketID, acc.Address, getInitialPrice(market.MarketID), simState.GenTimestamp.Add(time.Hour * 24)}
			genPrices = append(genPrices, genPrice)
		}
		newMarket := types.Market{market.MarketID, market.BaseAsset, market.QuoteAsset, oracles, market.Active}
		newMarkets = append(newMarkets, newMarket)
	}
	params = types.NewParams(newMarkets)
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

		// TODO QUESTION - are these prices necessary? they are re-set by getInitialPrice
		PostedPrices: []pricefeed.PostedPrice{
			// Bitcoin
			pricefeed.PostedPrice{
				MarketID:      "btc:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("7000.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			// Binance coin
			pricefeed.PostedPrice{
				MarketID:      "bnb:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("14.00"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
			// XRP ripple coin
			pricefeed.PostedPrice{
				MarketID:      "xrp:usd",
				OracleAddress: sdk.AccAddress{},
				Price:         sdk.MustNewDecFromStr("0.2"),
				Expiry:        time.Now().Add(1 * time.Hour),
			},
		},
	}
	return pricefeedGenesis.Params
}

func getInitialPrice(marketId string) (price sdk.Dec) {
	switch marketId {
	case "btc:usd":
		return sdk.MustNewDecFromStr("7000") // TODO QUESTION - add some randomization?
	case "bnb:usd":
		return sdk.MustNewDecFromStr("14")
	case "xrp:usd":
		return sdk.MustNewDecFromStr("0.2")
	}

	fmt.Printf("Invalid marketId in getInitialPrice: %s\n", marketId)
	return sdk.MustNewDecFromStr("0")

}
