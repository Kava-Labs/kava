package keeper

import (
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

var scalingFactor = 1e18

// AccumulateInterest calculates the new interest that has accrued for the input collateral type based on the total amount of principal
// that has been created with that collateral type and the amount of time that has passed since interest was last accumulated
func (k Keeper) AccumulateInterest(ctx sdk.Context, ctype string) error {
	previousAccrualTime, found := k.GetPreviousAccrualTime(ctx, ctype)
	if !found {
		k.SetPreviousAccrualTime(ctx, ctype, ctx.BlockTime())
		return nil
	}

	timeElapsed := int64(math.RoundToEven(
		ctx.BlockTime().Sub(previousAccrualTime).Seconds(),
	))
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
	interestMantissa := sdk.NewUintFromBigInt(perSecondInterestRate.MulInt(scalingFactorInt).RoundInt().BigInt())

	// Convert seconds elapsed to uint (*not scaled*)
	secondsElapsedUint := sdk.NewUintFromBigInt(secondsElapsed.BigInt())

	// Calculate the interest factor as a uint scaled by 1e18
	interestFactorMantissa := sdk.RelativePow(interestMantissa, secondsElapsedUint, scalingFactorUint)

	// Convert interest factor to an unscaled sdk.Dec
	return sdk.NewDecFromBigInt(interestFactorMantissa.BigInt()).QuoInt(scalingFactorInt)
}

// SynchronizeInterest updates the input cdp object to reflect the current accumulated interest, updates the cdp state in the store,
// and returns the updated cdp object
func (k Keeper) SynchronizeInterest(ctx sdk.Context, cdp types.CDP) types.CDP {
	globalInterestFactor, found := k.GetInterestFactor(ctx, cdp.Type)
	if !found {
		k.SetInterestFactor(ctx, cdp.Type, sdk.OneDec())
		cdp.InterestFactor = sdk.OneDec()
		cdp.FeesUpdated = ctx.BlockTime()
		if err := k.SetCDP(ctx, cdp); err != nil {
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
		if err := k.SetCDP(ctx, cdp); err != nil {
			panic(err)
		}
	}

	cdp.AccumulatedFees = cdp.AccumulatedFees.Add(accumulatedInterest)
	cdp.FeesUpdated = prevAccrualTime
	cdp.InterestFactor = globalInterestFactor
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())
	if err := k.UpdateCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio); err != nil {
		panic(err)
	}

	return cdp
}

// CalculateNewInterest returns the amount of interest that has accrued to the cdp since its interest was last synchronized
func (k Keeper) CalculateNewInterest(ctx sdk.Context, cdp types.CDP) sdk.Coin {
	globalInterestFactor, found := k.GetInterestFactor(ctx, cdp.Type)
	if !found {
		return sdk.NewCoin(cdp.AccumulatedFees.Denom, sdk.ZeroInt())
	}
	cdpInterestFactor := globalInterestFactor.Quo(cdp.InterestFactor)
	if cdpInterestFactor.Equal(sdk.OneDec()) {
		return sdk.NewCoin(cdp.AccumulatedFees.Denom, sdk.ZeroInt())
	}
	accumulatedInterest := cdp.GetTotalPrincipal().Amount.ToDec().Mul(cdpInterestFactor).RoundInt().Sub(cdp.GetTotalPrincipal().Amount)
	return sdk.NewCoin(cdp.AccumulatedFees.Denom, accumulatedInterest)
}

// SynchronizeInterestForRiskyCDPs synchronizes the interest for the slice of cdps with the lowest collateral:debt ratio
func (k Keeper) SynchronizeInterestForRiskyCDPs(ctx sdk.Context, targetRatio sdk.Dec, collateralParams types.CollateralParams) error {
	debtParam := k.GetParams(ctx).DebtParam

	cdpStore := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	collateralRatioStore := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)

	for _, cp := range collateralParams {
		cdpIDs := make([]uint64, 0, cp.CheckCollateralizationIndexCount.Int64())
		iterator := k.CdpCollateralRatioIndexIterator(ctx, cp.Type, targetRatio)

		for ; iterator.Valid(); iterator.Next() {
			_, id, _ := types.SplitCollateralRatioKey(iterator.Key())
			cdpIDs = append(cdpIDs, id)
			if int64(len(cdpIDs)) >= cp.CheckCollateralizationIndexCount.Int64() {
				break
			}
		}
		iterator.Close()

		globalInterestFactor, found := k.GetInterestFactor(ctx, cp.Type)
		if !found {
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
				panic(fmt.Sprintf("unkown debt param %s", cdp.GetTotalPrincipal().Denom))
			}

			//
			// HOOK
			//
			k.hooks.BeforeCDPModified(ctx, cdp)

			//
			// CALC INTEREST
			//
			accumulatedInterest := sdk.ZeroInt()
			cdpInterestFactor := globalInterestFactor.Quo(cdp.InterestFactor)
			if !cdpInterestFactor.Equal(sdk.OneDec()) {
				accumulatedInterest = cdp.GetTotalPrincipal().Amount.ToDec().Mul(cdpInterestFactor).RoundInt().Sub(cdp.GetTotalPrincipal().Amount)
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

	}
	return nil
}

func calculateCollateralRatio(debtParam types.DebtParam, collateralParam types.CollateralParam, cdp types.CDP) sdk.Dec {
	debtTotal := sdk.NewDecFromInt(cdp.GetTotalPrincipal().Amount).Mul(sdk.NewDecFromIntWithPrec(sdk.OneInt(), debtParam.ConversionFactor.Int64()))

	if debtTotal.IsZero() || debtTotal.GTE(types.MaxSortableDec) {
		return types.MaxSortableDec.Sub(sdk.SmallestDec())
	} else {
		collateralBaseUnits := sdk.NewDecFromInt(cdp.Collateral.Amount).Mul(sdk.NewDecFromIntWithPrec(sdk.OneInt(), collateralParam.ConversionFactor.Int64()))
		return collateralBaseUnits.Quo(debtTotal)
	}
}
