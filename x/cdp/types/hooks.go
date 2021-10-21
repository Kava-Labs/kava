package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MultiCDPHooks combine multiple cdp hooks, all hook functions are run in array sequence
type MultiCDPHooks []CDPHooks

// NewMultiCDPHooks returns a new MultiCDPHooks
func NewMultiCDPHooks(hooks ...CDPHooks) MultiCDPHooks {
	return hooks
}

// BeforeCDPModified runs before a cdp is modified
func (h MultiCDPHooks) BeforeCDPModified(ctx sdk.Context, cdp CDP) {
	for i := range h {
		h[i].BeforeCDPModified(ctx, cdp)
	}
}

// AfterCDPCreated runs before a cdp is created
func (h MultiCDPHooks) AfterCDPCreated(ctx sdk.Context, cdp CDP) {
	for i := range h {
		h[i].AfterCDPCreated(ctx, cdp)
	}
}
