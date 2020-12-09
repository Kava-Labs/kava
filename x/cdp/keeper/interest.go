package keeper

import (
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
		k.SetPreviousAccrualTime(ctx, ctype, ctx.BlockTime())
		return nil
	}

	interestFactorPrior, foundInterestFactorPrior := k.GetInterestFactor(ctx, ctype)
	if !foundInterestFactorPrior {
		k.SetInterestFactor(ctx, ctype, sdk.OneDec())
		// set previous accrual time exit early because interest accumulated will be zero
		k.SetPreviousAccrualTime(ctx, ctype, ctx.BlockTime())
		return nil
	}

	borrowRateSpy := k.getFeeRate(ctx, ctype)
	if borrowRateSpy.Equal(sdk.OneDec()) {
		k.SetPreviousAccrualTime(ctx, ctype, ctx.BlockTime())
		return nil
	}
	interestFactor := CalculateInterestFactor(borrowRateSpy, sdk.NewInt(timeElapsed))
	interestAccumulated := (interestFactor.Mul(totalPrincipalPrior.ToDec())).RoundInt().Sub(totalPrincipalPrior)
	if interestAccumulated.IsZero() {
		// in the case accumulated interest rounds to zero, exit early without updating accrual time
		return nil
	}
	err := k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), sdk.NewCoin(types.DefaultStableDenom, interestAccumulated))
	if err != nil {
		return err
	}

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
