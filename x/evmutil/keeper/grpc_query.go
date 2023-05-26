package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kava-labs/kava/x/evmutil/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Params queries module params
func (s queryServer) Params(stdCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(stdCtx)
	params := s.keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// DeployedCosmosCoinContracts gets contract addresses for deployed erc20 contracts
// representing cosmos-sdk coins
func (s queryServer) DeployedCosmosCoinContracts(
	goCtx context.Context,
	req *types.QueryDeployedCosmosCoinContractsRequest,
) (res *types.QueryDeployedCosmosCoinContractsResponse, err error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if len(req.CosmosDenoms) > 0 {
		panic("unimplemented")
	} else {
		res, err = getAllDeployedCosmosCoinContractsPage(&s.keeper, ctx, req.Pagination)
	}

	return res, err
}

// getAllDeployedCosmosCoinContractsPage gets a page of deployed contracts (no filtering)
func getAllDeployedCosmosCoinContractsPage(
	k *Keeper, ctx sdk.Context, pagination *query.PageRequest,
) (*types.QueryDeployedCosmosCoinContractsResponse, error) {
	contracts := make([]types.DeployedCosmosCoinContract, 0)
	contractStore := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.DeployedCosmosCoinContractKeyPrefix,
	)

	pageRes, err := query.FilteredPaginate(contractStore, pagination,
		func(key []byte, value []byte, accumulate bool) (bool, error) {
			address := types.BytesToInternalEVMAddress(value)
			contract := types.DeployedCosmosCoinContract{
				CosmosDenom: string(key),
				Address:     &address,
			}
			contracts = append(contracts, contract)
			return true, nil
		})
	if err != nil {
		return &types.QueryDeployedCosmosCoinContractsResponse{}, err
	}

	return &types.QueryDeployedCosmosCoinContractsResponse{
		DeployedCosmosCoinContracts: contracts,
		Pagination:                  pageRes,
	}, nil
}
