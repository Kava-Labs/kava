package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/cdp/types"
)

// AddPrincipal adds debt to a cdp if the additional debt does not put the cdp below the liquidation ratio
func (k Keeper) AddPrincipal(ctx sdk.Context, owner sdk.AccAddress, denom string, principal sdk.Coins) error {
	// validation
	cdp, found := k.GetCdpByOwnerAndDenom(ctx, owner, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrCdpNotFound, "owner (%s), denom (%s)", owner, denom)
	}
	err := k.ValidatePrincipalDraw(ctx, principal)
	if err != nil {
		return err
	}

	err = k.ValidateDebtLimit(ctx, cdp.Collateral[0].Denom, principal)
	if err != nil {
		return err
	}

	// fee calculation
	periods := sdk.NewInt(ctx.BlockTime().Unix()).Sub(sdk.NewInt(cdp.FeesUpdated.Unix()))
	fees := k.CalculateFees(ctx, cdp.Principal.Add(cdp.AccumulatedFees...), periods, cdp.Collateral[0].Denom)

	err = k.ValidateCollateralizationRatio(ctx, cdp.Collateral, cdp.Principal.Add(principal...), cdp.AccumulatedFees.Add(fees...))
	if err != nil {
		return err
	}

	// mint the principal and send it to the cdp owner
	err = k.supplyKeeper.MintCoins(ctx, types.ModuleName, principal)
	if err != nil {
		panic(err)
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, principal)
	if err != nil {
		panic(err)
	}

	// mint the corresponding amount of debt coins in the cdp module account
	err = k.MintDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), principal)
	if err != nil {
		panic(err)
	}

	// emit cdp draw event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpDraw,
			sdk.NewAttribute(sdk.AttributeKeyAmount, principal.String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)

	// remove old collateral:debt index
	oldCollateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees...))
	k.RemoveCdpCollateralRatioIndex(ctx, denom, cdp.ID, oldCollateralToDebtRatio)

	// update cdp state
	cdp.Principal = cdp.Principal.Add(principal...)
	cdp.AccumulatedFees = cdp.AccumulatedFees.Add(fees...)
	cdp.FeesUpdated = ctx.BlockTime()

	// increment total principal for the input collateral type
	k.IncrementTotalPrincipal(ctx, cdp.Collateral[0].Denom, principal)

	// set cdp state and indexes in the store
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees...))
	k.SetCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio)

	return nil
}

// RepayPrincipal removes debt from the cdp
// If all debt is repaid, the collateral is returned to depositors and the cdp is removed from the store
func (k Keeper) RepayPrincipal(ctx sdk.Context, owner sdk.AccAddress, denom string, payment sdk.Coins) error {
	// validation
	cdp, found := k.GetCdpByOwnerAndDenom(ctx, owner, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrCdpNotFound, "owner (%s), denom (%s)", owner, denom)
	}

	// calculate fees
	periods := sdk.NewInt(ctx.BlockTime().Unix()).Sub(sdk.NewInt(cdp.FeesUpdated.Unix()))
	fees := k.CalculateFees(ctx, cdp.Principal.Add(cdp.AccumulatedFees...), periods, cdp.Collateral[0].Denom)
	err := k.ValidatePaymentCoins(ctx, cdp, payment, cdp.Principal.Add(cdp.AccumulatedFees...).Add(fees...))
	if err != nil {
		return err
	}

	// calculate fee and principal payment
	feePayment, principalPayment := k.calculatePayment(ctx, cdp.Principal.Add(cdp.AccumulatedFees...).Add(fees...), cdp.AccumulatedFees.Add(fees...), payment)

	// send the payment from the sender to the cpd module
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, feePayment.Add(principalPayment...))
	if err != nil {
		return err
	}

	// burn the payment coins
	err = k.supplyKeeper.BurnCoins(ctx, types.ModuleName, feePayment.Add(principalPayment...))
	if err != nil {
		panic(err)
	}

	// burn the corresponding amount of debt coins
	cdpDebt := k.getModAccountDebt(ctx, types.ModuleName)
	paymentAmount := sdk.ZeroInt()
	for _, c := range feePayment.Add(principalPayment...) {
		paymentAmount = paymentAmount.Add(c.Amount)
	}
	coinsToBurn := sdk.NewCoins(sdk.NewCoin(k.GetDebtDenom(ctx), paymentAmount))
	if paymentAmount.GT(cdpDebt) {
		coinsToBurn = sdk.NewCoins(sdk.NewCoin(k.GetDebtDenom(ctx), cdpDebt))
	}
	err = k.BurnDebtCoins(ctx, types.ModuleName, k.GetDebtDenom(ctx), coinsToBurn)
	if err != nil {
		panic(err)
	}

	// emit repayment event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCdpRepay,
			sdk.NewAttribute(sdk.AttributeKeyAmount, feePayment.Add(principalPayment...).String()),
			sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
		),
	)

	// remove the old collateral:debt ratio index
	oldCollateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees...))
	k.RemoveCdpCollateralRatioIndex(ctx, denom, cdp.ID, oldCollateralToDebtRatio)

	// update cdp state
	if !principalPayment.IsZero() {
		cdp.Principal = cdp.Principal.Sub(principalPayment)
	}
	cdp.AccumulatedFees = cdp.AccumulatedFees.Add(fees...).Sub(feePayment)
	cdp.FeesUpdated = ctx.BlockTime()

	// decrement the total principal for the input collateral type
	k.DecrementTotalPrincipal(ctx, denom, feePayment.Add(principalPayment...))

	// if the debt is fully paid, return collateral to depositors,
	// and remove the cdp and indexes from the store
	if cdp.Principal.IsZero() && cdp.AccumulatedFees.IsZero() {
		k.ReturnCollateral(ctx, cdp)
		k.DeleteCDP(ctx, cdp)
		k.RemoveCdpOwnerIndex(ctx, cdp)

		// emit cdp close event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCdpClose,
				sdk.NewAttribute(types.AttributeKeyCdpID, fmt.Sprintf("%d", cdp.ID)),
			),
		)
		return nil
	}

	// set cdp state and update indexes
	collateralToDebtRatio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees...))
	k.SetCdpAndCollateralRatioIndex(ctx, cdp, collateralToDebtRatio)
	return nil
}

// ValidatePaymentCoins validates that the input coins are valid for repaying debt
func (k Keeper) ValidatePaymentCoins(ctx sdk.Context, cdp types.CDP, payment sdk.Coins, debt sdk.Coins) error {
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
		return sdkerrors.Wrapf(types.ErrInvalidPayment, "cdp %d: expected %s, got %s", cdp.ID, principalDenoms, paymentDenoms)
	}
	for _, dc := range payment {
		dp, _ := k.GetDebtParam(ctx, dc.Denom)
		proposedBalance := cdp.Principal.AmountOf(dc.Denom).Sub(dc.Amount)
		if proposedBalance.GT(sdk.ZeroInt()) && proposedBalance.LT(dp.DebtFloor) {
			return sdkerrors.Wrapf(types.ErrBelowDebtFloor, "proposed %s < minimum %s", sdk.NewCoins(sdk.NewCoin(dc.Denom, proposedBalance)), dp.DebtFloor)
		}
	}
	return nil
}

// ReturnCollateral returns collateral to depositors on a cdp and removes deposits from the store
func (k Keeper) ReturnCollateral(ctx sdk.Context, cdp types.CDP) {
	deposits := k.GetDeposits(ctx, cdp.ID)
	for _, deposit := range deposits {
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, deposit.Depositor, deposit.Amount)
		if err != nil {
			panic(err)
		}
		k.DeleteDeposit(ctx, cdp.ID, deposit.Depositor)
	}
}

func (k Keeper) calculatePayment(ctx sdk.Context, owed sdk.Coins, fees sdk.Coins, payment sdk.Coins) (sdk.Coins, sdk.Coins) {
	// divides repayment into principal and fee components, with fee payment applied first.

	feePayment := sdk.NewCoins()
	principalPayment := sdk.NewCoins()
	overpayment := sdk.NewCoins()
	// TODO must compare denoms directly if there are multiple principal denoms
	if payment.IsAllGT(owed) {
		overpayment = payment.Sub(owed)
		payment = payment.Sub(overpayment)
	}
	if fees.IsZero() {
		return sdk.NewCoins(), payment
	}
	for _, fc := range fees {
		if payment.AmountOf(fc.Denom).IsPositive() {
			if payment.AmountOf(fc.Denom).GT(fc.Amount) {
				feePayment = feePayment.Add(fc)
				pc := sdk.NewCoin(fc.Denom, payment.AmountOf(fc.Denom).Sub(fc.Amount))
				principalPayment = principalPayment.Add(pc)
			} else {
				fc := sdk.NewCoin(fc.Denom, payment.AmountOf(fc.Denom))
				feePayment = feePayment.Add(fc)
			}
		}
	}
	return feePayment, principalPayment
}
