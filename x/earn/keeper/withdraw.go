package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/earn/types"
)

// Withdraw removes the amount of supplied tokens from a vault and transfers it
// back to the account.
func (k *Keeper) Withdraw(ctx sdk.Context, from sdk.AccAddress, wantAmount sdk.Coin) error {
	// Get AllowedVault, if not found (not a valid vault), return error
	allowedVault, found := k.GetAllowedVault(ctx, wantAmount.Denom)
	if !found {
		return types.ErrInvalidVaultDenom
	}

	if wantAmount.IsZero() {
		return types.ErrInsufficientAmount
	}

	// Check if VaultRecord exists
	vaultRecord, found := k.GetVaultRecord(ctx, wantAmount.Denom)
	if !found {
		return types.ErrVaultRecordNotFound
	}

	// Get account value for vault
	vaultAccValue, err := k.GetVaultAccountValue(ctx, wantAmount.Denom, from)
	if err != nil {
		return err
	}

	if vaultAccValue.IsZero() {
		panic("vault account value is zero")
	}

	// Get account share record for the vault
	vaultShareRecord, found := k.GetVaultShareRecord(ctx, wantAmount.Denom, from)
	if !found {
		return types.ErrVaultShareRecordNotFound
	}

	// Percent of vault account value the account is withdrawing
	// This is the total account value, not just the supplied amount.
	withdrawAmountPercent := wantAmount.Amount.Quo(vaultAccValue.Amount)

	// Check if account is not withdrawing more than they have
	// account value < want withdraw amount
	if vaultAccValue.Amount.LT(wantAmount.Amount) {
		return sdkerrors.Wrapf(
			types.ErrInsufficientValue,
			"account vault value of %s is less than %s desired withdraw amount",
			vaultAccValue,
			wantAmount,
		)
	}

	// Get the strategy for the vault
	strategy, err := k.GetStrategy(allowedVault.VaultStrategy)
	if err != nil {
		return err
	}

	// Not necessary to check if amount denom is allowed for the strategy, as
	// there would be no vault record if it weren't allowed.

	// Withdraw the wantAmount from the strategy
	if err := strategy.Withdraw(ctx, wantAmount); err != nil {
		return fmt.Errorf("failed to withdraw from strategy: %w", err)
	}

	// Send coins back to account, must withdraw from strategy first or the
	// module account may not have any funds to send.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		from,
		sdk.NewCoins(wantAmount),
	); err != nil {
		return err
	}

	// Shares withdrawn from vault
	// For example:
	// account supplied = 10hard
	// account value    = 20hard
	// wantAmount       = 10hard
	// withdrawAmountPercent = 10hard / 20hard = 0.5
	// sharesWithdrawn = 0.5 * 10hard = 5hard
	sharesWithdrawn := vaultShareRecord.AmountSupplied.Amount.Mul(withdrawAmountPercent)

	// Decrement VaultRecord and VaultShareRecord supplies
	vaultRecord.TotalSupply.Amount = vaultRecord.TotalSupply.Amount.Sub(sharesWithdrawn)
	vaultShareRecord.AmountSupplied.Amount = vaultShareRecord.AmountSupplied.Amount.Sub(sharesWithdrawn)

	// Update VaultRecord and VaultShareRecord, deletes if zero supply
	k.UpdateVaultRecord(ctx, vaultRecord)
	k.UpdateVaultShareRecord(ctx, vaultShareRecord)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVaultWithdraw,
			sdk.NewAttribute(types.AttributeKeyVaultDenom, wantAmount.Denom),
			sdk.NewAttribute(types.AttributeKeyOwner, from.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, wantAmount.Amount.String()),
		),
	)

	return nil
}
