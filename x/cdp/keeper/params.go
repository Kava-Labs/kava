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
func (k Keeper) GetCollateral(ctx sdk.Context, collateralType string) (types.CollateralParam, bool) {
	params := k.GetParams(ctx)
	for _, cp := range params.CollateralParams {
		if cp.Type == collateralType {
			return cp, true
		}
	}
	return types.CollateralParam{}, false
}

// GetDebtParam returns the debt param with matching denom
func (k Keeper) GetDebtParam(ctx sdk.Context, denom string) (types.DebtParam, bool) {
	dp := k.GetParams(ctx).DebtParam
	if dp.Denom == denom {
		return dp, true
	}
	return types.DebtParam{}, false
}

// GetDenomPrefix returns the prefix of the matching denom
func (k Keeper) GetDenomPrefix(ctx sdk.Context, collateralType string) (byte, bool) {
	params := k.GetParams(ctx)
	for _, cp := range params.CollateralParams {
		if cp.Type == collateralType {
			return cp.Prefix, true
		}
	}
	return 0x00, false
}

// private methods assume collateral has been validated, panic if the input is invalid
func (k Keeper) getDenomFromByte(ctx sdk.Context, db byte) string {
	params := k.GetParams(ctx)
	for _, cp := range params.CollateralParams {
		if cp.Prefix == db {
			return cp.Denom
		}
	}
	panic(fmt.Sprintf("no collateral denom with prefix %b", db))
}

func (k Keeper) getSpotMarketID(ctx sdk.Context, collateralType string) string {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.SpotMarketID
}

func (k Keeper) getliquidationMarketID(ctx sdk.Context, collateralType string) string {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.LiquidationMarketID
}

func (k Keeper) getLiquidationRatio(ctx sdk.Context, collateralType string) sdk.Dec {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.LiquidationRatio
}

func (k Keeper) getLiquidationPenalty(ctx sdk.Context, collateralType string) sdk.Dec {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.LiquidationPenalty
}

func (k Keeper) getAuctionSize(ctx sdk.Context, collateralType string) sdk.Int {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.AuctionSize
}

// GetFeeRate returns the per second fee rate for the input denom
func (k Keeper) getFeeRate(ctx sdk.Context, collateralType string) (fee sdk.Dec) {
	collalateralParam, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("could not get fee rate for %s, collateral not found", collateralType))
	}
	return collalateralParam.StabilityFee
}
