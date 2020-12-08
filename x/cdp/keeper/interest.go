package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

var (
	scalingFactor  = 1e18
	secondsPerYear = 31536000
)

// AccumulateInterest calculates the new interest that has accrued for the input collateral type based on the total amount of principal
// that has been created with that collateral type and the amount of time that has passed since interest was last accumulated
func (k Keeper) AccumulateInterest(ctx sdk.Context, ctype string) error {
	previousAccrualTime, found := k.GetPreviousAccrualTime(ctx, ctype)
	if !found {
		k.SetPreviousAccrualTime(ctx, ctype, ctx.BlockTime())
		return nil
	}

	timeElapsed := ctx.BlockTime().Unix() - previousAccrualTime.Unix()
	if timeElapsed == 0 {
		return nil
	}

	totalPrincipalPrior := k.GetTotalPrincipal(ctx, ctype, types.DefaultStableDenom)
	if totalPrincipalPrior.IsZero() || totalPrincipalPrior.IsNegative() {
		return nil
	}

	interestFactorPrior, foundInterestFactorPrior := k.GetInterestFactor(ctx, ctype)
	if !foundInterestFactorPrior {
		interestFactorPrior = sdk.OneDec()
		k.SetInterestFactor(ctx, ctype, interestFactorPrior)
		return nil
	}

	borrowRateSpy := k.getFeeRate(ctx, ctype)
	if borrowRateSpy.Equal(sdk.OneDec()) {
		return nil
	}
	interestFactor := CalculateInterestFactor(borrowRateSpy, sdk.NewInt(timeElapsed))
	interestAccumulated := (interestFactor.Mul(totalPrincipalPrior.ToDec())).RoundInt().Sub(totalPrincipalPrior)
	k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), sdk.NewCoin(types.DefaultStableDenom, interestAccumulated))

	interestFactorNew := interestFactorPrior.Mul(interestFactor)
	totalPrincipalNew := totalPrincipalPrior.Add(interestAccumulated)

	k.SetTotalPrincipal(ctx, ctype, types.DefaultStableDenom, totalPrincipalNew)
	k.SetInterestFactor(ctx, ctype, interestFactorNew)
	k.SetPreviousAccrualTime(ctx, ctype, ctx.BlockTime())

	return nil
}

// CalculateInterestFactor calculates the simple interest scaling factor,
// which is equal to: (per-second interest rate ** number of seconds elapsed)
// Will return 1.000x, multiply by principal to get new principal with added interest
func CalculateInterestFactor(perSecondInterestRate sdk.Dec, secondsElapsed sdk.Int) sdk.Dec {
	scalingFactorUint := sdk.NewUint(uint64(scalingFactor))
	scalingFactorInt := sdk.NewInt(int64(scalingFactor))

	// Convert per-second interest rate to a uint scaled by 1e18
	interestMantissa := sdk.NewUint(perSecondInterestRate.MulInt(scalingFactorInt).RoundInt().Uint64())
	// Convert seconds elapsed to uint (*not scaled*)
	secondsElapsedUint := sdk.NewUint(secondsElapsed.Uint64())
	// Calculate the interest factor as a uint scaled by 1e18
	interestFactorMantissa := sdk.RelativePow(interestMantissa, secondsElapsedUint, scalingFactorUint)

	// Convert interest factor to an unscaled sdk.Dec
	return sdk.NewDecFromBigInt(interestFactorMantissa.BigInt()).QuoInt(scalingFactorInt)
}

// UpdateFeesForAllCdps updates the fees for each of the CDPs
func (k Keeper) UpdateFeesForAllCdps(ctx sdk.Context, collateralType string) error {
	var iterationErr error
	k.IterateCdpsByCollateralType(ctx, collateralType, func(cdp types.CDP) bool {
		oldCollateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())
		// periods = bblock timestamp - fees updated
		periods := sdk.NewInt(ctx.BlockTime().Unix()).Sub(sdk.NewInt(cdp.FeesUpdated.Unix()))

		newFees := k.CalculateFees(ctx, cdp.Principal, periods, collateralType)

		// exit without updating fees if amount has rounded down to zero
		// cdp will get updated next block when newFees, newFeesSavings, newFeesSurplus >0
		if newFees.IsZero() {
			return false
		}

		dp, found := k.GetDebtParam(ctx, cdp.Principal.Denom)
		if !found {
			return false
		}
		savingsRate := dp.SavingsRate

		newFeesSavings := sdk.NewDecFromInt(newFees.Amount).Mul(savingsRate).RoundInt()
		newFeesSurplus := newFees.Amount.Sub(newFeesSavings)

		// similar to checking for rounding to zero of all fees, but in this case we
		// need to handle cases where we expect surplus or savings fees to be zero, namely
		// if newFeesSavings = 0, check if savings rate is not zero
		// if newFeesSurplus = 0, check if savings rate is not one
		if (newFeesSavings.IsZero() && !savingsRate.IsZero()) || (newFeesSurplus.IsZero() && !savingsRate.Equal(sdk.OneDec())) {
			return false
		}

		// mint debt coins to the cdp account
		err := k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), newFees)
		if err != nil {
			iterationErr = err
			return true
		}

		previousDebt := k.GetTotalPrincipal(ctx, cdp.Type, dp.Denom)
		newDebt := previousDebt.Add(newFees.Amount)
		k.SetTotalPrincipal(ctx, cdp.Type, dp.Denom, newDebt)

		// mint surplus coins divided between the liquidator and savings module accounts.
		err = k.supplyKeeper.MintCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, newFeesSurplus)))
		if err != nil {
			iterationErr = err
			return true
		}

		err = k.supplyKeeper.MintCoins(ctx, types.SavingsRateMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, newFeesSavings)))
		if err != nil {
			iterationErr = err
			return true
		}

		// now add the new fees to the accumulated fees for the cdp
		cdp.AccumulatedFees = cdp.AccumulatedFees.Add(newFees)

		// and set the fees updated time to the current block time since we just updated it
		cdp.FeesUpdated = ctx.BlockTime()
		collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())
		k.RemoveCdpCollateralRatioIndex(ctx, cdp.Type, cdp.ID, oldCollateralToDebtRatio)
		err = k.SetCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio)
		if err != nil {
			iterationErr = err
			return true
		}
		return false // this returns true when you want to stop iterating. Since we want to iterate through all we return false
	})
	return iterationErr
}

// CalculateFees returns the fees accumulated since fees were last calculated based on
// the input amount of outstanding debt (principal) and the number of periods (seconds) that have passed
func (k Keeper) CalculateFees(ctx sdk.Context, principal sdk.Coin, periods sdk.Int, collateralType string) sdk.Coin {
	// how fees are calculated:
	// feesAccumulated = (outstandingDebt * (feeRate^periods)) - outstandingDebt
	// Note that since we can't do x^y using sdk.Decimal, we are converting to int and using RelativePow
	feePerSecond := k.getFeeRate(ctx, collateralType)
	scalar := sdk.NewInt(1000000000000000000)
	feeRateInt := feePerSecond.Mul(sdk.NewDecFromInt(scalar)).TruncateInt()
	accumulator := sdk.NewDecFromInt(types.RelativePow(feeRateInt, periods, scalar)).Mul(sdk.SmallestDec())
	feesAccumulated := (sdk.NewDecFromInt(principal.Amount).Mul(accumulator)).Sub(sdk.NewDecFromInt(principal.Amount))
	newFees := sdk.NewCoin(principal.Denom, feesAccumulated.TruncateInt())
	return newFees
}

// IncrementTotalPrincipal increments the total amount of debt that has been drawn with that collateral type
func (k Keeper) IncrementTotalPrincipal(ctx sdk.Context, collateralType string, principal sdk.Coin) {
	total := k.GetTotalPrincipal(ctx, collateralType, principal.Denom)
	total = total.Add(principal.Amount)
	k.SetTotalPrincipal(ctx, collateralType, principal.Denom, total)
}

// DecrementTotalPrincipal decrements the total amount of debt that has been drawn for a particular collateral type
func (k Keeper) DecrementTotalPrincipal(ctx sdk.Context, collateralType string, principal sdk.Coin) {
	total := k.GetTotalPrincipal(ctx, collateralType, principal.Denom)
	// NOTE: negative total principal can happen in tests due to rounding errors
	// in fee calculation
	total = sdk.MaxInt(total.Sub(principal.Amount), sdk.ZeroInt())
	k.SetTotalPrincipal(ctx, collateralType, principal.Denom, total)
}

// GetTotalPrincipal returns the total amount of principal that has been drawn for a particular collateral
func (k Keeper) GetTotalPrincipal(ctx sdk.Context, collateralType, principalDenom string) (total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	bz := store.Get([]byte(collateralType + principalDenom))
	if bz == nil {
		k.SetTotalPrincipal(ctx, collateralType, principalDenom, sdk.ZeroInt())
		return sdk.ZeroInt()
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &total)
	return total
}

// SetTotalPrincipal sets the total amount of principal that has been drawn for the input collateral
func (k Keeper) SetTotalPrincipal(ctx sdk.Context, collateralType, principalDenom string, total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	_, found := k.GetCollateralTypePrefix(ctx, collateralType)
	if !found {
		panic(fmt.Sprintf("collateral not found: %s", collateralType))
	}
	store.Set([]byte(collateralType+principalDenom), k.cdc.MustMarshalBinaryLengthPrefixed(total))
}
