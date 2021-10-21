package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
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

	var records types.ShareRecords
	if len(params.Owner) > 0 {
		records = k.GetAllDepositorSharesByOwner(ctx, params.Owner)
	} else {
		unfilteredRecords := k.GetAllDepositorShares(ctx)
		records = filterShareRecords(ctx, unfilteredRecords, params)
	}

	// Augment each deposit result with the actual share value of depositor's shares
	var queryResults types.DepositsQueryResults
	for _, record := range records {
		pool, err := k.loadDenominatedPool(ctx, record.PoolID)
		if err != nil {
			return nil, err
		}
		shareValue := pool.ShareValue(record.SharesOwned)
		queryResult := types.NewDepositsQueryResult(record, shareValue)
		queryResults = append(queryResults, queryResult)
	}

	var bz []byte
	bz, err = codec.MarshalJSONIndent(types.ModuleCdc, queryResults)
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
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "must specify pool param")

	}

	pool, err := k.loadDenominatedPool(ctx, params.Pool)
	if err != nil {
		return nil, err
	}
	totalCoins := pool.ShareValue(pool.TotalShares())
	poolStats := types.NewPoolStatsQueryResult(params.Pool, totalCoins, pool.TotalShares())

	var bz []byte
	bz, err = codec.MarshalJSONIndent(types.ModuleCdc, poolStats)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetPools(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	pools := k.GetAllPools(ctx)

	var queryResults types.PoolStatsQueryResults
	for _, pool := range pools {
		denomPool, err := k.loadDenominatedPool(ctx, pool.PoolID)
		if err != nil {
			return nil, err
		}
		totalCoins := denomPool.ShareValue(denomPool.TotalShares())
		queryResult := types.NewPoolStatsQueryResult(pool.PoolID, totalCoins, denomPool.TotalShares())
		queryResults = append(queryResults, queryResult)
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, queryResults)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// filterShareRecords retrieves share records filtered by a given set of params.
// If no filters are provided, all share records will be returned in paginated form.
func filterShareRecords(ctx sdk.Context, records types.ShareRecords, params types.QueryDepositsParams) types.ShareRecords {
	filteredRecords := make(types.ShareRecords, 0, len(records))

	for _, s := range records {
		matchOwner, matchPool := true, true

		// match owner address (if supplied)
		if len(params.Owner) > 0 {
			matchOwner = s.Depositor.Equals(params.Owner)
		}

		// match pool ID (if supplied)
		if len(params.Pool) > 0 {
			matchPool = strings.Compare(s.PoolID, params.Pool) == 0
		}

		if matchOwner && matchPool {
			filteredRecords = append(filteredRecords, s)
		}
	}

	start, end := client.Paginate(len(filteredRecords), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		filteredRecords = types.ShareRecords{}
	} else {
		filteredRecords = filteredRecords[start:end]
	}

	return filteredRecords
}

func (k Keeper) loadDenominatedPool(ctx sdk.Context, poolID string) (*types.DenominatedPool, error) {
	poolRecord, found := k.GetPool(ctx, poolID)
	if !found {
		return &types.DenominatedPool{}, types.ErrInvalidPool
	}
	denominatedPool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
	if err != nil {
		return &types.DenominatedPool{}, types.ErrInvalidPool
	}
	return denominatedPool, nil
}
