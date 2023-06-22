package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/savings/types"
)

// Implements StakingHooks interface
var _ types.SavingsHooks = Keeper{}

// AfterSavingsDepositCreated - call hook if registered
func (k Keeper) AfterSavingsDepositCreated(ctx sdk.Context, deposit types.Deposit) {
	if k.hooks != nil {
		k.hooks.AfterSavingsDepositCreated(ctx, deposit)
	}
}

// BeforeSavingsDepositModified - call hook if registered
func (k Keeper) BeforeSavingsDepositModified(ctx sdk.Context, deposit types.Deposit, incomingDenoms []string) {
	if k.hooks != nil {
		k.hooks.BeforeSavingsDepositModified(ctx, deposit, incomingDenoms)
	}
}
