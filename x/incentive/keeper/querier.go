package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetParams:
			return queryGetParams(ctx, req, k)
		case types.QueryGetRewards:
			return queryGetRewards(ctx, req, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

// query params in the store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	// Get params
	params := k.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryGetRewards(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryRewardsParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	hasType := len(params.Type) > 0
	owner := len(params.Owner) > 0

	var usdxMintingClaims types.USDXMintingClaims
	var hardClaims types.HardLiquidityProviderClaims
	switch {
	case owner && hasType:
		if strings.ToLower(params.Type) == "usdx-minting" {
			usdxMintingClaim, foundUsdxMintingClaim := k.GetUSDXMintingClaim(ctx, params.Owner)
			if foundUsdxMintingClaim {
				usdxMintingClaims = append(usdxMintingClaims, usdxMintingClaim)
			}
		} else if strings.ToLower(params.Type) == "hard" {
			hardClaim, foundHardClaim := k.GetHardLiquidityProviderClaim(ctx, params.Owner)
			if foundHardClaim {
				hardClaims = append(hardClaims, hardClaim)
			}
		} else {
			return nil, types.ErrInvalidClaimType
		}
	case owner:
		usdxMintingClaim, foundUsdxMintingClaim := k.GetUSDXMintingClaim(ctx, params.Owner)
		if foundUsdxMintingClaim {
			usdxMintingClaims = append(usdxMintingClaims, usdxMintingClaim)
		}
		hardClaim, foundHardClaim := k.GetHardLiquidityProviderClaim(ctx, params.Owner)
		if foundHardClaim {
			hardClaims = append(hardClaims, hardClaim)
		}
	case hasType:
		if strings.ToLower(params.Type) == "usdx-minting" {
			usdxMintingClaims = k.GetAllUSDXMintingClaims(ctx)
		} else if strings.ToLower(params.Type) == "hard" {
			hardClaims = k.GetAllHardLiquidityProviderClaims(ctx)
		} else {
			return nil, types.ErrInvalidClaimType
		}
	default:
		usdxMintingClaims = k.GetAllUSDXMintingClaims(ctx)
		hardClaims = k.GetAllHardLiquidityProviderClaims(ctx)
	}

	var paginatedUsdxMintingClaims types.USDXMintingClaims
	startU, endU := client.Paginate(len(usdxMintingClaims), params.Page, params.Limit, 100)
	if startU < 0 || endU < 0 {
		paginatedUsdxMintingClaims = types.USDXMintingClaims{}
	} else {
		paginatedUsdxMintingClaims = usdxMintingClaims[startU:endU]
	}

	// TODO: use hardClaimsLimited instead of hardClaims to enforce that global query limit is not exceeded
	// remainingClaims := params.Limit - len(paginatedUsdxMintingClaims)
	// hardClaimsLimited := hardClaims[0:remainingClaims]

	var paginatedHardClaims types.HardLiquidityProviderClaims
	startH, endH := client.Paginate(len(hardClaims), params.Page, params.Limit, 100)
	if startH < 0 || endH < 0 {
		paginatedHardClaims = types.HardLiquidityProviderClaims{}
	} else {
		paginatedHardClaims = hardClaims[startH:endH]
	}

	// Marshal USDX minting claims
	bzUsdxMintingClaims, err := codec.MarshalJSONIndent(k.cdc, paginatedUsdxMintingClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	// Marshal Hard claims
	bzHardClaims, err := codec.MarshalJSONIndent(k.cdc, paginatedHardClaims)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	// Return concatenated bytes
	bz := append(bzUsdxMintingClaims[:], bzHardClaims[:]...)
	return bz, nil
}
