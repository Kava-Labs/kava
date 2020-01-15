package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// SeizeCollateral liquidates the collateral in the input cdp.
// the following operations are performed:
// 1. updates the fees for the input cdp,
// 2. sends collateral for all deposits from the cdp module to the liquidator module,
// 3. moves debt coins from the cdp module to the liquidator module,
// 4. decrements the total amount of principal outstanding for that collateral type
// (this is the equivalent of saying that fees are no longer accumulated by a cdp once it
// gets liquidated)
func (k Keeper) SeizeCollateral(ctx sdk.Context, cdp types.CDP) sdk.Error {
	// Calculate the previous collateral ratio
	oldCollateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees))
	// Update fees
	periods := sdk.NewInt(ctx.BlockTime().Unix()).Sub(sdk.NewInt(cdp.FeesUpdated.Unix()))
	fees := k.CalculateFees(ctx, cdp.Principal.Add(cdp.AccumulatedFees), periods, cdp.Collateral[0].Denom)
	cdp.AccumulatedFees = cdp.AccumulatedFees.Add(fees)
	cdp.FeesUpdated = ctx.BlockTime()

	// TODO implement liquidation penalty

	// Move debt coins from cdp to liquidator account
	deposits := k.GetDeposits(ctx, cdp.ID)
	debt := sdk.ZeroInt()
	for _, pc := range cdp.Principal {
		debt = debt.Add(pc.Amount)
	}
	for _, dc := range cdp.AccumulatedFees {
		debt = debt.Add(dc.Amount)
	}
	debtCoin := sdk.NewCoin(k.GetDebtDenom(ctx), debt)
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorMacc, sdk.NewCoins(debtCoin))
	if err != nil {
		return err
	}

	// liquidate deposits and send collateral from cdp to liquidator
	for _, dep := range deposits {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCdpLiquidation,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
				sdk.NewAttribute(types.AttributeKeyDepositor, fmt.Sprintf("%s", dep.Depositor)),
			),
		)
		err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorMacc, dep.Amount)
		if err != nil {
			return err
		}
		k.DeleteDeposit(ctx, dep.CdpID, dep.Depositor)
	}
	err = k.AuctionCollateral(ctx, deposits, debt, cdp.Principal[0].Denom)
	if err != nil {
		return err
	}

	// Decrement total principal for this collateral type
	for _, dc := range cdp.Principal {
		feeAmount := cdp.AccumulatedFees.AmountOf(dc.Denom)
		coinsToDecrement := sdk.NewCoins(dc)
		if feeAmount.IsPositive() {
			feeCoins := sdk.NewCoins(sdk.NewCoin(dc.Denom, feeAmount))
			coinsToDecrement = coinsToDecrement.Add(feeCoins)
		}
		k.DecrementTotalPrincipal(ctx, cdp.Collateral[0].Denom, coinsToDecrement)
	}
	k.RemoveCdpOwnerIndex(ctx, cdp)
	k.RemoveCdpCollateralRatioIndex(ctx, cdp.Collateral[0].Denom, cdp.ID, oldCollateralToDebtRatio)
	k.DeleteCDP(ctx, cdp)
	return nil
}

// HandleNewDebt compounds the accumulated fees for the input collateral and principal coins.
// the following operations are performed:
// 1. mints the fee coins in the liquidator module account,
// 2. mints the same amount of debt coins in the cdp module account
// 3. updates the total amount of principal for the input collateral type in the store,
func (k Keeper) HandleNewDebt(ctx sdk.Context, collateralDenom string, principalDenom string, periods sdk.Int) {
	previousDebt := k.GetTotalPrincipal(ctx, collateralDenom, principalDenom)
	feeCoins := sdk.NewCoins(sdk.NewCoin(principalDenom, previousDebt))
	newFees := k.CalculateFees(ctx, feeCoins, periods, collateralDenom)
	k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), newFees)
	k.supplyKeeper.MintCoins(ctx, types.LiquidatorMacc, newFees)
	k.SetTotalPrincipal(ctx, collateralDenom, principalDenom, feeCoins.Add(newFees).AmountOf(principalDenom))
}

// LiquidateCdps seizes collateral from all CDPs below the input liquidation ratio
func (k Keeper) LiquidateCdps(ctx sdk.Context, marketID string, denom string, liquidationRatio sdk.Dec) sdk.Error {
	price, err := k.pricefeedKeeper.GetCurrentPrice(ctx, marketID)
	if err != nil {
		return err
	}
	normalizedRatio := sdk.OneDec().Quo(price.Price.Quo(liquidationRatio))
	cdpsToLiquidate := k.GetAllCdpsByDenomAndRatio(ctx, denom, normalizedRatio)
	for _, c := range cdpsToLiquidate {
		err := k.SeizeCollateral(ctx, c)
		if err != nil {
			return err
		}
	}
	return nil
}
