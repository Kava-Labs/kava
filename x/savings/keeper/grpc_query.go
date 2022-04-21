package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kava-labs/kava/x/savings/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Params implements the gRPC service handler for querying x/savings parameters.
func (s queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (s queryServer) Deposits(ctx context.Context, req *types.QueryDepositsRequest) (*types.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	hasDenom := len(req.Denom) > 0
	hasOwner := len(req.Owner) > 0

	var owner sdk.AccAddress
	var err error
	if hasOwner {
		owner, err = sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
		}
	}

	var deposits types.Deposits
	switch {
	case hasOwner && hasDenom:
		deposit, found := s.keeper.GetDeposit(sdkCtx, owner)
		if found {
			for _, coin := range deposit.Amount {
				if coin.Denom == req.Denom {
					deposits = append(deposits, deposit)
				}
			}
		}
	case hasOwner:
		deposit, found := s.keeper.GetDeposit(sdkCtx, owner)
		if found {
			deposits = append(deposits, deposit)
		}
	case hasDenom:
		s.keeper.IterateDeposits(sdkCtx, func(deposit types.Deposit) (stop bool) {
			if deposit.Amount.AmountOf(req.Denom).IsPositive() {
				deposits = append(deposits, deposit)
			}
			return false
		})
	default:
		s.keeper.IterateDeposits(sdkCtx, func(deposit types.Deposit) (stop bool) {
			deposits = append(deposits, deposit)
			return false
		})
	}

	page, limit, err := query.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	start, end := client.Paginate(len(deposits), page, limit, 100)
	if start < 0 || end < 0 {
		deposits = types.Deposits{}
	} else {
		deposits = deposits[start:end]
	}

	return &types.QueryDepositsResponse{
		Deposits:   deposits,
		Pagination: nil,
	}, nil
}
