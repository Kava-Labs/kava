package keeper

import (
	"context"
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

var scalingFactor = 1e18

// AccumulateInterest calculates the new interest that has accrued for the input collateral type based on the total amount of principal
// that has been created with that collateral type and the amount of time that has passed since interest was last accumulated
func (k Keeper) AccumulateInterest(ctx context.Context, ctype string) error {
	fmt.Println("AccumulateInterest", ctype)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 6))
	//		cdp.BeginBlocker(suite.ctx, suite.keeper)
	//		// block time with balance 2024-10-24 04:02:57.651655 +0000 UTC 1000000027
	//		fmt.Println("balance time with balance", suite.ctx.BlockTime(), bk.GetBalance(suite.ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	cdpMacc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	fmt.Println("AccumulateInterest 1", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	previousAccrualTime, found := k.GetPreviousAccrualTime(ctx, ctype)
	fmt.Println("previousAccrualTime", previousAccrualTime, found)
	if !found {
		k.SetPreviousAccrualTime(ctx, ctype, sdkCtx.BlockTime())
		return nil
	}

	fmt.Println("AccumulateInterest 2", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	timeElapsed := int64(math.RoundToEven(
		sdkCtx.BlockTime().Sub(previousAccrualTime).Seconds(),
	))
	if timeElapsed == 0 {
		return nil
	}

	fmt.Println("timeElapsed", timeElapsed)
	fmt.Println("AccumulateInterest 3", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	totalPrincipalPrior := k.GetTotalPrincipal(ctx, ctype, types.DefaultStableDenom)
	if totalPrincipalPrior.IsZero() || totalPrincipalPrior.IsNegative() {
		k.SetPreviousAccrualTime(ctx, ctype, sdkCtx.BlockTime())
		return nil
	}

	fmt.Println("totalPrincipalPrior", totalPrincipalPrior)
	fmt.Println("AccumulateInterest 4", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	interestFactorPrior, foundInterestFactorPrior := k.GetInterestFactor(ctx, ctype)
	if !foundInterestFactorPrior {
		k.SetInterestFactor(ctx, ctype, sdkmath.LegacyOneDec())
		// set previous accrual time exit early because interest accumulated will be zero
		k.SetPreviousAccrualTime(ctx, ctype, sdkCtx.BlockTime())
		return nil
	}

	fmt.Println("interestFactorPrior", interestFactorPrior)

	borrowRateSpy := k.getFeeRate(ctx, ctype)
	if borrowRateSpy.Equal(sdkmath.LegacyOneDec()) {
		k.SetPreviousAccrualTime(ctx, ctype, sdkCtx.BlockTime())
		return nil
	}
	interestFactor := CalculateInterestFactor(borrowRateSpy, sdkmath.NewInt(timeElapsed))
	interestAccumulated := (interestFactor.Mul(sdkmath.LegacyNewDecFromInt(totalPrincipalPrior))).RoundInt().Sub(totalPrincipalPrior)
	if interestAccumulated.IsZero() {
		// in the case accumulated interest rounds to zero, exit early without updating accrual time
		return nil
	}
	err := k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(sdkCtx), sdk.NewCoin(types.DefaultStableDenom, interestAccumulated))
	if err != nil {
		return err
	}

	fmt.Println("interestFactor", interestFactor)
	fmt.Println("interestAccumulated", interestAccumulated)

	dp, found := k.GetDebtParam(ctx, types.DefaultStableDenom)
	if !found {
		panic(fmt.Sprintf("Debt parameters for %s not found", types.DefaultStableDenom))
	}

	newFeesSurplus := interestAccumulated

	// mint surplus coins to the liquidator module account.
	if newFeesSurplus.IsPositive() {
		err := k.bankKeeper.MintCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, newFeesSurplus)))
		if err != nil {
			return err
		}
	}

	interestFactorNew := interestFactorPrior.Mul(interestFactor)
	totalPrincipalNew := totalPrincipalPrior.Add(interestAccumulated)

	fmt.Println("interestFactorNew", interestFactorNew)
	fmt.Println("totalPrincipalNew", totalPrincipalNew)

	k.SetTotalPrincipal(ctx, ctype, types.DefaultStableDenom, totalPrincipalNew)
	k.SetInterestFactor(ctx, ctype, interestFactorNew)
	k.SetPreviousAccrualTime(ctx, ctype, sdkCtx.BlockTime())

	fmt.Println("AccumulateInterest last", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	return nil
}

// CalculateInterestFactor calculates the simple interest scaling factor,
// which is equal to: (per-second interest rate ** number of seconds elapsed)
// Will return 1.000x, multiply by principal to get new principal with added interest
func CalculateInterestFactor(perSecondInterestRate sdkmath.LegacyDec, secondsElapsed sdkmath.Int) sdkmath.LegacyDec {
	fmt.Println("CalculateInterestFactor", perSecondInterestRate, secondsElapsed)
	scalingFactorUint := sdkmath.NewUint(uint64(scalingFactor))
	scalingFactorInt := sdkmath.NewInt(int64(scalingFactor))

	// Convert per-second interest rate to a uint scaled by 1e18
	interestMantissa := sdkmath.NewUintFromBigInt(perSecondInterestRate.MulInt(scalingFactorInt).RoundInt().BigInt())

	// Convert seconds elapsed to uint (*not scaled*)
	secondsElapsedUint := sdkmath.NewUintFromBigInt(secondsElapsed.BigInt())

	// Calculate the interest factor as a uint scaled by 1e18
	interestFactorMantissa := sdkmath.RelativePow(interestMantissa, secondsElapsedUint, scalingFactorUint)

	// Convert interest factor to an unscaled sdkmath.LegacyDec
	return sdkmath.LegacyNewDecFromBigInt(interestFactorMantissa.BigInt()).QuoInt(scalingFactorInt)
}

// SynchronizeInterest updates the input cdp object to reflect the current accumulated interest, updates the cdp state in the store,
// and returns the updated cdp object
func (k Keeper) SynchronizeInterest(ctx context.Context, cdp types.CDP) types.CDP {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	globalInterestFactor, found := k.GetInterestFactor(ctx, cdp.Type)
	if !found {
		k.SetInterestFactor(ctx, cdp.Type, sdkmath.LegacyOneDec())
		cdp.InterestFactor = sdkmath.LegacyOneDec()
		cdp.FeesUpdated = sdkCtx.BlockTime()
		if err := k.SetCDP(sdkCtx, cdp); err != nil {
			panic(err)
		}
		return cdp
	}

	accumulatedInterest := k.CalculateNewInterest(ctx, cdp)
	prevAccrualTime, found := k.GetPreviousAccrualTime(ctx, cdp.Type)
	if !found {
		return cdp
	}
	if accumulatedInterest.IsZero() {
		// accumulated interest is zero if apy is zero or are if the total fees for all cdps round to zero
		if cdp.FeesUpdated.Equal(prevAccrualTime) {
			// if all fees are rounding to zero, don't update FeesUpdated
			return cdp
		}
		// if apy is zero, we need to update FeesUpdated
		cdp.FeesUpdated = prevAccrualTime
		if err := k.SetCDP(sdkCtx, cdp); err != nil {
			panic(err)
		}
	}

	cdp.AccumulatedFees = cdp.AccumulatedFees.Add(accumulatedInterest)
	cdp.FeesUpdated = prevAccrualTime
	cdp.InterestFactor = globalInterestFactor
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(sdkCtx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())
	if err := k.UpdateCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio); err != nil {
		panic(err)
	}

	return cdp
}

// CalculateNewInterest returns the amount of interest that has accrued to the cdp since its interest was last synchronized
func (k Keeper) CalculateNewInterest(ctx context.Context, cdp types.CDP) sdk.Coin {
	globalInterestFactor, found := k.GetInterestFactor(ctx, cdp.Type)
	if !found {
		return sdk.NewCoin(cdp.AccumulatedFees.Denom, sdkmath.ZeroInt())
	}
	cdpInterestFactor := globalInterestFactor.Quo(cdp.InterestFactor)
	if cdpInterestFactor.Equal(sdkmath.LegacyOneDec()) {
		return sdk.NewCoin(cdp.AccumulatedFees.Denom, sdkmath.ZeroInt())
	}
	accumulatedInterest := sdkmath.LegacyNewDecFromInt(cdp.GetTotalPrincipal().Amount).Mul(cdpInterestFactor).RoundInt().Sub(cdp.GetTotalPrincipal().Amount)
	return sdk.NewCoin(cdp.AccumulatedFees.Denom, accumulatedInterest)
}

// SynchronizeInterestForRiskyCDPs synchronizes the interest for the slice of cdps with the lowest collateral:debt ratio
func (k Keeper) SynchronizeInterestForRiskyCDPs(ctx context.Context, targetRatio sdkmath.LegacyDec, cp types.CollateralParam) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	debtParam := k.GetParams(ctx).DebtParam

	cdpMacc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	fmt.Println("SynchronizeInterestForRiskyCDPs 1", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	cdpStore := prefix.NewStore(sdkCtx.KVStore(k.key), types.CdpKeyPrefix)
	collateralRatioStore := prefix.NewStore(sdkCtx.KVStore(k.key), types.CollateralRatioIndexPrefix)

	cdpIDs := make([]uint64, 0, cp.CheckCollateralizationIndexCount.Int64())

	iterator := collateralRatioStore.Iterator(types.CollateralRatioIterKey(cp.Type, sdkmath.LegacyZeroDec()), types.CollateralRatioIterKey(cp.Type, targetRatio))
	for ; iterator.Valid(); iterator.Next() {
		_, id, _ := types.SplitCollateralRatioKey(iterator.Key())
		cdpIDs = append(cdpIDs, id)
		if int64(len(cdpIDs)) >= cp.CheckCollateralizationIndexCount.Int64() {
			break
		}
	}
	iterator.Close()

	globalInterestFactor, found := k.GetInterestFactor(ctx, cp.Type)
	if !found && len(cdpIDs) > 0 {
		panic(fmt.Sprintf("global interest factor not found for type %s", cp.Type))
	}
	prevAccrualTime, found := k.GetPreviousAccrualTime(ctx, cp.Type)
	if !found {
		panic(fmt.Sprintf("previous accrual time not found for type %s", cp.Type))
	}

	for _, cdpID := range cdpIDs {
		//
		// GET CDP
		//
		bz := cdpStore.Get(types.CdpKey(cp.Type, cdpID))
		if bz == nil {
			panic(fmt.Sprintf("cdp %d does not exist", cdpID))
		}
		var cdp types.CDP
		k.cdc.MustUnmarshal(bz, &cdp)

		if debtParam.Denom != cdp.GetTotalPrincipal().Denom {
			panic(fmt.Sprintf("unknown debt param %s", cdp.GetTotalPrincipal().Denom))
		}

		//
		// HOOK
		//
		k.hooks.BeforeCDPModified(sdkCtx, cdp)

		//
		// CALC INTEREST
		//
		accumulatedInterest := sdkmath.ZeroInt()
		cdpInterestFactor := globalInterestFactor.Quo(cdp.InterestFactor)
		if !cdpInterestFactor.Equal(sdkmath.LegacyOneDec()) {
			accumulatedInterest = sdkmath.LegacyNewDecFromInt(cdp.GetTotalPrincipal().Amount).Mul(cdpInterestFactor).RoundInt().Sub(cdp.GetTotalPrincipal().Amount)
		}

		if accumulatedInterest.IsZero() {
			// accumulated interest is zero if apy is zero or are if the total fees for all cdps round to zero
			if cdp.FeesUpdated.Equal(prevAccrualTime) {
				// if all fees are rounding to zero, don't update FeesUpdated
				continue
			}
			// if apy is zero, we need to update FeesUpdated
			cdp.FeesUpdated = prevAccrualTime
			bz = k.cdc.MustMarshal(&cdp)
			cdpStore.Set(types.CdpKey(cdp.Type, cdp.ID), bz)
		}

		//
		// GET OLD RATIO
		//
		previousCollateralRatio := calculateCollateralRatio(debtParam, cp, cdp)

		//
		// UPDATE CDP
		//
		cdp.AccumulatedFees = cdp.AccumulatedFees.Add(sdk.NewCoin(cdp.AccumulatedFees.Denom, accumulatedInterest))
		cdp.FeesUpdated = prevAccrualTime
		cdp.InterestFactor = globalInterestFactor

		//
		// CALC NEW RATIO
		//
		updatedCollateralRatio := calculateCollateralRatio(debtParam, cp, cdp)

		//
		// UPDATE STORE
		//
		collateralRatioStore.Delete(types.CollateralRatioKey(cdp.Type, cdp.ID, previousCollateralRatio))
		bz = k.cdc.MustMarshal(&cdp)
		cdpStore.Set(types.CdpKey(cdp.Type, cdp.ID), bz)
		collateralRatioStore.Set(types.CollateralRatioKey(cdp.Type, cdp.ID, updatedCollateralRatio), types.GetCdpIDBytes(cdp.ID))
	}

	fmt.Println("SynchronizeInterestForRiskyCDPs end", sdkCtx.BlockTime(), k.bankKeeper.GetBalance(ctx, cdpMacc.GetAddress(), "debt").Amount.Int64())

	return nil
}

func calculateCollateralRatio(debtParam types.DebtParam, collateralParam types.CollateralParam, cdp types.CDP) sdkmath.LegacyDec {
	debtTotal := sdkmath.LegacyNewDecFromInt(cdp.GetTotalPrincipal().Amount).Mul(sdkmath.LegacyNewDecFromIntWithPrec(sdkmath.OneInt(), debtParam.ConversionFactor.Int64()))

	if debtTotal.IsZero() || debtTotal.GTE(types.MaxSortableDec) {
		return types.MaxSortableDec.Sub(sdkmath.LegacySmallestDec())
	} else {
		collateralBaseUnits := sdkmath.LegacyNewDecFromInt(cdp.Collateral.Amount).Mul(sdkmath.LegacyNewDecFromIntWithPrec(sdkmath.OneInt(), collateralParam.ConversionFactor.Int64()))
		return collateralBaseUnits.Quo(debtTotal)
	}
}
