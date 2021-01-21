package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/hard/types"
)

// Withdraw returns some or all of a deposit back to original depositor
func (k Keeper) Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error {
	// Get current stored LTV based on stored borrows/deposits
	prevLtv, err := k.GetStoreLTV(ctx, depositor)
	if err != nil {
		return err
	}

	deposit, found := k.GetDeposit(ctx, depositor)
	if !found {
		return sdkerrors.Wrapf(types.ErrDepositNotFound, "no deposit found for %s", depositor)
	}
	// Call incentive hook
	k.BeforeDepositModified(ctx, deposit)

	k.SyncBorrowInterest(ctx, depositor)
	k.SyncSupplyInterest(ctx, depositor)

	amount, err := k.CalculateWithdrawAmount(deposit.Amount, coins)
	if err != nil {
		return err
	}

	borrow, found := k.GetBorrow(ctx, depositor)
	if !found {
		borrow = types.Borrow{}
	}

	proposedDeposit := types.NewDeposit(deposit.Depositor, deposit.Amount.Sub(amount), types.SupplyInterestFactors{})
	valid, err := k.IsWithinValidLtvRange(ctx, proposedDeposit, borrow)
	if err != nil {
		return err
	}
	if !valid {
		return sdkerrors.Wrapf(types.ErrInvalidWithdrawAmount, "proposed withdraw outside loan-to-value range")
	}

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, depositor, amount)
	if err != nil {
		return err
	}

	deposit.Amount = deposit.Amount.Sub(amount)
	newLtv, err := k.CalculateLtv(ctx, deposit, borrow)
	if err != nil {
		return err
	}
	k.UpdateDepositAndLtvIndex(ctx, deposit, newLtv, prevLtv)

	// Update total supplied amount
	k.DecrementBorrowedCoins(ctx, amount)

	// Call incentive hook
	k.AfterDepositModified(ctx, deposit)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardWithdrawal,
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
		),
	)
	return nil
}

// CalculateWithdrawAmount enables full withdraw of deposited coins by adjusting withdraw amount
// to equal total deposit amount if the requested withdraw amount > current deposit amount
func (k Keeper) CalculateWithdrawAmount(available sdk.Coins, request sdk.Coins) (sdk.Coins, error) {
	result := sdk.Coins{}

	if !request.DenomsSubsetOf(available) {
		return result, types.ErrInvalidWithdrawDenom
	}

	for _, coin := range request {
		if coin.Amount.GT(available.AmountOf(coin.Denom)) {
			result = append(result, sdk.NewCoin(coin.Denom, available.AmountOf(coin.Denom)))
		} else {
			result = append(result, coin)
		}
	}
	return result, nil
}
