package keeper

import (
  sdk "github.com/cosmos/cosmos-sdk/types"

  "github.com/kava-labs/kava/x/hard/types"
)

// Implements StakingHooks interface
var _ types.HARDHooks = Keeper{}

// BeforeDepositCreated - call hook if registered
func (k Keeper) BeforeDepositCreated(ctx sdk.Context, deposit types.Deposit) {
  if k.hooks != nil {
    k.hooks.BeforeDepositCreated(ctx, deposit)
  }
}

// BeforeDepositModified - call hook if registered
func (k Keeper) BeforeDepositModified(ctx sdk.Context, deposit types.Deposit) {
  if k.hooks != nil {
    k.hooks.BeforeDepositModified(ctx, deposit)
  }
}

// BeforeBorrowCreated - call hook if registered
func (k Keeper) BeforeBorrowCreated(ctx sdk.Context, borrow types.Borrow) {
  if k.hooks != nil {
    k.hooks.BeforeBorrowCreated(ctx, borrow)
  }
}

// BeforeBorrowModified - call hook if registered
func (k Keeper) BeforeBorrowModified(ctx sdk.Context, borrow types.Borrow) {
  if k.hooks != nil {
    k.hooks.BeforeBorrowModified(ctx, borrow)
  }
}
