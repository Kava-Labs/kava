package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, k)
		case types.QueryGetClaims:
			return queryGetClaims(ctx, req, k)
		case types.QueryGetRewardPeriods:
			return queryGetRewardPeriods(ctx, req, k)
		case types.QueryGetClaimPeriods:
			return queryGetClaimPeriods(ctx, req, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

// query params in the store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	// Get params
	params := k.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query reward periods in the store
func queryGetRewardPeriods(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	// Get params
	rewardPeriods := k.GetAllRewardPeriods(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(k.cdc, rewardPeriods)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// query claim periods in the store
func queryGetClaimPeriods(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	// Get params
	claimPeriods := k.GetAllClaimPeriods(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(k.cdc, claimPeriods)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetClaims(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var requestParams types.QueryClaimsParams
	err := k.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	claims, _ := k.GetAllClaimsByAddressAndCollateralType(ctx, requestParams.Owner, requestParams.CollateralType)

	bz, err := codec.MarshalJSONIndent(k.cdc, claims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
