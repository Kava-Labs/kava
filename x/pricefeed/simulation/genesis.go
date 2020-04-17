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

var (
	// BaseAssets is a list of collateral asset denoms
	BaseAssets = [3]string{"bnb", "xrp", "btc"}
	QuoteAsset = "usd"
)

// RandomizedGenState generates a random GenesisState for pricefeed
func RandomizedGenState(simState *module.SimulationState) {
	pricefeedGenesis := loadPricefeedGenState(simState)
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, pricefeedGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(pricefeedGenesis)
}

// loadPricefeedGenState loads a valid pricefeed gen state
func loadPricefeedGenState(simState *module.SimulationState) pricefeed.GenesisState {
	var markets []pricefeed.Market
	var postedPrices []pricefeed.PostedPrice
	for _, denom := range BaseAssets {
		// Select an account to be the oracle
		oracle := simState.Accounts[simulation.RandIntBetween(simState.Rand, 0, len(simState.Accounts))]

		marketID := fmt.Sprintf("%s:%s", denom, QuoteAsset)
		// Construct market for asset
		market := pricefeed.Market{
			MarketID:   marketID,
			BaseAsset:  denom,
			QuoteAsset: QuoteAsset,
			Oracles:    []sdk.AccAddress{oracle.Address},
			Active:     true,
		}

		// Construct posted price for asset
		postedPrice := pricefeed.PostedPrice{
			MarketID:      market.MarketID,
			OracleAddress: oracle.Address,
			Price:         getInitialPrice(marketID),
			Expiry:        simState.GenTimestamp.Add(time.Hour * 24),
		}
		markets = append(markets, market)
		postedPrices = append(postedPrices, postedPrice)
	}
	params := pricefeed.NewParams(markets)
	return pricefeed.NewGenesisState(params, postedPrices)
}

// getInitialPrice gets the starting price for each of the base assets
func getInitialPrice(marketID string) (price sdk.Dec) {
	switch marketID {
	case "btc:usd":
		return sdk.MustNewDecFromStr("7000")
	case "bnb:usd":
		return sdk.MustNewDecFromStr("14")
	case "xrp:usd":
		return sdk.MustNewDecFromStr("0.2")
	default:
		return sdk.MustNewDecFromStr("20") // Catch future additional assets
	}
}
