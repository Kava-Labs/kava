package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetCollateral returns the collateral param with corresponding denom
func (k Keeper) GetCollateral(ctx sdk.Context, denom string) (types.CollateralParam, bool) {
	params := k.GetParams(ctx)
	for _, cp := range params.CollateralParams {
		if cp.Denom == denom {
			return cp, true
		}
	}
	return types.CollateralParam{}, false
}

// GetDebt returns the debt param with matching denom
func (k Keeper) GetDebt(ctx sdk.Context, denom string) (types.DebtParam, bool) {
	params := k.GetParams(ctx)
	for _, dp := range params.DebtParams {
		if dp.Denom == denom {
			return dp, true
		}
	}
	return types.DebtParam{}, false
}

// GetDenomPrefix returns the prefix of the matching denom
func (k Keeper) GetDenomPrefix(ctx sdk.Context, denom string) (byte, bool) {
	params := k.GetParams(ctx)
	for _, cp := range params.CollateralParams {
		if cp.Denom == denom {
			return cp.Prefix, true
		}
	}
	return 0xff, false
}

func (k Keeper) getStabilityFee(ctx sdk.Context, denom string) sdk.Dec {
	cp, found := k.GetCollateral(ctx, denom)
	if !found {
		panic(fmt.Sprintf("no collateral found for %s", denom))
	}
	return cp.StabilityFee
}

func (k Keeper) getDenomFromByte(ctx sdk.Context, db byte) string {
	params := k.GetParams(ctx)
	for _, cp := range params.CollateralParams {
		if cp.Prefix == db {
			return cp.Denom
		}
	}
	panic(fmt.Sprintf("no collateral denom with prefix %b", db))
}

func (k Keeper) getMarketID(ctx sdk.Context, denom string) string {
	cp, found := k.GetCollateral(ctx, denom)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", denom))
	}
	return cp.MarketID
}

func (k Keeper) getLiquidationRatio(ctx sdk.Context, denom string) sdk.Dec {
	cp, found := k.GetCollateral(ctx, denom)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", denom))
	}
	return cp.LiquidationRatio
}
