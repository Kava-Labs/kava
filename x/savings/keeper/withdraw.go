package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/savings/types"
)

// Withdraw returns some or all of a deposit back to original depositor
func (k Keeper) Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error {
	deposit, found := k.GetDeposit(ctx, depositor)
	if !found {
		return sdkerrors.Wrap(types.ErrNoDepositFound, fmt.Sprintf(" for address: %s", depositor.String()))
	}

	amount, err := k.CalculateWithdrawAmount(deposit.Amount, coins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, depositor, amount)
	if err != nil {
		return err
	}

	deposit.Amount = deposit.Amount.Sub(amount)
	if deposit.Amount.Empty() {
		k.DeleteDeposit(ctx, deposit)
	} else {
		k.SetDeposit(ctx, deposit)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSavingsWithdrawal,
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
