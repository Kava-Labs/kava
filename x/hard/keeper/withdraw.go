package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/types"
)

// Withdraw returns some or all of a deposit back to original depositor
func (k Keeper) Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error {
	// Get current stored LTV based on stored borrows/deposits
	prevLtv, shouldRemoveIndex, err := k.GetStoreLTV(ctx, depositor)
	if err != nil {
		return err
	}

	k.SyncBorrowInterest(ctx, depositor)
	k.SyncSupplyInterest(ctx, depositor)

	deposit, found := k.GetDeposit(ctx, depositor)
	if !found {
		return sdkerrors.Wrapf(types.ErrDepositNotFound, "no deposit found for %s", depositor)
	}

	borrow, found := k.GetBorrow(ctx, depositor)
	if !found {
		borrow = types.Borrow{}
	}

	proposedDepositAmount, isNegative := deposit.Amount.SafeSub(coins)
	if isNegative {
		return types.ErrNegativeBorrowedCoins
	}
	proposedDeposit := types.NewDeposit(deposit.Depositor, proposedDepositAmount, types.SupplyInterestFactors{})

	valid, err := k.IsWithinValidLtvRange(ctx, proposedDeposit, borrow)
	if err != nil {
		return err
	}

	if !valid {
		return sdkerrors.Wrapf(types.ErrInvalidWithdrawAmount, "proposed withdraw outside loan-to-value range")
	}

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, depositor, coins)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardWithdrawal,
			sdk.NewAttribute(sdk.AttributeKeyAmount, coins.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
		),
	)

	if deposit.Amount.IsEqual(coins) {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDeleteHardDeposit,
				sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			),
		)
		k.DeleteDeposit(ctx, deposit)
		return nil
	}

	deposit.Amount = deposit.Amount.Sub(coins)
	k.SetDeposit(ctx, deposit)

	k.UpdateItemInLtvIndex(ctx, prevLtv, shouldRemoveIndex, depositor)

	return nil
}
