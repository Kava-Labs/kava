package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryGetCdp:
			return queryGetCdp(ctx, req, keeper)
		case types.QueryGetCdps:
			return queryGetCdpsByDenom(ctx, req, keeper)
		case types.QueryGetCdpsByCollateralization:
			return queryGetCdpsByRatio(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}

// query a specific cdp
func queryGetCdp(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var requestParams types.QueryCdpParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, types.ErrInvalidCollateralDenom(keeper.codespace, requestParams.CollateralDenom)
	}

	cdp, found := keeper.GetCdpByOwnerAndDenom(ctx, requestParams.Owner, requestParams.CollateralDenom)
	if !found {
		return nil, types.ErrCdpNotFound(keeper.codespace, requestParams.Owner, requestParams.CollateralDenom)
	}

	augmentedCDP, err := keeper.LoadAugmentedCDP(ctx, cdp)
	if err != nil {
		// TODO: types.ErrLoadingAugmentedCDP()
		return nil, types.ErrCdpNotFound(keeper.codespace, requestParams.Owner, requestParams.CollateralDenom)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, augmentedCDP)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil

}

// query cdps with matching denom and ratio LESS THAN the input ratio
func queryGetCdpsByRatio(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var requestParams types.QueryCdpsByRatioParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, types.ErrInvalidCollateralDenom(keeper.codespace, requestParams.CollateralDenom)
	}

	cdps := keeper.GetAllCdpsByDenomAndRatio(ctx, requestParams.CollateralDenom, requestParams.Ratio)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, cdps)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// query all cdps with matching collateral denom
func queryGetCdpsByDenom(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var requestParams types.QueryCdpsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, types.ErrInvalidCollateralDenom(keeper.codespace, requestParams.CollateralDenom)
	}

	cdps := keeper.GetAllCdpsByDenom(ctx, requestParams.CollateralDenom)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, cdps)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// query params in the cdp store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
