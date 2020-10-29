package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/harvest/types"
)

// USDX is the USDX coin's denom
const USDX = "usdx"

// Borrow funds
func (k Keeper) Borrow(ctx sdk.Context, borrower sdk.AccAddress, amount sdk.Coin) error {
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, borrower, types.ModuleAccountName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	borrow, found := k.GetBorrow(ctx, borrower, amount.Denom)
	if !found {
		borrow = types.NewBorrow(borrower, amount)
	} else {
		borrow.Amount = borrow.Amount.Add(amount)
	}

	k.SetBorrow(ctx, borrow)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestBorrow,
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyBorrower, borrow.Borrower.String()),
			sdk.NewAttribute(types.AttributeKeyBorrowDenom, borrow.Amount.Denom),
		),
	)

	return nil
}
