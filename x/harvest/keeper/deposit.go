package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/kava-labs/kava/x/harvest/types"
)

// Deposit deposit
func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coin, depositType types.DepositType) error {

	err := k.ValidateDeposit(ctx, amount, depositType)
	if err != nil {
		return err
	}

	switch depositType {
	case types.LP:
		err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, sdk.NewCoins(amount))
	default:
		return sdkerrors.Wrap(types.ErrInvalidDepositType, string(depositType))
	}
	if err != nil {
		return err
	}

	deposit, found := k.GetDeposit(ctx, depositor, amount.Denom, depositType)
	if !found {
		deposit = types.NewDeposit(depositor, amount, depositType)
	} else {
		deposit.Amount = deposit.Amount.Add(amount)
	}

	k.SetDeposit(ctx, deposit)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, deposit.Depositor.String()),
			sdk.NewAttribute(types.AttributeKeyDepositDenom, deposit.Amount.Denom),
			sdk.NewAttribute(types.AttributeKeyDepositType, string(depositType)),
		),
	)

	return nil
}

// ValidateDeposit validates a deposit
func (k Keeper) ValidateDeposit(ctx sdk.Context, amount sdk.Coin, depositType types.DepositType) error {
	var err error
	switch depositType {
	case types.LP:
		err = k.ValidateLPDeposit(ctx, amount, depositType)
	default:
		return sdkerrors.Wrap(types.ErrInvalidDepositType, string(depositType))
	}
	if err != nil {
		return err
	}
	return nil
}

// ValidateLPDeposit validates that a liquidity provider deposit
func (k Keeper) ValidateLPDeposit(ctx sdk.Context, amount sdk.Coin, depositType types.DepositType) error {
	params := k.GetParams(ctx)
	for _, lps := range params.LiquidityProviderSchedules {
		if lps.DepositDenom == amount.Denom {
			return nil
		}
	}
	return sdkerrors.Wrapf(types.ErrInvalidDepositDenom, "liquidity provider denom %s not found", amount.Denom)
}

// Withdraw returns some or all of a deposit back to original depositor
func (k Keeper) Withdraw(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coin, depositType types.DepositType) error {
	deposit, found := k.GetDeposit(ctx, depositor, amount.Denom, depositType)
	if !found {
		return sdkerrors.Wrapf(types.ErrDepositNotFound, "no %s %s deposit found for %s", amount.Denom, depositType, depositor)
	}
	if !deposit.Amount.IsGTE(amount) {
		return sdkerrors.Wrapf(types.ErrInvaliWithdrawAmount, "%s>%s", amount, deposit.Amount)
	}

	var err error
	switch depositType {
	case types.LP:
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, depositor, sdk.NewCoins(amount))
	default:
		return sdkerrors.Wrap(types.ErrInvalidDepositType, string(depositType))
	}
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestWithdrawal,
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			sdk.NewAttribute(types.AttributeKeyDepositDenom, amount.Denom),
			sdk.NewAttribute(types.AttributeKeyDepositType, string(depositType)),
		),
	)

	if deposit.Amount.IsEqual(amount) {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDeleteHarvestDeposit,
				sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
				sdk.NewAttribute(types.AttributeKeyDepositDenom, amount.Denom),
				sdk.NewAttribute(types.AttributeKeyDepositType, string(depositType)),
			),
		)
		k.DeleteDeposit(ctx, deposit)
		return nil
	}

	deposit.Amount = deposit.Amount.Sub(amount)
	k.SetDeposit(ctx, deposit)

	return nil
}

// GetTotalDeposited returns the total amount deposited for the input deposit type and deposit denom
func (k Keeper) GetTotalDeposited(ctx sdk.Context, depositType types.DepositType, depositDenom string) (total sdk.Int) {

	var macc supplyExported.ModuleAccountI
	switch depositType {
	case types.LP:
		macc = k.supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	}
	return macc.GetCoins().AmountOf(depositDenom)
}
