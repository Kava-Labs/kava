package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryPrice:
			return queryPrice(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryPrices:
			return queryPrices(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryRawPrices:
			return queryRawPrices(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryOracles:
			return queryOracles(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryMarkets:
			return queryMarkets(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

func queryPrice(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) (res []byte, sdkErr error) {
	var requestParams types.QueryWithMarketIDParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, found := keeper.GetMarket(ctx, requestParams.MarketID)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrAssetNotFound, requestParams.MarketID)
	}
	currentPrice, sdkErr := keeper.GetCurrentPrice(ctx, requestParams.MarketID)
	if sdkErr != nil {
		return nil, sdkErr
	}
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, currentPrice)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryPrices(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) (res []byte, sdkErr error) {
	currentPrices := keeper.GetCurrentPrices(ctx)

	// Filter out invalid markets without a price
	var validCurrentPrices types.CurrentPrices
	for _, cp := range currentPrices {
		if cp.MarketID != "" {
			validCurrentPrices = append(validCurrentPrices, types.CurrentPrice(cp))
		}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, validCurrentPrices)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryRawPrices(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) (res []byte, sdkErr error) {
	var requestParams types.QueryWithMarketIDParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, found := keeper.GetMarket(ctx, requestParams.MarketID)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrAssetNotFound, requestParams.MarketID)
	}

	rawPrices := keeper.GetRawPrices(ctx, requestParams.MarketID)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, rawPrices)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryOracles(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) (res []byte, sdkErr error) {
	var requestParams types.QueryWithMarketIDParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	oracles, err := keeper.GetOracles(ctx, requestParams.MarketID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrAssetNotFound, requestParams.MarketID)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, oracles)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryMarkets(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) (res []byte, sdkErr error) {
	markets := keeper.GetMarkets(ctx)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, markets)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// query params in the pricefeed store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
