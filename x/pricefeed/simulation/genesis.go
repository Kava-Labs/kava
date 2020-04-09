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

	// SET UP THE PRICEFEED GENESIS STATE
	pricefeedGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				pricefeed.Market{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
				pricefeed.Market{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
			},
		},
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

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, pricefeedGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(pricefeedGenesis)

	// now go through and modify the params adding the accounts as an oracle to each of the markets

	params := pricefeedGenesis.Params // TODO QUESTION? IS THIS CORRECT? NEED TO CALL ANOTHER METHOD EG getPricefeedSimulationParams() ??
	genPrices := []pricefeed.PostedPrice{}
	for _, market := range params.Markets {
		for _, acc := range simState.Accounts {
			market.Oracles = append(market.Oracles, acc.Address)
			// TODO QUESTION is this the right way to get market id??
			genPrice := types.PostedPrice{market.MarketID, acc.Address, getInitialPrice(market.MarketID), simState.GenTimestamp.Add(time.Hour * 24)}
			genPrices = append(genPrices, genPrice)
		}
	}

}

func getInitialPrice(marketId string) (price sdk.Dec) {
	switch marketId {
	case "btc":
		return sdk.MustNewDecFromStr("7000") // TODO QUESTION - add some randomization?
	case "bnb":
		return sdk.MustNewDecFromStr("14")
	case "xrp":
		return sdk.MustNewDecFromStr("0.2")
	}

	return sdk.MustNewDecFromStr("0")

}
