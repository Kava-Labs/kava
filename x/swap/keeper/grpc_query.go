package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kava-labs/kava/x/swap/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Params implements the gRPC service handler for querying x/swap parameters.
func (s queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Pools implements the Query/Pools gRPC method
func (s queryServer) Pools(c context.Context, req *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(s.keeper.key), types.PoolKeyPrefix)

	var queryResults []types.PoolResponse
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_, value []byte, shouldAccumulate bool) (bool, error) {
		var poolRecord types.PoolRecord
		err := s.keeper.cdc.Unmarshal(value, &poolRecord)
		if err != nil {
			return false, err
		}

		if (len(req.PoolId) > 0) && strings.Compare(poolRecord.PoolID, req.PoolId) != 0 {
			return false, nil
		}

		if shouldAccumulate {
			denominatedPool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
			if err != nil {
				return true, types.ErrInvalidPool
			}
			totalCoins := denominatedPool.ShareValue(denominatedPool.TotalShares())
			queryResult := types.PoolResponse{
				Name:        poolRecord.PoolID,
				Coins:       totalCoins,
				TotalShares: denominatedPool.TotalShares(),
			}
			queryResults = append(queryResults, queryResult)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryPoolsResponse{
		Pools:      queryResults,
		Pagination: pageRes,
	}, nil
}

// Deposits implements the Query/Deposits gRPC method
func (s queryServer) Deposits(c context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(s.keeper.key), types.DepositorPoolSharesPrefix)

	records := types.ShareRecords{}
	pageRes, err := query.FilteredPaginate(
		store,
		req.Pagination,
		func(key []byte, value []byte, accumulate bool) (bool, error) {

			var record types.ShareRecord
			err := s.keeper.cdc.Unmarshal(value, &record)
			if err != nil {
				return false, err
			}

			// Filter for results match the request's pool ID/owner params if given
			matchOwner, matchPool := true, true
			if len(req.Owner) > 0 {
				matchOwner = record.Depositor.String() == req.Owner
			}
			if len(req.PoolId) > 0 {
				matchPool = strings.Compare(record.PoolID, req.PoolId) == 0
			}
			if !(matchOwner && matchPool) {
				// inform paginate that there was no match on this key
				return false, nil
			}
			if accumulate {
				// only add to results if paginate tells us to
				records = append(records, record)
			}
			// inform paginate that were was a match on this key
			return true, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var queryResults []types.DepositResponse
	for _, record := range records {
		pool, err := s.keeper.loadDenominatedPool(ctx, record.PoolID)
		if err != nil {
			return nil, err
		}
		shareValue := pool.ShareValue(record.SharesOwned)
		queryResult := types.DepositResponse{
			Depositor:   record.Depositor.String(),
			PoolId:      record.PoolID,
			SharesOwned: record.SharesOwned,
			SharesValue: shareValue,
		}
		queryResults = append(queryResults, queryResult)
	}

	return &types.QueryDepositsResponse{
		Deposits:   queryResults,
		Pagination: pageRes,
	}, nil
}
