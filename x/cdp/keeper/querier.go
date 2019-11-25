package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)



func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryGetCdps:
			return queryGetCdps(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}


// queryGetCdps fetches CDPs, optionally filtering by any of the query params (in QueryCdpsParams).
// While CDPs do not have an ID, this method can be used to get one CDP by specifying the collateral and owner.
func queryGetCdps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Decode request
	var requestParams types.QueryCdpsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	// Get CDPs
	var cdps types.CDPs
	if len(requestParams.Owner) != 0 {
		if len(requestParams.CollateralDenom) != 0 {
			// owner and collateral specified - get a single CDP
			cdp, found := keeper.GetCDP(ctx, requestParams.Owner, requestParams.CollateralDenom)
			if !found {
				cdp = types.CDP{Owner: requestParams.Owner, CollateralDenom: requestParams.CollateralDenom, CollateralAmount: sdk.ZeroInt(), Debt: sdk.ZeroInt()}
			}
			cdps = types.CDPs{cdp}
		} else {
			// owner, but no collateral specified - get all CDPs for one address
			return nil, sdk.ErrInternal("getting all CDPs belonging to one owner not implemented")
		}
	} else {
		// owner not specified -- get all CDPs or all CDPs of one collateral type, optionally filtered by price
		var errSdk sdk.Error // := doesn't work here
		cdps, errSdk = keeper.GetCDPs(ctx, requestParams.CollateralDenom, requestParams.UnderCollateralizedAt)
		if errSdk != nil {
			return nil, errSdk
		}

	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, cdps)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// queryGetParams fetches the cdp module parameters
// TODO does this need to exist? Can you use cliCtx.QueryStore instead?
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
