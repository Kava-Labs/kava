package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hvt/types"
)

// Deposit deposit
func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coin, depositType types.DepositType) error {

	err := k.ValidateDeposit(ctx, amount, depositType)
	if err != nil {
		return err
	}

	switch depositType {
	case types.LP:
		err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.LPAccount, sdk.NewCoins(amount))
	case types.Gov:
		err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.GovAccount, sdk.NewCoins(amount))
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

	return nil
}

// ValidateDeposit validates a deposit
func (k Keeper) ValidateDeposit(ctx sdk.Context, amount sdk.Coin, depositType types.DepositType) error {
	var err error
	switch depositType {
	case types.LP:
		err = k.ValidateLPDeposit(ctx, amount, depositType)
	case types.Gov:
		err = k.ValidateGovDeposit(ctx, amount, depositType)
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
	found := false
	for _, lps := range params.LiquidityProviderSchedules {
		if lps.DepositDenom == amount.Denom {
			found = true
		}
	}
	if found {
		return nil
	}
	return sdkerrors.Wrapf(types.ErrInvalidDepositDenom, "liquidity provider denom %s not found", amount.Denom)
}

// ValidateGovDeposit validates that a governance distribution deposit
func (k Keeper) ValidateGovDeposit(ctx sdk.Context, amount sdk.Coin, depositType types.DepositType) error {
	params := k.GetParams(ctx)
	found := false
	for _, gds := range params.GovernanceDistributionSchedules {
		if gds.DepositDenom == amount.Denom {
			found = true
		}
	}
	if found {
		return nil
	}
	return sdkerrors.Wrapf(types.ErrInvalidDepositDenom, "governance distribution denom %s not found", amount.Denom)
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
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.LPAccount, depositor, sdk.NewCoins(amount))
	case types.Gov:
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.GovAccount, depositor, sdk.NewCoins(amount))
	default:
		return sdkerrors.Wrap(types.ErrInvalidDepositType, string(depositType))
	}
	if err != nil {
		return err
	}
	if deposit.Amount.IsEqual(amount) {
		k.DeleteDeposit(ctx, deposit)
		return nil
	}
	deposit.Amount = deposit.Amount.Sub(amount)
	k.SetDeposit(ctx, deposit)

	return nil
}
