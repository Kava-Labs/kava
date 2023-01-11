package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/savings/types"
)

// Deposit deposit
func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error {
	err := k.ValidateDeposit(ctx, coins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, coins)
	if err != nil {
		return err
	}

	deposit, foundDeposit := k.GetDeposit(ctx, depositor)
	if foundDeposit {
		// Call hook with the deposit before it is modified
		newDenoms := setDifference(getDenoms(coins), getDenoms(deposit.Amount))
		k.BeforeSavingsDepositModified(ctx, deposit.Depositor, deposit.Amount, newDenoms)

		// Update existing deposit with new coins
		deposit.Amount = deposit.Amount.Add(coins...)
	} else {
		// Create new deposit with the provided coins
		deposit = types.NewDeposit(depositor, coins)
	}

	k.SetDeposit(ctx, deposit)

	if !foundDeposit {
		k.AfterSavingsDepositCreated(ctx, deposit.Depositor, deposit.Amount)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSavingsDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, coins.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, deposit.Depositor.String()),
		),
	)

	return nil
}

// ValidateDeposit validates a deposit
func (k Keeper) ValidateDeposit(ctx sdk.Context, coins sdk.Coins) error {
	for _, coin := range coins {
		supported := k.IsDenomSupported(ctx, coin.Denom)
		if !supported {
			return sdkerrors.Wrapf(types.ErrInvalidDepositDenom, ": %s", coin.Denom)
		}
	}

	return nil
}

// GetTotalDeposited returns the total amount deposited for the deposit denom
func (k Keeper) GetTotalDeposited(ctx sdk.Context, depositDenom string) (total sdk.Int) {
	macc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	return k.bankKeeper.GetBalance(ctx, macc.GetAddress(), depositDenom).Amount
}

// Set setDifference: A - B
func setDifference(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
