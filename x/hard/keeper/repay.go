package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/types"
)

// Repay borrowed funds
func (k Keeper) Repay(ctx sdk.Context, sender, owner sdk.AccAddress, coins sdk.Coins) error {
	// Check borrow exists here to avoid duplicating store read in ValidateRepay
	borrow, found := k.GetBorrow(ctx, owner)
	if !found {
		return types.ErrBorrowNotFound
	}
	// Call incentive hook
	k.BeforeBorrowModified(ctx, borrow)

	// Sync borrow interest so loan is up-to-date
	k.SyncBorrowInterest(ctx, owner)

	// Validate that sender holds coins for repayment
	err = k.ValidateRepay(ctx, sender, coins)
	if err != nil {
		return err
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

	// If any coin denoms have been completely repaid reset the denom's borrow index factor
	for _, coin := range payment {
		if coin.Amount.Equal(borrow.Amount.AmountOf(coin.Denom)) {
			borrowIndex, removed := borrow.Index.RemoveInterestFactor(coin.Denom)
			if !removed {
				return sdkerrors.Wrapf(types.ErrInvalidIndexFactorDenom, "%s", coin.Denom)
			}
			borrow.Index = borrowIndex
		}
	}

	// Update user's borrow in store
	borrow.Amount = borrow.Amount.Sub(payment)

	if borrow.Amount.Empty() {
		k.DeleteBorrow(ctx, borrow)
	} else {
		k.SetBorrow(ctx, borrow)
	}

	// Update total borrowed amount
	k.DecrementBorrowedCoins(ctx, payment)

	// Call incentive hook
	if !borrow.Amount.Empty() {
		k.AfterBorrowModified(ctx, borrow)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardRepay,
			sdk.NewAttribute(types.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
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
