package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
)

// Hooks wrapper struct for hooks
type Hooks struct {
	k Keeper
}

var _ cdptypes.CDPHooks = Hooks{}
var _ hardtypes.HARDHooks = Hooks{}
var _ stakingtypes.StakingHooks = Hooks{}

// Hooks create new incentive hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// ------------------- Cdp Module Hooks -------------------

// AfterCDPCreated function that runs after a cdp is created
func (h Hooks) AfterCDPCreated(ctx sdk.Context, cdp cdptypes.CDP) {
	h.k.InitializeUSDXMintingClaim(ctx, cdp)
}

// BeforeCDPModified function that runs before a cdp is modified
// note that this is called immediately after interest is synchronized, and so could potentially
// be called AfterCDPInterestUpdated or something like that, if we we're to expand the scope of cdp hooks
func (h Hooks) BeforeCDPModified(ctx sdk.Context, cdp cdptypes.CDP) {
	h.k.SynchronizeUSDXMintingReward(ctx, cdp)
}

// ------------------- Hard Module Hooks -------------------

// AfterDepositCreated function that runs after a deposit is created
func (h Hooks) AfterDepositCreated(ctx sdk.Context, deposit hardtypes.Deposit) {
	h.k.InitializeHardSupplyReward(ctx, deposit)
}

// BeforeDepositModified function that runs before a deposit is modified
func (h Hooks) BeforeDepositModified(ctx sdk.Context, deposit hardtypes.Deposit) {
	h.k.SynchronizeHardSupplyReward(ctx, deposit)
}

// AfterDepositModified function that runs after a deposit is modified
func (h Hooks) AfterDepositModified(ctx sdk.Context, deposit hardtypes.Deposit) {
	h.k.UpdateHardSupplyIndexDenoms(ctx, deposit)
}

// AfterBorrowCreated function that runs after a borrow is created
func (h Hooks) AfterBorrowCreated(ctx sdk.Context, borrow hardtypes.Borrow) {
	h.k.InitializeHardBorrowReward(ctx, borrow)
}

// BeforeBorrowModified function that runs before a borrow is modified
func (h Hooks) BeforeBorrowModified(ctx sdk.Context, borrow hardtypes.Borrow) {
	h.k.SynchronizeHardBorrowReward(ctx, borrow)
}

// AfterBorrowModified function that runs after a borrow is modified
func (h Hooks) AfterBorrowModified(ctx sdk.Context, borrow hardtypes.Borrow) {
	h.k.UpdateHardBorrowIndexDenoms(ctx, borrow)
}

/* ------------------- Staking Module Hooks -------------------

Rewards are calculated based on total delegated tokens to bonded validators (not shares).
We need to sync the claim before the user's delegated tokens are changed.

When delegated tokens (to bonded validators) are changed:
- user creates new delegation
  - total bonded delegation increases
- user delegates or beginUnbonding or beginRedelegate an existing delegation
  - total bonded delegation increases or decreases
- validator is slashed and Jailed/Tombstoned (tokens reduce, and validator is unbonded)
  - slash: total bonded delegation decreases (less tokens)
  - jail: total bonded delegation decreases (tokens no longer bonded (after end blocker runs))
- validator becomes unbonded (ie when they drop out of the top 100)
  - total bonded delegation decreases (tokens no longer bonded)

*/

// BeforeDelegationCreated runs before a delegation is created
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// Add a claim if one doesn't exist, otherwise sync the existing.
	h.k.InitializeHardDelegatorReward(ctx, delAddr)
}

// BeforeDelegationSharesModified runs before an existing delegation is modified
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// Sync rewards based on total delegated to bonded validators.
	h.k.SynchronizeHardDelegatorRewards(ctx, delAddr, nil, false)
}

// BeforeValidatorSlashed is called before a validator is slashed
// Validator status is not updated when Slash or Jail is called
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators.
	for _, delegation := range h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		h.k.SynchronizeHardDelegatorRewards(ctx, delegation.DelegatorAddress, nil, false)
	}
}

// AfterValidatorBeginUnbonding is called after a validator begins unbonding
// Validator status is set to Unbonding prior to hook running
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators, and also delegations to valAddr.
	// valAddr's status has just been set to Unbonding, but we want to include delegations to it in the sync.
	for _, delegation := range h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		h.k.SynchronizeHardDelegatorRewards(ctx, delegation.DelegatorAddress, valAddr, true)
	}
}

// AfterValidatorBonded is called after a validator is bonded
// Validator status is set to Bonded prior to hook running
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators, except for delegations to valAddr.
	// valAddr's status has just been set to Bonded, but we don't want to include delegations to it in the sync
	for _, delegation := range h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		h.k.SynchronizeHardDelegatorRewards(ctx, delegation.DelegatorAddress, valAddr, false)
	}
}

// NOTE: following hooks are just implemented to ensure StakingHooks interface compliance

// AfterDelegationModified runs after a delegation is modified
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}

// BeforeDelegationRemoved runs directly before a delegation is deleted
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}

// AfterValidatorCreated runs after a validator is created
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {}

// BeforeValidatorModified runs before a validator is modified
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}

// AfterValidatorRemoved runs after a validator is removed
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
