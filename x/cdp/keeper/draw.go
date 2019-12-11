package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// AddPrincipal adds debt to a cdp if the additional debt does not put the cdp below the liquidation ratio
func (k Keeper) AddPrincipal(ctx sdk.Context, owner sdk.AccAddress, denom string, principal sdk.Coins) sdk.Error {
	cdp, found := k.GetCdpByOwnerAndDenom(ctx, owner, denom)
	if !found {
		return types.ErrCdpNotFound(k.codespace, owner, denom)
	}
	if !found {
		return types.ErrCdpNotFound(k.codespace, owner, denom)
	}
	err := k.ValidatePrincipal(ctx, principal)
	if err != nil {
		return err
	}
	cdp.AccumulatedFees = k.CalculateFees(ctx, cdp)
	cdp.FeesUpdated = ctx.BlockTime()
	collateralRatio, err := k.CalculateCollateralizationRatio(ctx, cdp.Collateral, cdp.Principal.Add(principal), cdp.AccumulatedFees)
	if err != nil {
		return err
	}
	liquidationRatio := k.getLiquidationRatio(ctx, denom)
	if collateralRatio.LT(liquidationRatio) {
		return types.ErrInvalidCollateralRatio(k.codespace, denom, collateralRatio, liquidationRatio)
	}
	err = k.supplyKeeper.MintCoins(ctx, types.ModuleName, principal)
	if err != nil {
		return err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, principal)
	if err != nil {
		return err
	}

	err = k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), principal)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpDraw,
			sdk.NewAttribute(sdk.AttributeKeyAmount, principal.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)

	cdp.Principal = cdp.Principal.Add(principal)
	k.IncrementTotalPrincipal(ctx, cdp.Collateral[0].Denom, principal)
	k.SetCDP(ctx, cdp)
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees))
	k.IndexCdpByCollateralRatio(ctx, cdp, collateralToDebtRatio)

	return nil
}

// RepayPrincipal removes debt from the cdp
// If all debt is repaid, the collateral is returned to depositors and the cdp is removed from the store
func (k Keeper) RepayPrincipal(ctx sdk.Context, owner sdk.AccAddress, denom string, payment sdk.Coins) sdk.Error {
	cdp, found := k.GetCdpByOwnerAndDenom(ctx, owner, denom)
	if !found {
		return types.ErrCdpNotFound(k.codespace, owner, denom)
	}
	err := k.ValidatePaymentCoins(ctx, cdp, payment)
	if err != nil {
		return err
	}
	cdp.AccumulatedFees = k.CalculateFees(ctx, cdp)
	cdp.FeesUpdated = ctx.BlockTime()
	feePayment, principalPayment := k.calculatePayment(ctx, cdp.AccumulatedFees, payment)
	if !principalPayment.IsZero() {
		cdp.Principal = cdp.Principal.Sub(principalPayment)
	}
	cdp.AccumulatedFees = cdp.AccumulatedFees.Sub(feePayment)
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, payment)
	if err != nil {
		return err
	}
	err = k.supplyKeeper.BurnCoins(ctx, types.ModuleName, payment)
	if err != nil {
		return err
	}
	err = k.BurnDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), payment)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpRepay,
			sdk.NewAttribute(sdk.AttributeKeyAmount, payment.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)

	if cdp.Collateral.IsZero() && cdp.AccumulatedFees.IsZero() {
		k.ReturnCollateral(ctx, cdp)
		k.DeleteCDP(ctx, cdp)
		k.RemoveCdpOwnerIndex(ctx, cdp)
		k.RemoveCdpLiquidationRatioIndex(ctx, cdp)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCdpClose,
				sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
			),
		)
		return nil
	}
	k.SetCDP(ctx, cdp)
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees))
	k.IndexCdpByCollateralRatio(ctx, cdp, collateralToDebtRatio)

	return nil
}

// ValidatePaymentCoins validates that the input coins are valid for repaying debt
func (k Keeper) ValidatePaymentCoins(ctx sdk.Context, cdp types.CDP, payment sdk.Coins) sdk.Error {
	subset := payment.DenomsSubsetOf(cdp.Principal)
	if !subset {
		var paymentDenoms []string
		var principalDenoms []string
		for _, pc := range cdp.Principal {
			principalDenoms = append(principalDenoms, pc.Denom)
		}
		for _, pc := range payment {
			paymentDenoms = append(paymentDenoms, pc.Denom)
		}
		return types.ErrInvalidPaymentDenom(k.codespace, cdp.ID, principalDenoms, paymentDenoms)
	}
	return nil
}

// ReturnCollateral returns collateral to depositors on a cdp and removes deposits from the store
func (k Keeper) ReturnCollateral(ctx sdk.Context, cdp types.CDP) {
	k.IterateDeposits(ctx, cdp.ID, func(deposit types.Deposit) bool {
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, deposit.Depositor, deposit.Amount)
		if err != nil {
			panic(err)
		}
		k.DeleteDeposit(ctx, cdp.ID, deposit.Depositor)
		return false
	})
}

func (k Keeper) calculatePayment(ctx sdk.Context, fees sdk.Coins, payment sdk.Coins) (sdk.Coins, sdk.Coins) {
	feePayment := sdk.NewCoins()
	principalPayment := sdk.NewCoins()
	for _, fc := range fees {
		if payment.AmountOf(fc.Denom).IsPositive() {
			if payment.AmountOf(fc.Denom).GT(fc.Amount) {
				feePayment = feePayment.Add(sdk.NewCoins(fc))
				pc := sdk.NewCoin(fc.Denom, payment.AmountOf(fc.Denom).Sub(fc.Amount))
				principalPayment = principalPayment.Add(sdk.NewCoins(pc))
			} else {
				fc := sdk.NewCoin(fc.Denom, payment.AmountOf(fc.Denom))
				feePayment = feePayment.Add(sdk.NewCoins(fc))
			}
		}
	}
	return feePayment, principalPayment
}
