package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// Deposit adds the provided amount from a depositor to a vault. The vault is
// specified by the denom in the amount.
func (k *Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coin) error {
	// Get AllowedVault, if not found (not a valid vault), return error
	allowedVault, found := k.GetAllowedVault(ctx, amount.Denom)
	if !found {
		return types.ErrInvalidVaultDenom
	}

	if amount.IsZero() {
		return types.ErrInsufficientAmount
	}

	// Check if VaultRecord exists, create if not exist
	vaultRecord, found := k.GetVaultRecord(ctx, amount.Denom)
	if !found {
		// Create a new VaultRecord with 0 supply
		vaultRecord = types.NewVaultRecord(amount.Denom)
	}

	// Get the strategy for the vault
	strategy, err := k.GetStrategy(allowedVault.VaultStrategy)
	if err != nil {
		return err
	}

	// Transfer amount to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		depositor,
		types.ModuleName,
		sdk.NewCoins(amount),
	); err != nil {
		return err
	}

	// Get VaultShareRecord for account, create if not exist
	vaultShareRecord, found := k.GetVaultShareRecord(ctx, depositor)
	if !found {
		// Create a new empty VaultShareRecord with 0 supply
		vaultShareRecord = types.NewVaultShareRecord(depositor, types.NewVaultShares())
	}

	shares, err := k.ConvertToShares(ctx, amount)
	if err != nil {
		return fmt.Errorf("failed to convert assets to shares: %w", err)
	}

	// Increment VaultRecord total shares and account shares
	vaultRecord.TotalShares = vaultRecord.TotalShares.Add(shares)
	vaultShareRecord.Shares = vaultShareRecord.Shares.Add(shares)

	// Update VaultRecord and VaultShareRecord
	k.SetVaultRecord(ctx, vaultRecord)
	k.SetVaultShareRecord(ctx, vaultShareRecord)

	// Deposit to the strategy
	if err := strategy.Deposit(ctx, amount); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVaultDeposit,
			sdk.NewAttribute(types.AttributeKeyVaultDenom, amount.Denom),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.Amount.String()),
		),
	)

	return nil
}
