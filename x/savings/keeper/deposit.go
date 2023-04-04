package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

	currDeposit, foundDeposit := k.GetDeposit(ctx, depositor)

	deposit := types.NewDeposit(depositor, coins)
	if foundDeposit {
		deposit.Amount = deposit.Amount.Add(currDeposit.Amount...)
		k.BeforeSavingsDepositModified(ctx, deposit, setDifference(getDenoms(coins), getDenoms(deposit.Amount)))

	}

	k.SetDeposit(ctx, deposit)

	if !foundDeposit {
		k.AfterSavingsDepositCreated(ctx, deposit)
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
			return errorsmod.Wrapf(types.ErrInvalidDepositDenom, ": %s", coin.Denom)
		}
	}

	return nil
}

// GetTotalDeposited returns the total amount deposited for the deposit denom
func (k Keeper) GetTotalDeposited(ctx sdk.Context, depositDenom string) (total sdkmath.Int) {
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

func getDenoms(coins sdk.Coins) []string {
	denoms := []string{}
	for _, coin := range coins {
		denoms = append(denoms, coin.Denom)
	}
	return denoms
}
