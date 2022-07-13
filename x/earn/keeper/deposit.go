package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k *Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coin) error {
	// Get AllowedVault

	// Check if VaultRecord exists, create if not exist

	// Transfer amount to module account

	// Increment VaultShareRecord for account, create if not exist

	// Increment VaultRecord supply

	return nil
}

func (k *Keeper) Withdraw(ctx sdk.Context, from sdk.AccAddress, amount sdk.Coin) error {

	return nil
}
