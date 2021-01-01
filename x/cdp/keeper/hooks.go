package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/kava-labs/kava/x/bank"
	"github.com/kava-labs/kava/x/cdp/types"
)

// Implements StakingHooks interface
var _ types.CDPHooks = Keeper{}

// AfterCDPCreated - call hook if registered
func (k Keeper) AfterCDPCreated(ctx sdk.Context, cdp types.CDP) {
	if k.hooks != nil {
		k.hooks.AfterCDPCreated(ctx, cdp)
	}
}

// BeforeCDPModified - call hook if registered
func (k Keeper) BeforeCDPModified(ctx sdk.Context, cdp types.CDP) {
	if k.hooks != nil {
		k.hooks.BeforeCDPModified(ctx, cdp)
	}
}

// Hooks wrapper struct for hooks
type Hooks struct {
	k Keeper
}

// Hooks create new cdp hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// BeforeSend hook registered on bank keeper, runs before each transfer operation
func (h Hooks) BeforeSend(ctx sdk.Context, sender, receiver sdk.AccAddress, amount sdk.Coins) error {
	return h.k.SyncUSDXSavingsRateSend(ctx, sender, receiver, amount)
}

// BeforeMultiSend hook registered on bank keeper, runs before each multi-send operation
func (h Hooks) BeforeMultiSend(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	return h.k.SyncUSDXSavingsRateMultiSend(ctx, inputs, outputs)
}
