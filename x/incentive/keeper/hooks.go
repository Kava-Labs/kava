package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// Hooks wrapper struct for hooks
type Hooks struct {
	k Keeper
}

var (
	_ cdptypes.CDPHooks         = Hooks{}
	_ hardtypes.HARDHooks       = Hooks{}
	_ savingstypes.SavingsHooks = Hooks{}
	_ earntypes.EarnHooks       = Hooks{}
)

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

// ------------------- Savings Module Hooks -------------------

// AfterSavingsDepositCreated function that runs after a deposit is created
func (h Hooks) AfterSavingsDepositCreated(ctx sdk.Context, deposit savingstypes.Deposit) {
	// h.k.InitializeSavingsReward(ctx, deposit)
}

// BeforeSavingsDepositModified function that runs before a deposit is modified
func (h Hooks) BeforeSavingsDepositModified(ctx sdk.Context, deposit savingstypes.Deposit, incomingDenoms []string) {
	// h.k.SynchronizeSavingsReward(ctx, deposit, incomingDenoms)
}

// ------------------- Earn Module Hooks -------------------

// AfterVaultDepositCreated function that runs after a vault deposit is created
func (h Hooks) AfterVaultDepositCreated(
	ctx sdk.Context,
	vaultDenom string,
	depositor sdk.AccAddress,
	_ sdk.Dec,
) {
	h.k.InitializeEarnReward(ctx, vaultDenom, depositor)
}

// BeforeVaultDepositModified function that runs before a vault deposit is modified
func (h Hooks) BeforeVaultDepositModified(
	ctx sdk.Context,
	vaultDenom string,
	depositor sdk.AccAddress,
	sharesOwned sdk.Dec,
) {
	h.k.SynchronizeEarnReward(ctx, vaultDenom, depositor, sharesOwned)
}
