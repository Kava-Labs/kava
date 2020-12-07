package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/harvest/types"
)

// Repay borrowed funds
func (k Keeper) Repay(ctx sdk.Context, sender sdk.AccAddress, coins sdk.Coins) error {
	// Sync interest so loan is up-to-date
	k.SyncBorrowInterest(ctx, sender, coins)

	// Validate requested repay
	err := k.ValidateRepay(ctx, sender, coins)
	if err != nil {
		return err
	}

	// Check borrow exists here to avoid duplicating store read in ValidateRepay
	borrow, found := k.GetBorrow(ctx, sender)
	if !found {
		return types.ErrBorrowNotFound
	}

	payment := k.CalculatePaymentAmount(borrow.Amount, coins)

	// Sends coins from user to Harvest module account
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleAccountName, payment)
	if err != nil {
		return err
	}

	// Update user's borrow in store
	borrow.Amount = borrow.Amount.Sub(payment)
	// TODO: Once interest functionality is merged update the user's borrow index here
	k.SetBorrow(ctx, borrow)

	// Update total borrowed amount
	k.DecrementBorrowedCoins(ctx, payment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestRepay,
			sdk.NewAttribute(types.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyRepayCoins, payment.String()),
		),
	)

	return nil
}

// ValidateRepay validates a requested loan repay
func (k Keeper) ValidateRepay(ctx sdk.Context, sender sdk.AccAddress, coins sdk.Coins) error {
	senderAcc := k.accountKeeper.GetAccount(ctx, sender)
	senderCoins := senderAcc.GetCoins()
	for _, coin := range coins {
		if senderCoins.AmountOf(coin.Denom).LT(coin.Amount) {
			return sdkerrors.Wrapf(types.ErrInsufficientBalanceForRepay, "account can only repay up to %s%s", senderCoins.AmountOf(coin.Denom), coin.Denom)
		}
	}

	return nil
}

// CalculatePaymentAmount prevents overpayment when repaying borrowed coins
func (k Keeper) CalculatePaymentAmount(owed sdk.Coins, payment sdk.Coins) sdk.Coins {
	repayment := sdk.Coins{}
	for _, coin := range payment {
		if coin.Amount.GT(owed.AmountOf(coin.Denom)) {
			repayment.Add(sdk.NewCoin(coin.Denom, owed.AmountOf(coin.Denom)))
		} else {
			repayment.Add(coin)
		}
	}
	return repayment
}
