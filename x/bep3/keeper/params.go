package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// GetParams returns the total set of bep3 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the bep3 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetChain returns the chain param with corresponding chain ID
func (k Keeper) GetChain(ctx sdk.Context, chainID string) (types.ChainParam, bool) {
	params := k.GetParams(ctx)
	for _, cp := range params.Chains {
		if cp.ChainID == chainID {
			return cp, true
		}
	}
	return types.ChainParam{}, false
}

// GetChainAssets returns a list containing a chain's supported assets
func (k Keeper) GetChainAssets(ctx sdk.Context, chainID string) (types.AssetParams, bool) {
	chain, found := k.GetChain(ctx, chainID)
	if !found {
		return types.AssetParams{}, false
	}
	return chain.SupportedAssets, len(chain.SupportedAssets) > 0
}

// GetAssetByDenom returns an asset a specific chain by its denom
func (k Keeper) GetAssetByDenom(ctx sdk.Context, chainID string, denom string) (types.AssetParam, bool) {
	chain, found := k.GetChain(ctx, chainID)
	if !found {
		return types.AssetParam{}, false
	}
	for _, ap := range chain.SupportedAssets {
		if ap.Denom == denom {
			return ap, true
		}
	}
	return types.AssetParam{}, false
}

// GetAssetByCoinID returns an asset a specific chain by its coin ID
func (k Keeper) GetAssetByCoinID(ctx sdk.Context, chainID string, coinID string) (types.AssetParam, bool) {
	chain, found := k.GetChain(ctx, chainID)
	if !found {
		return types.AssetParam{}, false
	}
	for _, ap := range chain.SupportedAssets {
		if ap.CoinID == coinID {
			return ap, true
		}
	}
	return types.AssetParam{}, false
}
