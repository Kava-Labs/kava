package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryPrice:
			return queryPrice(ctx, req, keeper)
		case types.QueryRawPrices:
			return queryRawPrices(ctx, req, keeper)
		case types.QueryOracles:
			return queryOracles(ctx, req, keeper)
		case types.QueryMarkets:
			return queryMarkets(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}

}

func queryPrice(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, sdkErr error) {
	var requestParams types.QueryWithMarketIDParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, found := keeper.GetMarket(ctx, requestParams.MarketID)
	if !found {
		return []byte{}, sdkerrors.Wrap(types.ErrAssetNotFound, requestParams.MarketID)
	}
	currentPrice, sdkErr := keeper.GetCurrentPrice(ctx, requestParams.MarketID)
	if sdkErr != nil {
		return nil, sdkErr
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, currentPrice)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryRawPrices(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, sdkErr error) {
	var requestParams types.QueryWithMarketIDParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, found := keeper.GetMarket(ctx, requestParams.MarketID)
	if !found {
		return []byte{}, sdkerrors.Wrap(types.ErrAssetNotFound, requestParams.MarketID)
	}
	rawPrices := keeper.GetRawPrices(ctx, requestParams.MarketID)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, rawPrices)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryOracles(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, sdkErr error) {
	var requestParams types.QueryWithMarketIDParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	oracles, err := keeper.GetOracles(ctx, requestParams.MarketID)
	if err != nil {
		return []byte{}, sdkerrors.Wrap(types.ErrAssetNotFound, requestParams.MarketID)
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, oracles)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryMarkets(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, sdkErr error) {
	markets := keeper.GetMarkets(ctx)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, markets)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// query params in the pricefeed store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
