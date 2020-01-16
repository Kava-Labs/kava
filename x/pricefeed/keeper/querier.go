package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryCurrentPrice:
			return queryCurrentPrice(ctx, path[1:], req, keeper)
		case types.QueryRawPrices:
			return queryRawPrices(ctx, path[1:], req, keeper)
		case types.QueryMarkets:
			return queryMarkets(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown pricefeed query endpoint")
		}
	}

}

func queryCurrentPrice(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, sdkErr sdk.Error) {
	var requestParams types.QueryPricesParams
	_, found := keeper.GetMarket(ctx, requestParams.MarketID)
	if !found {
		return []byte{}, sdk.ErrUnknownRequest("asset not found")
	}
	currentPrice, sdkErr := keeper.GetCurrentPrice(ctx, requestParams.MarketID)
	if sdkErr != nil {
		return nil, sdkErr
	}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, currentPrice)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryRawPrices(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, sdkErr sdk.Error) {
	var requestParams types.QueryPricesParams
	_, found := keeper.GetMarket(ctx, requestParams.MarketID)
	if !found {
		return []byte{}, sdk.ErrUnknownRequest("asset not found")
	}
	rawPrices := keeper.GetRawPrices(ctx, requestParams.MarketID)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, rawPrices)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryMarkets(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, sdkErr sdk.Error) {
	markets := keeper.GetMarkets(ctx)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, markets)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}
