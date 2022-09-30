package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/earn/types"
)

// Withdraw removes the amount of supplied tokens from a vault and transfers it
// back to the account.
func (k *Keeper) Withdraw(
	ctx sdk.Context,
	from sdk.AccAddress,
	wantAmount sdk.Coin,
	withdrawStrategy types.StrategyType,
) (sdk.Coin, error) {
	// Get AllowedVault, if not found (not a valid vault), return error
	allowedVault, found := k.GetAllowedVault(ctx, wantAmount.Denom)
	if !found {
		return sdk.Coin{}, types.ErrInvalidVaultDenom
	}

	if wantAmount.IsZero() {
		return sdk.Coin{}, types.ErrInsufficientAmount
	}

	// Check if withdraw strategy is supported by vault
	if !allowedVault.IsStrategyAllowed(withdrawStrategy) {
		return sdk.Coin{}, types.ErrInvalidVaultStrategy
	}

	// Check if VaultRecord exists
	vaultRecord, found := k.GetVaultRecord(ctx, wantAmount.Denom)
	if !found {
		return sdk.Coin{}, types.ErrVaultRecordNotFound
	}

	// Get account share record for the vault
	vaultShareRecord, found := k.GetVaultShareRecord(ctx, from)
	if !found {
		return sdk.Coin{}, types.ErrVaultShareRecordNotFound
	}

	withdrawShares, err := k.ConvertToShares(ctx, wantAmount)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to convert assets to shares: %w", err)
	}

	accCurrentShares := vaultShareRecord.Shares.AmountOf(wantAmount.Denom)
	// Check if account is not withdrawing more shares than they have
	if accCurrentShares.LT(withdrawShares.Amount) {
		return sdk.Coin{}, sdkerrors.Wrapf(
			types.ErrInsufficientValue,
			"account has less %s vault shares than withdraw shares, %s < %s",
			wantAmount.Denom,
			accCurrentShares,
			withdrawShares.Amount,
		)
	}

	// Convert shares to amount to get truncated true share value
	withdrawAmount, err := k.ConvertToAssets(ctx, withdrawShares)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to convert shares to assets: %w", err)
	}

	accountValue, err := k.GetVaultAccountValue(ctx, wantAmount.Denom, from)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to get account value: %w", err)
	}

	// Check if withdrawAmount > account value
	if withdrawAmount.Amount.GT(accountValue.Amount) {
		return sdk.Coin{}, sdkerrors.Wrapf(
			types.ErrInsufficientValue,
			"account has less %s vault value than withdraw amount, %s < %s",
			withdrawAmount.Denom,
			accountValue.Amount,
			withdrawAmount.Amount,
		)
	}

	// Get the strategy for the vault
	strategy, err := k.GetStrategy(allowedVault.Strategies[0])
	if err != nil {
		return sdk.Coin{}, err
	}

	// Not necessary to check if amount denom is allowed for the strategy, as
	// there would be no vault record if it weren't allowed.

	// Withdraw the withdrawAmount from the strategy
	if err := strategy.Withdraw(ctx, withdrawAmount); err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to withdraw from strategy: %w", err)
	}

	// Send coins back to account, must withdraw from strategy first or the
	// module account may not have any funds to send.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		from,
		sdk.NewCoins(withdrawAmount),
	); err != nil {
		return sdk.Coin{}, err
	}

	// Check if new account balance of shares results in account share value
	// of < 1 of a sdk.Coin. This share value is not able to be withdrawn and
	// should just be removed.
	isDust, err := k.ShareIsDust(
		ctx,
		vaultShareRecord.Shares.GetShare(withdrawAmount.Denom).Sub(withdrawShares),
	)
	if err != nil {
		return sdk.Coin{}, err
	}

	if isDust {
		// Modify withdrawShares to subtract entire share balance for denom
		// This does not modify the actual withdraw coin amount as the
		// difference is < 1coin.
		withdrawShares = vaultShareRecord.Shares.GetShare(withdrawAmount.Denom)
	}

	// Call hook before record is modified
	k.BeforeVaultDepositModified(ctx, wantAmount.Denom, from, vaultRecord.TotalShares.Amount)

	// Decrement VaultRecord and VaultShareRecord supplies - must delete same
	// amounts
	vaultShareRecord.Shares = vaultShareRecord.Shares.Sub(withdrawShares)
	vaultRecord.TotalShares = vaultRecord.TotalShares.Sub(withdrawShares)

	// Update VaultRecord and VaultShareRecord, deletes if zero supply
	k.UpdateVaultRecord(ctx, vaultRecord)
	k.UpdateVaultShareRecord(ctx, vaultShareRecord)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVaultWithdraw,
			sdk.NewAttribute(types.AttributeKeyVaultDenom, withdrawAmount.Denom),
			sdk.NewAttribute(types.AttributeKeyOwner, from.String()),
			sdk.NewAttribute(types.AttributeKeyShares, withdrawShares.Amount.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, withdrawAmount.Amount.String()),
		),
	)

	return withdrawAmount, nil
}

// WithdrawFromModuleAccount removes the amount of supplied tokens from a vault and transfers it
// back to the module account. The module account must be unblocked from receiving transfers.
func (k *Keeper) WithdrawFromModuleAccount(
	ctx sdk.Context,
	from string,
	wantAmount sdk.Coin,
	withdrawStrategy types.StrategyType,
) (sdk.Coin, error) {
	// Ensure the module account exists to prevent SendCoins from creating a new non-module account.
	acc := k.accountKeeper.GetModuleAccount(ctx, from)
	if acc == nil {
		return sdk.Coin{}, fmt.Errorf("module account not found: %s", from)
	}
	return k.Withdraw(ctx, acc.GetAddress(), wantAmount, withdrawStrategy)
}
