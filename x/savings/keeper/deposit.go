package keeper

import (
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
	var amount sdk.Coins
	if foundDeposit {
		amount = currDeposit.Amount.Add(coins...)
	} else {
		amount = coins
	}
	deposit := types.NewDeposit(depositor, amount)

	k.SetDeposit(ctx, deposit)

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
		if supported == false {
			return types.ErrInvalidDepositDenom
		}
	}

	return nil
}

// GetTotalDeposited returns the total amount deposited for the deposit denom
func (k Keeper) GetTotalDeposited(ctx sdk.Context, depositDenom string) (total sdk.Int) {
	macc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	return k.bankKeeper.GetBalance(ctx, macc.GetAddress(), depositDenom).Amount
}
