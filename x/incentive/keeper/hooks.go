package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// Hooks wrapper struct for hooks
type Hooks struct {
	k Keeper
}

var _ cdptypes.CDPHooks = Hooks{}

// Hooks create new incentive hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// AfterCDPCreated function that runs after a cdp is created
func (h Hooks) AfterCDPCreated(ctx sdk.Context, cdp cdptypes.CDP) {
	h.k.InitializeClaim(ctx, cdp)
}

// BeforeCDPModified function that runs before a cdp is modified
// note that this is called immediately after interest is synchronized, and so could potentially
// be called AfterCDPInterestUpdated or something like that, if we we're to expand the scope of cdp hooks
func (h Hooks) BeforeCDPModified(ctx sdk.Context, cdp cdptypes.CDP) {
	h.k.SynchronizeReward(ctx, cdp)
}
