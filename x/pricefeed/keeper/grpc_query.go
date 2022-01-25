package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Params implements the gRPC service handler for querying x/pricefeed parameters.
func (s queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Price implements the gRPC service handler for querying x/pricefeed price.
func (s queryServer) Price(c context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, found := s.keeper.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}
	currentPrice, sdkErr := s.keeper.GetCurrentPrice(ctx, req.MarketId)
	if sdkErr != nil {
		return nil, sdkErr
	}

	return &types.QueryPriceResponse{
		Price: types.CurrentPriceResponse(currentPrice)}, nil
}

func (s queryServer) Prices(c context.Context, req *types.QueryPricesRequest) (*types.QueryPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var currentPrices types.CurrentPriceResponses
	for _, cp := range s.keeper.GetCurrentPrices(ctx) {
		if cp.MarketID != "" {
			currentPrices = append(currentPrices, types.CurrentPriceResponse(cp))
		}
	}

	return &types.QueryPricesResponse{
		Prices: currentPrices,
	}, nil
}

func (s queryServer) RawPrices(c context.Context, req *types.QueryRawPricesRequest) (*types.QueryRawPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, found := s.keeper.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}

	var prices types.PostedPriceResponses
	for _, rp := range s.keeper.GetRawPrices(ctx, req.MarketId) {
		prices = append(prices, types.PostedPriceResponse{
			MarketID:      rp.MarketID,
			OracleAddress: rp.OracleAddress.String(),
			Price:         rp.Price,
			Expiry:        rp.Expiry,
		})
	}

	return &types.QueryRawPricesResponse{
		RawPrices: prices,
	}, nil
}

func (s queryServer) Oracles(c context.Context, req *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	oracles, err := s.keeper.GetOracles(ctx, req.MarketId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}

	var strOracles []string
	for _, oracle := range oracles {
		strOracles = append(strOracles, oracle.String())
	}

	return &types.QueryOraclesResponse{
		Oracles: strOracles,
	}, nil
}

func (s queryServer) Markets(c context.Context, req *types.QueryMarketsRequest) (*types.QueryMarketsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var markets types.MarketResponses
	for _, market := range s.keeper.GetMarkets(ctx) {
		markets = append(markets, market.ToMarketResponse())
	}

	return &types.QueryMarketsResponse{
		Markets: markets,
	}, nil
}
