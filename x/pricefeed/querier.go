package pricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"

	abci "github.com/tendermint/tendermint/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryCurrentPrice:
			return queryCurrentPrice(ctx, path[1:], req, keeper)
		case QueryRawPrices:
			return queryRawPrices(ctx, path[1:], req, keeper)
		case QueryAssets:
			return queryAssets(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown pricefeed query endpoint")
		}
	}

}

func queryCurrentPrice(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	assetCode := path[0]
	_, found := keeper.GetAsset(ctx, assetCode)
	if !found {
		return []byte{}, sdk.ErrUnknownRequest("asset not found")
	}
	currentPrice := keeper.GetCurrentPrice(ctx, assetCode)

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, currentPrice)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryRawPrices(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var priceList QueryRawPricesResp
	assetCode := path[0]
	_, found := keeper.GetAsset(ctx, assetCode)
	if !found {
		return []byte{}, sdk.ErrUnknownRequest("asset not found")
	}
	rawPrices := keeper.GetRawPrices(ctx, assetCode)
	for _, price := range rawPrices {
		priceList = append(priceList, price.String())
	}
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, priceList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryAssets(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var assetList QueryAssetsResp
	assets := keeper.GetAssets(ctx)
	for _, asset := range assets {
		assetList = append(assetList, asset.String())
	}
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, assetList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}