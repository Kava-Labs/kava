package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/hard/types"
)

var (
	scalingFactor  = 1e18
	secondsPerYear = 31536000
)

// ApplyInterestRateUpdates translates the current interest rate models from the params to the store,
// with each money market accruing interest.
func (k Keeper) ApplyInterestRateUpdates(ctx sdk.Context) {
	denomSet := map[string]bool{}

	params := k.GetParams(ctx)
	for _, mm := range params.MoneyMarkets {
		// Set any new money markets in the store
		moneyMarket, found := k.GetMoneyMarket(ctx, mm.Denom)
		if !found {
			moneyMarket = mm
			k.SetMoneyMarket(ctx, mm.Denom, moneyMarket)
		}

		// Accrue interest according to the current money markets in the store
		err := k.AccrueInterest(ctx, mm.Denom)
		if err != nil {
			panic(err)
		}

		// Update the interest rate in the store if the params have changed
		if !moneyMarket.Equal(mm) {
			k.SetMoneyMarket(ctx, mm.Denom, mm)
		}
		denomSet[mm.Denom] = true
	}

	// Edge case: money markets removed from params that still exist in the store
	k.IterateMoneyMarkets(ctx, func(denom string, i types.MoneyMarket) bool {
		if !denomSet[denom] {
			// Accrue interest according to current store money market
			err := k.AccrueInterest(ctx, denom)
			if err != nil {
				panic(err)
			}

			// Delete the money market from the store
			k.DeleteMoneyMarket(ctx, denom)
		}
		return false
	})
}

// AccrueInterest applies accrued interest to total borrows and reserves by calculating
// interest from the last checkpoint time and writing the updated values to the store.
func (k Keeper) AccrueInterest(ctx sdk.Context, denom string) error {

	previousAccrualTime, found := k.GetPreviousAccrualTime(ctx, denom)
	if !found {
		k.SetPreviousAccrualTime(ctx, denom, ctx.BlockTime())
		return nil
	}

	timeElapsed := ctx.BlockTime().Unix() - previousAccrualTime.Unix()
	if timeElapsed == 0 {
		return nil
	}

	// Get available hard module account cash on hand
	cashPrior := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins().AmountOf(denom)

	// Get prior borrows
	borrowsPrior := sdk.NewCoin(denom, sdk.ZeroInt())
	borrowCoinsPrior, foundBorrowCoinsPrior := k.GetBorrowedCoins(ctx)
	if foundBorrowCoinsPrior {
		borrowsPrior = sdk.NewCoin(denom, borrowCoinsPrior.AmountOf(denom))
	}

	reservesPrior, foundReservesPrior := k.GetTotalReserves(ctx, denom)
	if !foundReservesPrior {
		newReservesPrior := sdk.NewCoin(denom, sdk.ZeroInt())
		k.SetTotalReserves(ctx, denom, newReservesPrior)
		reservesPrior = newReservesPrior
	}

	borrowInterestFactorPrior, foundBorrowInterestFactorPrior := k.GetBorrowInterestFactor(ctx, denom)
	if !foundBorrowInterestFactorPrior {
		newBorrowInterestFactorPrior := sdk.MustNewDecFromStr("1.0")
		k.SetBorrowInterestFactor(ctx, denom, newBorrowInterestFactorPrior)
		borrowInterestFactorPrior = newBorrowInterestFactorPrior
	}

	supplyInterestFactorPrior, foundSupplyInterestFactorPrior := k.GetSupplyInterestFactor(ctx, denom)
	if !foundSupplyInterestFactorPrior {
		newSupplyInterestFactorPrior := sdk.MustNewDecFromStr("1.0")
		k.SetSupplyInterestFactor(ctx, denom, newSupplyInterestFactorPrior)
		supplyInterestFactorPrior = newSupplyInterestFactorPrior
	}

	// Fetch money market from the store
	mm, found := k.GetMoneyMarket(ctx, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrMoneyMarketNotFound, "%s", denom)
	}

	// GetBorrowRate calculates the current interest rate based on utilization (the fraction of supply that has been borrowed)
	borrowRateApy, err := CalculateBorrowRate(mm.InterestRateModel, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowsPrior.Amount), sdk.NewDecFromInt(reservesPrior.Amount))
	if err != nil {
		return err
	}

	// Convert from APY to SPY, expressed as (1 + borrow rate)
	borrowRateSpy, err := APYToSPY(sdk.OneDec().Add(borrowRateApy))
	if err != nil {
		return err
	}

	// Calculate borrow interest factor and update
	borrowInterestFactor := CalculateBorrowInterestFactor(borrowRateSpy, sdk.NewInt(timeElapsed))
	interestAccumulated := (borrowInterestFactor.Mul(sdk.NewDecFromInt(borrowsPrior.Amount)).TruncateInt()).Sub(borrowsPrior.Amount)
	totalBorrowInterestAccumulated := sdk.NewCoins(sdk.NewCoin(denom, interestAccumulated))
	totalReservesNew := reservesPrior.Add(sdk.NewCoin(denom, sdk.NewDecFromInt(interestAccumulated).Mul(mm.ReserveFactor).TruncateInt()))
	borrowInterestFactorNew := borrowInterestFactorPrior.Mul(borrowInterestFactor)
	k.SetBorrowInterestFactor(ctx, denom, borrowInterestFactorNew)
	k.IncrementBorrowedCoins(ctx, totalBorrowInterestAccumulated)

	// Calculate supply interest factor and update
	borrowInterestFactorDiff := borrowInterestFactorNew.Sub(borrowInterestFactor)
	supplyInterestFactor := CalculateSupplyInterestFactor(borrowInterestFactorDiff, sdk.NewDecFromInt(cashPrior), sdk.NewDecFromInt(borrowsPrior.Amount), sdk.NewDecFromInt(reservesPrior.Amount), mm.ReserveFactor)
	supplyInterestFactorNew := supplyInterestFactorPrior.Mul(supplyInterestFactor)
	k.SetSupplyInterestFactor(ctx, denom, supplyInterestFactorNew)

	// Set accumulation keys in store
	k.SetTotalReserves(ctx, denom, totalReservesNew)
	k.SetPreviousAccrualTime(ctx, denom, ctx.BlockTime())

	return nil
}

// CalculateBorrowRate calculates the borrow rate, which is the current APY expressed as a decimal
// based on the current utilization.
func CalculateBorrowRate(model types.InterestRateModel, cash, borrows, reserves sdk.Dec) (sdk.Dec, error) {
	utilRatio := CalculateUtilizationRatio(cash, borrows, reserves)

	// Calculate normal borrow rate (under kink)
	if utilRatio.LTE(model.Kink) {
		return utilRatio.Mul(model.BaseMultiplier).Add(model.BaseRateAPY), nil
	}

	// Calculate jump borrow rate (over kink)
	normalRate := model.Kink.Mul(model.BaseMultiplier).Add(model.BaseRateAPY)
	excessUtil := utilRatio.Sub(model.Kink)
	return excessUtil.Mul(model.JumpMultiplier).Add(normalRate), nil
}

// CalculateUtilizationRatio calculates an asset's current utilization rate
func CalculateUtilizationRatio(cash, borrows, reserves sdk.Dec) sdk.Dec {
	// Utilization rate is 0 when there are no borrows
	if borrows.Equal(sdk.ZeroDec()) {
		return sdk.ZeroDec()
	}

	totalSupply := cash.Add(borrows).Sub(reserves)
	if totalSupply.IsNegative() {
		return sdk.OneDec()
	}

	return sdk.MinDec(sdk.OneDec(), borrows.Quo(totalSupply))
}

// CalculateBorrowInterestFactor calculates the simple interest scaling factor,
// which is equal to: (per-second interest rate * number of seconds elapsed)
// Will return 1.000x, multiply by principal to get new principal with added interest
func CalculateBorrowInterestFactor(perSecondInterestRate sdk.Dec, secondsElapsed sdk.Int) sdk.Dec {
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

// CalculateSupplyInterestFactor calculates the supply interest factor, which is the percentage of borrow interest
// that flows to each unit of supply, i.e. at 50% utilization and 0% reserve factor, a 5% borrow interest will
// correspond to a 2.5% supply interest.
func CalculateSupplyInterestFactor(borrowInterestFactorIncrement, cash, borrows, reserves, reserveFactor sdk.Dec) sdk.Dec {
	utilRatio := CalculateUtilizationRatio(cash, borrows, reserves)
	supplyInterestFactorIncrement := borrowInterestFactorIncrement.Mul(utilRatio).Mul((sdk.OneDec().Sub(reserveFactor)))
	return sdk.OneDec().Add(supplyInterestFactorIncrement)
}

// SyncBorrowInterest updates the user's owed interest on newly borrowed coins to the latest global state
func (k Keeper) SyncBorrowInterest(ctx sdk.Context, addr sdk.AccAddress) {
	totalNewInterest := sdk.Coins{}

	// Update user's borrow interest factor list for each asset in the 'coins' array.
	// We use a list of BorrowInterestFactors here because Amino doesn't support marshaling maps.
	borrow, found := k.GetBorrow(ctx, addr)
	if !found {
		return
	}
	for _, coin := range borrow.Amount {
		// Locate the borrow interest factor item by coin denom in the user's list of borrow indexes
		foundAtIndex := -1
		for i := range borrow.Index {
			if borrow.Index[i].Denom == coin.Denom {
				foundAtIndex = i
				break
			}
		}

		interestFactorValue, _ := k.GetBorrowInterestFactor(ctx, coin.Denom)
		if foundAtIndex == -1 { // First time user has borrowed this denom
			borrow.Index = append(borrow.Index, types.NewBorrowInterestFactor(coin.Denom, interestFactorValue))
		} else { // User has an existing borrow index for this denom
			// Calculate interest owed by user since asset's last borrow index update
			storedAmount := sdk.NewDecFromInt(borrow.Amount.AmountOf(coin.Denom))
			userLastInterestFactor := borrow.Index[foundAtIndex].Value
			interest := (storedAmount.Quo(userLastInterestFactor).Mul(interestFactorValue)).Sub(storedAmount)
			totalNewInterest = totalNewInterest.Add(sdk.NewCoin(coin.Denom, interest.TruncateInt()))
			// We're synced up, so update user's borrow index value to match the current global borrow index value
			borrow.Index[foundAtIndex].Value = interestFactorValue
		}
	}
	// Add all pending interest to user's borrow
	borrow.Amount = borrow.Amount.Add(totalNewInterest...)

	// Update user's borrow in the store
	k.SetBorrow(ctx, borrow)
}

// SyncSupplyInterest updates the user's earned interest on supplied coins based on the latest global state
func (k Keeper) SyncSupplyInterest(ctx sdk.Context, addr sdk.AccAddress) {
	totalNewInterest := sdk.Coins{}

	// Update user's supply index list for each asset in the 'coins' array.
	// We use a list of SupplyInterestFactors here because Amino doesn't support marshaling maps.
	deposit, found := k.GetDeposit(ctx, addr)
	if !found {
		return
	}

	for _, coin := range deposit.Amount {
		// Locate the deposit index item by coin denom in the user's list of deposit indexes
		foundAtIndex := -1
		for i := range deposit.Index {
			if deposit.Index[i].Denom == coin.Denom {
				foundAtIndex = i
				break
			}
		}

		interestFactorValue, _ := k.GetSupplyInterestFactor(ctx, coin.Denom)
		if foundAtIndex == -1 { // First time user has supplied this denom
			deposit.Index = append(deposit.Index, types.NewSupplyInterestFactor(coin.Denom, interestFactorValue))
		} else { // User has an existing supply index for this denom
			// Calculate interest earned by user since asset's last deposit index update
			storedAmount := sdk.NewDecFromInt(deposit.Amount.AmountOf(coin.Denom))
			userLastInterestFactor := deposit.Index[foundAtIndex].Value
			interest := (storedAmount.Quo(userLastInterestFactor).Mul(interestFactorValue)).Sub(storedAmount)
			totalNewInterest = totalNewInterest.Add(sdk.NewCoin(coin.Denom, interest.TruncateInt()))
			// We're synced up, so update user's deposit index value to match the current global deposit index value
			deposit.Index[foundAtIndex].Value = interestFactorValue
		}
	}
	// Add all pending interest to user's deposit
	deposit.Amount = deposit.Amount.Add(totalNewInterest...)

	// Update user's deposit in the store
	k.SetDeposit(ctx, deposit)
}

// APYToSPY converts the input annual interest rate. For example, 10% apy would be passed as 1.10.
// SPY = Per second compounded interest rate is how cosmos mathematically represents APY.
func APYToSPY(apy sdk.Dec) (sdk.Dec, error) {
	// Note: any APY 179 or greater will cause an out-of-bounds error
	root, err := apy.ApproxRoot(uint64(secondsPerYear))
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return root, nil
}

// minInt64 returns the smaller of x or y
func minDec(x, y sdk.Dec) sdk.Dec {
	if x.GT(y) {
		return y
	}
	return x
}
