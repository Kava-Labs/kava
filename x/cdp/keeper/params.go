package keeper

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx context.Context) types.Params {
	var p types.Params
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramSubspace.GetParamSetIfExists(sdkCtx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramSubspace.SetParamSet(sdkCtx, &params)
}

// GetCollateral returns the collateral param with corresponding denom
func (k Keeper) GetCollateral(ctx context.Context, collateralType string) (types.CollateralParam, bool) {
	// print stack from what function it was called from
	//fmt.Println("stack", string(debug.Stack()))
	params := k.GetParams(ctx)
	//fmt.Println("params.CollateralParams", params.CollateralParams)
	for _, cp := range params.CollateralParams {
		if cp.Type == collateralType {
			return cp, true
		}
	}
	return types.CollateralParam{}, false
}

// GetCollateralTypes returns an array of collateral types
func (k Keeper) GetCollateralTypes(ctx context.Context) []string {
	params := k.GetParams(ctx)
	var denoms []string
	for _, cp := range params.CollateralParams {
		denoms = append(denoms, cp.Type)
	}
	return denoms
}

// GetDebtParam returns the debt param with matching denom
func (k Keeper) GetDebtParam(ctx context.Context, denom string) (types.DebtParam, bool) {
	dp := k.GetParams(ctx).DebtParam
	if dp.Denom == denom {
		return dp, true
	}
	return types.DebtParam{}, false
}

func (k Keeper) getSpotMarketID(ctx context.Context, collateralType string) string {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.SpotMarketID
}

func (k Keeper) getliquidationMarketID(ctx context.Context, collateralType string) string {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.LiquidationMarketID
}

func (k Keeper) getLiquidationRatio(ctx context.Context, collateralType string) sdkmath.LegacyDec {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.LiquidationRatio
}

func (k Keeper) getLiquidationPenalty(ctx context.Context, collateralType string) sdkmath.LegacyDec {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.LiquidationPenalty
}

func (k Keeper) getAuctionSize(ctx context.Context, collateralType string) sdkmath.Int {
	cp, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	return cp.AuctionSize
}

// GetFeeRate returns the per second fee rate for the input denom
func (k Keeper) getFeeRate(ctx context.Context, collateralType string) (fee sdkmath.LegacyDec) {
	collalateralParam, found := k.GetCollateral(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("could not get fee rate for %s, collateral not found", collateralType))
	}
	return collalateralParam.StabilityFee
}
