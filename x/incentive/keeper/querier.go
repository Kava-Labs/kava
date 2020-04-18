package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, k)
		case types.QueryGetClaims:
			return queryGetClaims(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown incentive query endpoint")
		}
	}
}

// query params in the store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	// Get params
	params := k.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryGetClaims(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var requestParams types.QueryClaimsParams
	err := k.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	claims, _ := k.GetClaimsByAddressAndDenom(ctx, requestParams.Owner, requestParams.Denom)

	bz, err := codec.MarshalJSONIndent(k.cdc, claims)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
