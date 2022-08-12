package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/types"
)

// Implements EarnHooks interface
var _ types.EarnHooks = Keeper{}

// AfterVaultDepositCreated - call hook if registered
func (k Keeper) AfterVaultDepositCreated(
	ctx sdk.Context,
	vaultDenom string,
	depositor sdk.AccAddress,
	sharesOwned sdk.Dec,
) {
	if k.hooks != nil {
		k.hooks.AfterVaultDepositCreated(ctx, vaultDenom, depositor, sharesOwned)
	}
}

// BeforeVaultDepositModified - call hook if registered
func (k Keeper) BeforeVaultDepositModified(
	ctx sdk.Context,
	vaultDenom string,
	depositor sdk.AccAddress,
	sharesOwned sdk.Dec,
) {
	if k.hooks != nil {
		k.hooks.BeforeVaultDepositModified(ctx, vaultDenom, depositor, sharesOwned)
	}
}
