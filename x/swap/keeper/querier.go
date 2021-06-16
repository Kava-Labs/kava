package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/swap/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, k)
		case types.QueryGetDeposits:
			return queryGetDeposits(ctx, req, k)
		case types.QueryGetPool:
			return queryGetPool(ctx, req, k)
		case types.QueryGetPools:
			return queryGetPools(ctx, req, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

// query params in the swap store
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

func queryGetDeposits(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {

	var params types.QueryDepositsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	hasPoolParam := len(params.Pool) > 0
	hasOwnerParam := len(params.Owner) > 0

	if !hasPoolParam {
		return []byte{}, fmt.Errorf("must specify param 'pool'")
	}
	if !hasOwnerParam {
		return []byte{}, fmt.Errorf("must specify param 'owner'")
	}

	pool, found := k.GetPool(ctx, params.Pool)
	if !found {
		return []byte{}, fmt.Errorf("pool %s does not exist", params.Pool)
	}

	depositorShares, found := k.GetDepositorShares(ctx, params.Owner, params.Pool)
	if !found {
		return []byte{}, fmt.Errorf("error fetching depositor %s shares for pool %s", params.Owner, params.Pool)
	}

	shareValue, err := pool.ShareValue(depositorShares)
	if err != nil {
		return []byte{}, err
	}

	var bz []byte
	bz, err = codec.MarshalJSONIndent(types.ModuleCdc, shareValue)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetPool(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {

	var params types.QueryPoolParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	hasPoolParam := len(params.Pool) > 0

	if !hasPoolParam {
		return []byte{}, fmt.Errorf("must specify param 'pool'")
	}

	pool, found := k.GetPool(ctx, params.Pool)
	if !found {
		return []byte{}, fmt.Errorf("pool %s does not exist", params.Pool)
	}

	totalShares := pool.TotalShares
	totalCoins, err := pool.ShareValue(totalShares)
	if err != nil {
		return []byte{}, err
	}
	poolStats := types.NewPoolStatsQueryResult(pool.Name(), totalCoins, totalShares)

	var bz []byte
	bz, err = codec.MarshalJSONIndent(types.ModuleCdc, poolStats)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetPools(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {

	var pools []types.Pool
	k.IteratePools(ctx, func(pool types.Pool) bool {
		pools = append(pools, pool)
		return false
	})

	var allPoolStats types.PoolStatsQueryResults
	for _, pool := range pools {
		totalShares := pool.TotalShares
		totalCoins, err := pool.ShareValue(totalShares)
		if err != nil {
			return []byte{}, err
		}
		poolStats := types.NewPoolStatsQueryResult(pool.Name(), totalCoins, totalShares)
		allPoolStats = append(allPoolStats, poolStats)
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, allPoolStats)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
