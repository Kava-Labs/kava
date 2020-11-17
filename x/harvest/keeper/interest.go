package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/harvest/types"
)

var (
	scalingFactor  = 1e18
	secondsPerYear = 31536000
)

// ApplyInterestRateUpdates translates the current interest rate models from the params to the store
func (k Keeper) ApplyInterestRateUpdates(ctx sdk.Context) {
	denomSet := map[string]bool{}

	params := k.GetParams(ctx)
	for _, mm := range params.MoneyMarkets {
		model, found := k.GetInterestRateModel(ctx, mm.Denom)
		if !found {
			k.AccrueInterest(ctx, mm.Denom)
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
			continue
		}
		if !model.Equal(mm.InterestRateModel) {
			k.AccrueInterest(ctx, mm.Denom)
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
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

// AccrueInterest applies accrued interest to total borrows and reserves by calculating
// interest from the last checkpoint time and writing the updated values to the store.
func (k Keeper) AccrueInterest(ctx sdk.Context, denom string) error {
	previousAccrualTime, found := k.GetPreviousAccrualTime(ctx, denom)
	if !found {
		// TODO: Is this the proper place to SetPreviousAccrualTime's initial value?
		k.SetPreviousAccrualTime(ctx, denom, ctx.BlockTime())
		// return sdkerrors.Wrap(types.ErrPreviousAccrualTimeNotFound, "")
	}

	timeElapsed := ctx.BlockTime().Unix() - previousAccrualTime.Unix()
	if timeElapsed == 0 {
		return nil
	}

	// Get available harvest module account cash on hand
	cashPrior := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins().AmountOf(denom)

	// Get prior borrows
	borrowsPrior, foundBorrowsPrior := k.GetTotalBorrows(ctx, denom)
	if !foundBorrowsPrior {
		newBorrowsPrior := sdk.NewCoin(denom, sdk.ZeroInt())
		k.SetTotalBorrows(ctx, denom, sdk.NewCoin(denom, sdk.ZeroInt()))
		borrowsPrior = newBorrowsPrior
	}

	reservesPrior, foundReservesPrior := k.GetTotalReserves(ctx, denom)
	if !foundReservesPrior {
		newReservesPrior := sdk.NewCoin(denom, sdk.ZeroInt())
		k.SetTotalReserves(ctx, denom, newReservesPrior)
		reservesPrior = newReservesPrior
	}

	borrowIndexPrior, foundBorrowIndexPrior := k.GetBorrowIndex(ctx, denom)
	if !foundBorrowIndexPrior {
		newBorrowIndexPrior := sdk.MustNewDecFromStr("1.0")
		k.SetBorrowIndex(ctx, denom, newBorrowIndexPrior)
		borrowIndexPrior = newBorrowIndexPrior
	}

	// TODO: Add reserve_factor param to each MoneyMarket. Reserve factor is the % of protocol fees.
	// reserveFactor := k.GetReserveFactor(ctx, denom)
	reserveFactor := sdk.MustNewDecFromStr("1.01")

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
	borrowIndexNew := borrowIndexPrior.Mul(interestFactor)

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

	// TODO: Consider using scaling factor here
	totalSupply := cash.Add(borrows).Sub(reserves)
	return borrows.Quo(totalSupply)
}

// CalculateInterestFactor calculates the simple interest scaling factor,
// which is equal to: (per-second interest rate * number of seconds elapsed)
func CalculateInterestFactor(perSecondInterestRate sdk.Dec, secondsElapsed sdk.Int) sdk.Dec {
	// TODO: Consider overflow panics and optimize calculations
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

// APYToSPY converts the input annual interest rate. For example, 10% apy would be passed as 1.10.
func APYToSPY(apy sdk.Dec) (sdk.Dec, error) {
	root, err := apy.ApproxRoot(uint64(secondsPerYear))
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return root, nil
}
