package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/swap/types"
)

func (k Keeper) GetPoolShares(ctx sdk.Context, poolID string) (sdk.Int, bool) {
	// FIXME return pool shares once merged with acceptance branch
	return sdk.Int{}, false
}

func (k *Keeper) GetDepositorSharesInPool(ctx sdk.Context, depositor sdk.AccAddress, poolID string) (sdk.Int, bool) {
	// FIXME return depositor shares once merged with acceptance branch
	return sdk.Int{}, false
}

// Implements SwapHooks interface
var _ types.SwapHooks = Keeper{}

// FIXME call hooks within pool logic

// AfterPoolDepositCreated - call hook if registered
func (k Keeper) AfterPoolDepositCreated(ctx sdk.Context, poolID string, depositor sdk.AccAddress, sharesOwned sdk.Int) {
	if k.hooks != nil {
		k.hooks.AfterPoolDepositCreated(ctx, poolID, depositor, sharesOwned)
	}
}

// BeforePoolDepositModified - call hook if registered
func (k Keeper) BeforePoolDepositModified(ctx sdk.Context, poolID string, depositor sdk.AccAddress, sharesOwned sdk.Int) {
	if k.hooks != nil {
		k.hooks.BeforePoolDepositModified(ctx, poolID, depositor, sharesOwned)
	}
}
