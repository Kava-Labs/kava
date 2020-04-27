package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/cdp/types"
)

// NewQuerier returns a new querier function
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetCdp:
			return queryGetCdp(ctx, req, keeper)
		case types.QueryGetCdps:
			return queryGetCdpsByDenom(ctx, req, keeper)
		case types.QueryGetCdpsByCollateralization:
			return queryGetCdpsByRatio(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		case types.QueryGetCdpDeposits:
			return queryGetDeposits(ctx, req, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint %s", types.ModuleName, path[0])
		}
	}
}

// query a specific cdp
func queryGetCdp(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	cdp, found := keeper.GetCdpByOwnerAndDenom(ctx, requestParams.Owner, requestParams.CollateralDenom)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralDenom)
	}

	augmentedCDP, err := keeper.LoadAugmentedCDP(ctx, cdp)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrLoadingAugmentedCDP, fmt.Sprintf("%v: %d", err.Error(), cdp.ID))
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, augmentedCDP)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil

}

// query deposits on a particular cdp
func queryGetDeposits(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpDeposits
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	cdp, found := keeper.GetCdpByOwnerAndDenom(ctx, requestParams.Owner, requestParams.CollateralDenom)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrCdpNotFound, "owner %s, denom %s", requestParams.Owner, requestParams.CollateralDenom)
	}

	deposits := keeper.GetDeposits(ctx, cdp.ID)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, deposits)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil

}

// query cdps with matching denom and ratio LESS THAN the input ratio
func queryGetCdpsByRatio(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpsByRatioParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	ratio, err := keeper.CalculateCollateralizationRatioFromAbsoluteRatio(ctx, requestParams.CollateralDenom, requestParams.Ratio)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "couldn't get collateralization ratio from absolute ratio")
	}

	cdps := keeper.GetAllCdpsByDenomAndRatio(ctx, requestParams.CollateralDenom, ratio)
	// augment CDPs by adding collateral value and collateralization ratio
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range cdps {
		augmentedCDP, err := keeper.LoadAugmentedCDP(ctx, cdp)
		if err == nil {
			augmentedCDPs = append(augmentedCDPs, augmentedCDP)
		}
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, augmentedCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query all cdps with matching collateral denom
func queryGetCdpsByDenom(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var requestParams types.QueryCdpsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	_, valid := keeper.GetDenomPrefix(ctx, requestParams.CollateralDenom)
	if !valid {
		return nil, sdkerrors.Wrap(types.ErrCollateralNotSupported, requestParams.CollateralDenom)
	}

	cdps := keeper.GetAllCdpsByDenom(ctx, requestParams.CollateralDenom)
	// augment CDPs by adding collateral value and collateralization ratio
	var augmentedCDPs types.AugmentedCDPs
	for _, cdp := range cdps {
		augmentedCDP, err := keeper.LoadAugmentedCDP(ctx, cdp)
		if err == nil {
			augmentedCDPs = append(augmentedCDPs, augmentedCDP)
		}
	}
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, augmentedCDPs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query params in the cdp store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
