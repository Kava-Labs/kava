package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/harvest/types"
)

// Repay borrowed funds
func (k Keeper) Repay(ctx sdk.Context, sender sdk.AccAddress, coins sdk.Coins) error {
	// Validate requested repay
	err := k.ValidateRepay(ctx, sender, coins)
	if err != nil {
		return err
	}

	// Sends coins from user to Harvest module account
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleAccountName, coins)
	if err != nil {
		return err
	}

	// Update user's borrow in store
	borrow, _ := k.GetBorrow(ctx, sender)
	borrow.Amount = borrow.Amount.Sub(coins)
	// TODO: Once interest functionality is merged update the user's borrow index here
	k.SetBorrow(ctx, borrow)

	// Update total borrowed amount
	k.DecrementBorrowedCoins(ctx, coins)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestRepay,
			sdk.NewAttribute(types.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyRepayCoins, coins.String()),
		),
	)

	return nil
}

// ValidateRepay validates a requested loan repay
func (k Keeper) ValidateRepay(ctx sdk.Context, sender sdk.AccAddress, coins sdk.Coins) error {
	senderAcc := k.accountKeeper.GetAccount(ctx, sender)
	senderCoins := senderAcc.GetCoins()
	if coins.IsAnyGT(senderCoins) {
		return sdkerrors.Wrapf(types.ErrInsufficientBalanceForRepay, "account can only repay up to %s", senderCoins)
	}

	borrow, found := k.GetBorrow(ctx, sender)
	if !found {
		return types.ErrBorrowNotFound
	}

	// TODO: Since interest accumulates every block users will be slightly *underpaying* the outstanding balance
	if coins.IsAnyGT(borrow.Amount) {
		return types.ErrDebtOverpaid
	}

	return nil
}
