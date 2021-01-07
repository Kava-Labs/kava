package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/types"
)

// Repay borrowed funds
func (k Keeper) Repay(ctx sdk.Context, sender sdk.AccAddress, coins sdk.Coins) error {
	// Get current stored LTV based on stored borrows/deposits
	prevLtv, err := k.GetStoreLTV(ctx, sender)
	if err != nil {
		return err
	}

	// Sync borrow interest so loan is up-to-date
	k.SyncBorrowInterest(ctx, sender)

	// Validate requested repay
	err = k.ValidateRepay(ctx, sender, coins)
	if err != nil {
		return err
	}

	// Check borrow exists here to avoid duplicating store read in ValidateRepay
	borrow, found := k.GetBorrow(ctx, sender)
	if !found {
		return types.ErrBorrowNotFound
	}

	payment, err := k.CalculatePaymentAmount(borrow.Amount, coins)
	if err != nil {
		return err
	}

	// Sends coins from user to Hard module account
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleAccountName, payment)
	if err != nil {
		return err
	}

	// Update user's borrow in store
	borrow.Amount = borrow.Amount.Sub(payment)
	k.SetBorrow(ctx, borrow)

	k.UpdateItemInLtvIndex(ctx, prevLtv, sender)

	// Update total borrowed amount
	k.DecrementBorrowedCoins(ctx, payment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardRepay,
			sdk.NewAttribute(types.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyRepayCoins, payment.String()),
		),
	)

	return nil
}

// ValidateRepay validates a requested loan repay
func (k Keeper) ValidateRepay(ctx sdk.Context, sender sdk.AccAddress, coins sdk.Coins) error {
	senderAcc := k.accountKeeper.GetAccount(ctx, sender)
	senderCoins := senderAcc.SpendableCoins(ctx.BlockTime())

	for _, coin := range coins {
		if senderCoins.AmountOf(coin.Denom).LT(coin.Amount) {
			return sdkerrors.Wrapf(types.ErrInsufficientBalanceForRepay, "account can only repay up to %s%s", senderCoins.AmountOf(coin.Denom), coin.Denom)
		}
	}

	return nil
}

// CalculatePaymentAmount prevents overpayment when repaying borrowed coins
func (k Keeper) CalculatePaymentAmount(owed sdk.Coins, payment sdk.Coins) (sdk.Coins, error) {
	repayment := sdk.Coins{}

	if !payment.DenomsSubsetOf(owed) {
		return repayment, types.ErrInvalidRepaymentDenom
	}

	for _, coin := range payment {
		if coin.Amount.GT(owed.AmountOf(coin.Denom)) {
			repayment = append(repayment, sdk.NewCoin(coin.Denom, owed.AmountOf(coin.Denom)))
		} else {
			repayment = append(repayment, coin)
		}
	}
	return repayment, nil
}
