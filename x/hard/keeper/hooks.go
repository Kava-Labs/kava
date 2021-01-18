package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hard/types"
)

// Implements StakingHooks interface
var _ types.HARDHooks = Keeper{}

// BeforeDepositCreated - call hook if registered
func (k Keeper) BeforeDepositCreated(ctx sdk.Context, deposit types.Deposit, denom string) {
	if k.hooks != nil {
		k.hooks.BeforeDepositCreated(ctx, deposit, denom)
	}
}

// BeforeDepositModified - call hook if registered
func (k Keeper) BeforeDepositModified(ctx sdk.Context, deposit types.Deposit, denom string) {
	if k.hooks != nil {
		k.hooks.BeforeDepositModified(ctx, deposit, denom)
	}
}

// AfterDepositModified - call hook if registered
func (k Keeper) AfterDepositModified(ctx sdk.Context, deposit types.Deposit) {
	if k.hooks != nil {
		k.hooks.AfterDepositModified(ctx, deposit)
	}
}

// BeforeBorrowCreated - call hook if registered
func (k Keeper) BeforeBorrowCreated(ctx sdk.Context, borrow types.Borrow, denom string) {
	if k.hooks != nil {
		k.hooks.BeforeBorrowCreated(ctx, borrow, denom)
	}
}

// BeforeBorrowModified - call hook if registered
func (k Keeper) BeforeBorrowModified(ctx sdk.Context, borrow types.Borrow, denom string) {
	if k.hooks != nil {
		k.hooks.BeforeBorrowModified(ctx, borrow, denom)
	}
}

// AfterBorrowModified - call hook if registered
func (k Keeper) AfterBorrowModified(ctx sdk.Context, borrow types.Borrow) {
	if k.hooks != nil {
		k.hooks.AfterBorrowModified(ctx, borrow)
	}
}
