package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/auction/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetAuction:
			return queryAuction(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetAuctions:
			return queryAuctions(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryNextAuctionID:
			return queryNextAuctionID(ctx, req, keeper, legacyQuerierCdc)
		default:
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

// query params in the auction store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAuction(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryAuctionParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	auction, ok := k.GetAuction(ctx, params.AuctionID)
	if !ok {
		return nil, errorsmod.Wrap(types.ErrAuctionNotFound, fmt.Sprintf("%d", params.AuctionID))
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, auction)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryAuctions(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryAllAuctionParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	unfilteredAuctions := k.GetAllAuctions(ctx)
	auctions := filterAuctions(ctx, unfilteredAuctions, params, legacyQuerierCdc)
	if auctions == nil {
		auctions = []types.Auction{}
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, auctions)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryNextAuctionID(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	nextAuctionID, _ := keeper.GetNextAuctionID(ctx)

	bz, err := legacyQuerierCdc.MarshalJSON(nextAuctionID)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

// filterAuctions retrieves auctions filtered by a given set of params.
// If no filters are provided, all auctions will be returned in paginated form.
func filterAuctions(ctx sdk.Context, auctions []types.Auction, params types.QueryAllAuctionParams, legacyQuerierCdc *codec.LegacyAmino) []types.Auction {
	filteredAuctions := make([]types.Auction, 0, len(auctions))
	for _, auc := range auctions {
		isMatch := auctionIsMatch(auc, params)
		if isMatch {
			filteredAuctions = append(filteredAuctions, auc)
		}
	}

	start, end := client.Paginate(len(filteredAuctions), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		filteredAuctions = []types.Auction{}
	} else {
		filteredAuctions = filteredAuctions[start:end]
	}

	return filteredAuctions
}

func auctionIsMatch(auc types.Auction, params types.QueryAllAuctionParams) bool {
	matchType, matchOwner, matchDenom, matchPhase := true, true, true, true

	// match auction type (if supplied)
	if len(params.Type) > 0 {
		matchType = auc.GetType() == params.Type
	}

	// match auction owner (if supplied)
	if len(params.Owner) > 0 {
		if cAuc, ok := auc.(*types.CollateralAuction); ok {
			foundOwnerAddr := false
			for _, addr := range cAuc.GetLotReturns().Addresses {
				if addr.Equals(params.Owner) {
					foundOwnerAddr = true
					break
				}
			}
			if !foundOwnerAddr {
				matchOwner = false
			}
		}
	}

	// match auction denom (if supplied)
	if len(params.Denom) > 0 {
		matchDenom = auc.GetBid().Denom == params.Denom || auc.GetLot().Denom == params.Denom
	}

	// match auction phase (if supplied)
	if len(params.Phase) > 0 {
		matchPhase = auc.GetPhase() == params.Phase
	}

	if matchType && matchOwner && matchDenom && matchPhase {
		return true
	}
	return false
}
