package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

var _ types.QueryServer = Keeper{}

// Params implements the gRPC service handler for querying x/swap parameters.
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) Price(c context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, found := k.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "invalid market ID")
	}
	currentPrice, sdkErr := k.GetCurrentPrice(ctx, req.MarketId)
	if sdkErr != nil {
		return nil, sdkErr
	}

	return &types.QueryPriceResponse{
		Price: types.CurrentPriceResponse{
			MarketID: currentPrice.MarketID,
			Price:    currentPrice.Price,
		}}, nil

}
func (k Keeper) Prices(c context.Context, req *types.QueryPricesRequest) (*types.QueryPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var currentPrices []types.CurrentPriceResponse

	for _, cp := range k.GetCurrentPrices(ctx) {
		currentPrices = append(currentPrices, types.CurrentPriceResponse{
			MarketID: cp.MarketID,
			Price:    cp.Price,
		})
	}

	return &types.QueryPricesResponse{
		Prices: currentPrices,
	}, nil
}

func (k Keeper) RawPrices(c context.Context, req *types.QueryRawPricesRequest) (*types.QueryRawPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, found := k.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}

	var prices []types.PostedPriceResponse
	for _, rp := range k.GetRawPrices(ctx, req.MarketId) {
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

func (k Keeper) Oracles(c context.Context, req *types.QueryOraclesRequest) (*types.QueryOraclesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	oracles, err := k.GetOracles(ctx, req.MarketId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "invalid market ID")
	}

	return &types.QueryOraclesResponse{
		Oracles: oracles,
	}, nil
}

func (k Keeper) Markets(c context.Context, req *types.QueryMarketsRequest) (*types.QueryMarketsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	markets := k.GetMarkets(ctx)

	return &types.QueryMarketsResponse{
		Markets: markets,
	}, nil
}
