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

// ------------------- Staking Module Hooks -------------------

// BeforeDelegationCreated runs before a delegation is created
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.InitializeHardDelegatorReward(ctx, delAddr)
}

// BeforeDelegationSharesModified runs before an existing delegation is modified
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// TODO comments
	// sync rewards based on total delegated
	// total delegated is sum of total tokens delegated across all validators (from this delegator)
	// ignore unbonded/ing validators
	h.k.SynchronizeHardDelegatorRewards(ctx, delAddr, nil, false)
}

// BeforeValidatorSlashed is called before a validator is slashed
// Validator status is not updated when Slash or Jail is called
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	// TODO comments
	// get delgators
	// sync claim for delegator
	// sum bonded delegations
	for _, delegation := range h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		h.k.SynchronizeHardDelegatorRewards(ctx, delegation.DelegatorAddress, nil, false)
	}
}

// AfterValidatorBeginUnbonding is called after a validator begins unbonding
// Validator status is set to UnBonding prior to hook running
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	// TODO comments
	// get delgators
	// sync claim for delegator
	// sum bonded delegations, include valAddr even though it's unbonding
	for _, delegation := range h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		h.k.SynchronizeHardDelegatorRewards(ctx, delegation.DelegatorAddress, valAddr, true)
	}
}

// AfterValidatorBonded is called after a validator is bonded
// Validator status is set to Bonded prior to hook running
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	// TODO comments
	// get delegators
	// sync claim for delegator
	// sum up users bonded delegations, except for valAddr even though it's bonded
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
