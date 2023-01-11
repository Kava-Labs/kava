package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/savings/types"
)

// Implements StakingHooks interface
var _ types.SavingsHooks = Keeper{}

// AfterSavingsDepositCreated - call hook if registered
func (k Keeper) AfterSavingsDepositCreated(ctx sdk.Context, addr sdk.AccAddress, depositCoins sdk.Coins) {
	if k.hooks != nil {
		k.hooks.AfterSavingsDepositCreated(ctx, addr, depositCoins)
	}
}

// BeforeSavingsDepositModified - call hook if registered
func (k Keeper) BeforeSavingsDepositModified(ctx sdk.Context, addr sdk.AccAddress, depositCoins sdk.Coins, newDenoms []string) {
	if k.hooks != nil {
		k.hooks.BeforeSavingsDepositModified(ctx, addr, depositCoins, newDenoms)
	}
}

// ClearHooks clears the hooks on the keeper
func (k *Keeper) ClearHooks() {
	k.hooks = nil
}
