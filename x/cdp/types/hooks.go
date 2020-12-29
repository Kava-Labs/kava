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

// Accumulator that represents the units of ukava owed to each unit of minted usdx for that collateral type
// in begin blocker, call 'GetTotalPrincipal' for each incentivized pool, then update the accumulator based on the amount of time that has elapsed since last block

// for individual cdps

// after cdp created - run 'AfterCDPCreated', which creates/updates a claim object and sets the 'claim.RewardFactor' to the current globalRewardFactor
// before cdp modified - run 'BeforeCDPModified', which calls 'SynchronizeRewards(ctx, cdp)`, which updates the claim.Rewards to be: cdp.GetTotalPrincipal() * (globalRewardFactor - claim.RewardFactor) and sets claim.RewardFactor to the current globalRewardFactor
// draw/repay/deposit/withdraw/liquidate - note that synchronize interest doesn't count as a modification?? Ie, you want to synchronize interest, THEN pass the resulting CDP to 'BeforeCDPModified', then proceed to update CDP.
