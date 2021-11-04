package pricefeed

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

type QueryServer struct {
	keeper keeper.Keeper
}

// NewQueryServer returns an implementation of the pricefeed MsgServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper keeper.Keeper) types.QueryServer {
	return &QueryServer{keeper: keeper}
}

var _ types.QueryServer = QueryServer{}

// Params implements the gRPC service handler for querying x/pricefeed parameters.
func (k QueryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	params := k.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k QueryServer) Price(c context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, found := k.keeper.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}
	currentPrice, sdkErr := k.keeper.GetCurrentPrice(ctx, req.MarketId)
	if sdkErr != nil {
		return nil, sdkErr
	}

	return &types.QueryPriceResponse{
		Price: types.CurrentPrice{
			MarketID: currentPrice.MarketID,
			Price:    currentPrice.Price,
		}}, nil

}
func (k QueryServer) Prices(c context.Context, req *types.QueryPricesRequest) (*types.QueryPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	currentPrices := k.keeper.GetCurrentPrices(ctx)

	return &types.QueryPricesResponse{
		Prices: currentPrices,
	}, nil
}

func (k QueryServer) RawPrices(c context.Context, req *types.QueryRawPricesRequest) (*types.QueryRawPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, found := k.keeper.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}
	rawPrices := k.keeper.GetRawPrices(ctx, req.MarketId)

	return &types.QueryRawPricesResponse{
		RawPrices: rawPrices,
	}, nil
}

func (k QueryServer) Oracles(c context.Context, req *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	oracles, err := k.keeper.GetOracles(ctx, req.MarketId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}

	return &types.QueryOraclesResponse{
		Oracles: oracles,
	}, nil
}

func (k QueryServer) Markets(c context.Context, req *types.QueryMarketsRequest) (*types.QueryMarketsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	markets := k.keeper.GetMarkets(ctx)

	return &types.QueryMarketsResponse{
		Markets: markets,
	}, nil
}
