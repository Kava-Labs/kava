package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
)

// Hooks wrapper struct for hooks
type Hooks struct {
	k Keeper
}

var _ cdptypes.CDPHooks = Hooks{}
var _ hardtypes.HARDHooks = Hooks{}

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
// TODO: how to ensure that existing delegators get their rewards?

// delegate:
// 1a. if existing delegation: 	k.BeforeDelegationSharesModified(ctx, delAddr, validator.OperatorAddress)
// 1b. if new delegation: 		k.BeforeDelegationCreated(ctx, delAddr, validator.OperatorAddress)
// 2.  							k.AfterDelegationModified(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)

// unbond:
// 1	k.BeforeDelegationSharesModified(ctx, delAddr, valAddr)
// 2	k.AfterDelegationModified(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)

// BeforeDelegationCreated runs before a delegation is created
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// h.k.InitializeHardDelegationReward(ctx, delAddr, valAddr)
	// TODO: create delegation reward indexes inside hard reward object for the delegating address
}

// BeforeDelegationSharesModified runs before an existing delegation is modified
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// h.k.SynchronizeHardDelegationReward(ctx, borrow)
	// TODO: update delegation reward indexes inside hard reward object for the delegating address
}

// AfterDelegationModified runs after a delegation is modified
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// h.k.UpdateHardDelegationIndexDenoms(ctx, borrow)
	// TODO: update delegation reward indexes inside hard reward object for the delegating address
}

// BeforeDelegationRemoved runs directly before a delegation is deleted
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {

}
