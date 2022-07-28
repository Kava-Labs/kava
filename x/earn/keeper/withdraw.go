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

	// Get account share record for the vault
	vaultShareRecord, found := k.GetVaultShareRecord(ctx, from)
	if !found {
		return types.ErrVaultShareRecordNotFound
	}

	withdrawShares, err := k.ConvertToShares(ctx, wantAmount)
	if err != nil {
		return fmt.Errorf("failed to convert assets to shares: %w", err)
	}

	// Check if account is not withdrawing more shares than they have
	if vaultShareRecord.Shares.AmountOf(wantAmount.Denom).LT(withdrawShares.Amount) {
		return sdkerrors.Wrapf(
			types.ErrInsufficientValue,
			"account vault shares of %s is %s but withdraw shares is %s",
			wantAmount.Denom,
			wantAmount,
			withdrawShares.Amount,
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

	// Decrement VaultRecord and VaultShareRecord supplies
	vaultRecord.TotalShares = vaultRecord.TotalShares.Sub(withdrawShares)
	vaultShareRecord.Shares = vaultShareRecord.Shares.Sub(withdrawShares)

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
