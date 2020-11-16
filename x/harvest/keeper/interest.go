package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/harvest/types"
)

// ApplyInterestRateUpdates translates the current interest rate models from the params to the store
func (k Keeper) ApplyInterestRateUpdates(ctx sdk.Context) {
	denomSet := map[string]bool{}

	params := k.GetParams(ctx)
	for _, mm := range params.MoneyMarkets {
		model, found := k.GetInterestRateModel(ctx, mm.Denom)
		if !found {
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
			// TODO: set AccrueInterest variables
			continue
		}
		if !model.Equal(mm.InterestRateModel) {
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
			// TODO: Set AccrueInterest variables
		}
		denomSet[mm.Denom] = true
	}

	k.IterateInterestRateModels(ctx, func(denom string, i types.InterestRateModel) bool {
		if !denomSet[denom] {
			k.DeleteInterestRateModel(ctx, denom)
		}
		return false
	})
}

// AccrueInterest applies accrued interest to total borrows and reserves
// calculates interest from the last checkpoint time and writes the updated values to the store
func (k Keeper) AccrueInterest(ctx sdk.Context, denom string) error {
	previousAccrualTime, found := k.GetPreviousAccrualTime(ctx, denom)
	if !found {
		return sdkerrors.Wrap(types.ErrPreviousAccrualTimeNotFound, "")
	}
	timeElapsed := ctx.BlockTime().Unix() - previousAccrualTime.Unix()
	// short-circuit if no time has passed
	if timeElapsed == 0 {
		return nil
	}

	cashPrior := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins().AmountOf(denom)
	borrowsPrior, _ := k.GetTotalBorrows(ctx, denom)
	reservesPrior, _ := k.GetTotalReserves(ctx, denom)
	borrowIndexPrior, _ := k.GetBorrowIndex(ctx, denom)

	// TODO: Reserve factor is protocol % fees (param set on each money market)
	// reserveFactor := k.GetReserveFactor(ctx, denom)
	reserveFactor := sdk.NewDec(5.0)

	// GetBorrowRate calculates the current interest rate based on utilization (the fraction of supply that has been borrowed)
	borrowRateApy, err := k.CalculateBorrowRate(ctx, denom, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowsPrior.Amount), sdk.NewDecFromInt(reservesPrior.Amount))
	if err != nil {
		return err
	}

	borrowRateSpy, err := APYToSPY(borrowRateApy)
	if err != nil {
		return err
	}

	interestFactor := CalculateInterestFactor(borrowRateSpy, sdk.NewInt(timeElapsed))
	interestAccumulated := interestFactor.Mul(sdk.NewDecFromInt(borrowsPrior.Amount)).TruncateInt()
	totalBorrowsNew := borrowsPrior.Add(sdk.NewCoin(denom, interestAccumulated))
	totalReservesNew := reservesPrior.Add(sdk.NewCoin(denom, sdk.NewDecFromInt(interestAccumulated).Mul(reserveFactor).TruncateInt()))
	borrowIndexNew := borrowIndexPrior.Mul(interestFactor).Add(borrowIndexPrior)

	k.SetBorrowIndex(ctx, denom, borrowIndexNew)
	k.SetTotalBorrows(ctx, denom, totalBorrowsNew)
	k.SetTotalReserves(ctx, denom, totalReservesNew)
	k.SetPreviousAccrualTime(ctx, denom, ctx.BlockTime())
	return nil
}

// CalculateBorrowRate calculates the borrow rate
func (k Keeper) CalculateBorrowRate(ctx sdk.Context, denom string, cash, borrows, reserves sdk.Dec) (sdk.Dec, error) {
	utilRatio := CalculateUtilizationRatio(cash, borrows, reserves)

	moneyMarket, found := k.GetMoneyMarket(ctx, denom)
	if !found {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrMoneyMarketNotFound, "%s", denom)
	}

	// Calculate normal borrow rate (under kink)
	if utilRatio.LTE(moneyMarket.InterestRateModel.Kink) {
		return utilRatio.Mul(moneyMarket.InterestRateModel.BaseMultiplier).Add(moneyMarket.InterestRateModel.BaseRateAPY), nil
	}

	// Calculate jump borrow rate (over kink)
	normalRate := moneyMarket.InterestRateModel.Kink.Mul(moneyMarket.InterestRateModel.BaseMultiplier).Add(moneyMarket.InterestRateModel.BaseRateAPY)
	excessUtil := utilRatio.Sub(moneyMarket.InterestRateModel.Kink)
	return excessUtil.Mul(moneyMarket.InterestRateModel.JumpMultiplier).Add(normalRate), nil

}

// CalculateUtilizationRatio calculates an asset's current utilization rate
func CalculateUtilizationRatio(cash, borrows, reserves sdk.Dec) sdk.Dec {
	// Utilization rate is 0 when there are no borrows
	if borrows.Equal(sdk.ZeroDec()) {
		return sdk.ZeroDec()
	}

	// TODO: consider decimals (compound uses 1e18)
	totalSupply := cash.Add(borrows).Sub(reserves)
	return borrows.Quo(totalSupply)
}

// CalculateInterestFactor calculates the 'simple interest scaling factor' for the input per-second interest rate and the number of seconds elapsed
// notes: this doesn’t catch potential overflow panics
// there’s a lot of conversions that could probably be done more efficiently
func CalculateInterestFactor(perSecondInterestRate sdk.Dec, secondsElapsed sdk.Int) sdk.Dec {
	// the input dec is converted to a uint scaled by 1e18 - this could be stored as a module-level private var
	interestMantissa := sdk.NewUint(
		perSecondInterestRate.MulInt(sdk.NewInt(1e18)).RoundInt().Uint64())

	// the input int is convert to uint (*not scaled*)
	secondsElapsedUint := sdk.NewUint(secondsElapsed.Uint64())
	scalingFactor := sdk.NewUint(1e18)

	// calculate the interest factor as a uint scaled by 1e18
	interestFactorMantissa := sdk.RelativePow(interestMantissa, secondsElapsedUint, scalingFactor)

	//convert interest factor to an unscaled, sdk.Dec
	return sdk.NewDecFromBigInt(interestFactorMantissa.BigInt()).QuoInt(sdk.NewInt(1e18))
}

// APYToSPY converts the input annual interest rate (10% apy would be passed as 1.10) to the per-second interest rate.
func APYToSPY(apy sdk.Dec) (sdk.Dec, error) {
	secondsPerYear := uint64(31536000) // AprroxRoot takes a uint64 for whatever reason, this can be moved to a module-level private var
	root, err := apy.ApproxRoot(secondsPerYear)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return root, nil
}
