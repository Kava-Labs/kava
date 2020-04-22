package keeper

import (
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
		newFees = newFees.Add(sdk.NewCoin(pc.Denom, feesAccumulated.TruncateInt()))
	}
	return newFees
}

// UpdateFeesForAllCdps updates the fees for each of the CDPs
func (k Keeper) UpdateFeesForAllCdps(ctx sdk.Context, collateralDenom string) error {

	k.IterateCdpsByDenom(ctx, collateralDenom, func(cdp types.CDP) bool {

		oldCollateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees...))
		periods := sdk.NewInt(ctx.BlockTime().Unix()).Sub(sdk.NewInt(cdp.FeesUpdated.Unix()))

		newFees := k.CalculateFees(ctx, cdp.Principal, periods, collateralDenom)

		if newFees.IsZero() {
			return false
		}

		// note - only works if principal length is one
		for _, dc := range cdp.Principal {
			dp, found := k.GetDebtParam(ctx, dc.Denom)
			if !found {
				return false
			}
			savingsRate := dp.SavingsRate

			newFeesSavings := sdk.NewDecFromInt(newFees.AmountOf(dp.Denom)).Mul(savingsRate).RoundInt()
			newFeesSurplus := newFees.AmountOf(dp.Denom).Sub(newFeesSavings)

			if newFeesSavings.IsZero() || newFeesSurplus.IsZero() {
				return false
			}
			// mint debt coins to the cdp account
			k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), newFees)
			previousDebt := k.GetTotalPrincipal(ctx, collateralDenom, dp.Denom)
			feeCoins := sdk.NewCoins(sdk.NewCoin(dp.Denom, previousDebt))
			k.SetTotalPrincipal(ctx, collateralDenom, dp.Denom, feeCoins.Add(newFees...).AmountOf(dp.Denom))

			// mint surplus coins divided between the liquidator and savings module accounts.
			k.supplyKeeper.MintCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, newFeesSurplus)))
			k.supplyKeeper.MintCoins(ctx, types.SavingsRateMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, newFeesSavings)))
		}

		// now add the new fees fees to the accumulated fees for the cdp
		cdp.AccumulatedFees = cdp.AccumulatedFees.Add(newFees...)

		// and set the fees updated time to the current block time since we just updated it
		cdp.FeesUpdated = ctx.BlockTime()
		collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees...))
		k.RemoveCdpCollateralRatioIndex(ctx, cdp.Collateral[0].Denom, cdp.ID, oldCollateralToDebtRatio)
		k.SetCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio)
		return false // this returns true when you want to stop iterating. Since we want to iterate through all we return false
	})
	return nil
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
