package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// CalculateFees returns the fees accumulated since fees were last calculated based on
// the input amount of outstanding debt (principal) and the number of periods (seconds) that have passed
func (k Keeper) CalculateFees(ctx sdk.Context, principal sdk.Coins, periods sdk.Int, denom string) sdk.Coins {
	newFees := sdk.NewCoins()
	for _, pc := range principal {
		// how fees are calculated:
		// feesAccumulated = (outstandingDebt * (feeRate^periods)) - outstandingDebt
		// Note that since we can't do x^y using sdk.Decimal, we are converting to int and using RelativePow
		feePerSecond := k.getFeeRate(ctx, denom)
		scalar := sdk.NewInt(1000000000000000000)
		feeRateInt := feePerSecond.Mul(sdk.NewDecFromInt(scalar)).TruncateInt()
		accumulator := sdk.NewDecFromInt(types.RelativePow(feeRateInt, periods, scalar)).Mul(sdk.SmallestDec())
		feesAccumulated := (sdk.NewDecFromInt(pc.Amount).Mul(accumulator)).Sub(sdk.NewDecFromInt(pc.Amount))
		// TODO this will always round down, causing precision loss between the sum of all fees in CDPs and surplus coins in liquidator account
		newFees = newFees.Add(sdk.NewCoins(sdk.NewCoin(pc.Denom, feesAccumulated.TruncateInt())))
	}
	return newFees
}

// TODO UpdateFeesForRiskyCdps
// Select cdps with 10% of the liquidation ratio (ie. 150% * 0.1 = 15% = 165% or lower)
// call calculate fees for each of those cdps
// store the result of calculate fees in cdp.AccumulatedFees
// set cdp.FeesLastUpdated to the current block time (ie. ctx.BlockTime())

// UpdateFeesForRiskyCdps calculates fees for risky CDPs
// The overall logic is first select the CDPs with 10% of the liquidation ratio
// Then we call calculate fees on each of those CDPs
// Next we store the result of the fees in the cdp.AccumulatedFees field
// Finally we set the cdp.FeesUpdated time to the current block time (ctx.BlockTime()) since that
// is when we made the update
// TODO - this method signature should only take (ctx sdk.Context, cp types.Params) as parameters, need
// to fix / remove others
// TODO - question - is types.CollateralParam the correct type?
func (k Keeper) UpdateFeesForRiskyCdps(ctx sdk.Context, cp types.CollateralParam) {

	// first calculate the target ratio based on liquidation ratio plus ten percent
	value, err := sdk.NewDecFromStr("1.1")
	if err != nil {
		fmt.Errorf("got error: %s", err)
	}
	targetRatio := k.getLiquidationRatio(ctx, cp.Denom).Mul(value) // corresponds to 110% of the liquidation ratio

	// now iterate over all the cdps based on collateral ratio
	k.IterateCdpsByCollateralRatio(ctx, cp.Denom, targetRatio, func(cdp types.CDP) bool {
		additionalFees := k.CalculateFees(ctx, cdp.Principal, periods, cp.Denom)
		cdp.AccumulatedFees.Add(additionalFees)
		cdp.FeesUpdated = ctx.BlockTime()
		return false // TODO - is this the correct thing to return??
	})
	// this function does not return anything
}

// IncrementTotalPrincipal increments the total amount of debt that has been drawn with that collateral type
func (k Keeper) IncrementTotalPrincipal(ctx sdk.Context, collateralDenom string, principal sdk.Coins) {
	for _, pc := range principal {
		total := k.GetTotalPrincipal(ctx, collateralDenom, pc.Denom)
		total = total.Add(pc.Amount)
		k.SetTotalPrincipal(ctx, collateralDenom, pc.Denom, total)
	}
}

// DecrementTotalPrincipal decrements the total amount of debt that has been drawn for a particular collateral type
func (k Keeper) DecrementTotalPrincipal(ctx sdk.Context, collateralDenom string, principal sdk.Coins) {
	for _, pc := range principal {
		total := k.GetTotalPrincipal(ctx, collateralDenom, pc.Denom)
		total = total.Sub(pc.Amount)
		if total.IsNegative() {
			// can happen in tests due to rounding errors in fee calculation
			total = sdk.ZeroInt()
		}
		k.SetTotalPrincipal(ctx, collateralDenom, pc.Denom, total)
	}
}

// GetTotalPrincipal returns the total amount of principal that has been drawn for a particular collateral
func (k Keeper) GetTotalPrincipal(ctx sdk.Context, collateralDenom string, principalDenom string) (total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	bz := store.Get([]byte(collateralDenom + principalDenom))
	if bz == nil {
		k.SetTotalPrincipal(ctx, collateralDenom, principalDenom, sdk.ZeroInt())
		return sdk.ZeroInt()
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &total)
	return total
}

// SetTotalPrincipal sets the total amount of principal that has been drawn for the input collateral
func (k Keeper) SetTotalPrincipal(ctx sdk.Context, collateralDenom string, principalDenom string, total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	store.Set([]byte(collateralDenom+principalDenom), k.cdc.MustMarshalBinaryLengthPrefixed(total))
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	b := store.Get([]byte{})
	if b == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &blockTime)
	return blockTime, true
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryLengthPrefixed(blockTime))
}
