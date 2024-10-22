package keeper

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
	swaptypes "github.com/kava-labs/kava/x/swap/types"
)

// Hooks wrapper struct for hooks
type Hooks struct {
	k Keeper
}

var (
	_ cdptypes.CDPHooks         = Hooks{}
	_ hardtypes.HARDHooks       = Hooks{}
	_ stakingtypes.StakingHooks = Hooks{}
	_ swaptypes.SwapHooks       = Hooks{}
	_ savingstypes.SavingsHooks = Hooks{}
	_ earntypes.EarnHooks       = Hooks{}
)

// Hooks create new incentive hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// ------------------- Cdp Module Hooks -------------------

// AfterCDPCreated function that runs after a cdp is created
func (h Hooks) AfterCDPCreated(ctx sdk.Context, cdp cdptypes.CDP) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.k.InitializeUSDXMintingClaim(sdkCtx, cdp)
}

// BeforeCDPModified function that runs before a cdp is modified
// note that this is called immediately after interest is synchronized, and so could potentially
// be called AfterCDPInterestUpdated or something like that, if we we're to expand the scope of cdp hooks
func (h Hooks) BeforeCDPModified(ctx sdk.Context, cdp cdptypes.CDP) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.k.SynchronizeUSDXMintingReward(sdkCtx, cdp)
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
- validator becomes bonded (ie when they're promoted into the top 100)
  - total bonded delegation increases (tokens become bonded)

*/

// BeforeDelegationCreated runs before a delegation is created
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Add a claim if one doesn't exist, otherwise sync the existing.
	h.k.InitializeDelegatorReward(sdkCtx, delAddr)

	return nil
}

// BeforeDelegationSharesModified runs before an existing delegation is modified
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Sync rewards based on total delegated to bonded validators.
	h.k.SynchronizeDelegatorRewards(sdkCtx, delAddr, nil, false)

	return nil
}

// BeforeValidatorSlashed is called before a validator is slashed
// Validator status is not updated when Slash or Jail is called
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators.
	delegations, err := h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	for _, delegation := range delegations {
		h.k.SynchronizeDelegatorRewards(sdkCtx, []byte(delegation.GetDelegatorAddr()), nil, false)
	}

	return nil
}

// AfterValidatorBeginUnbonding is called after a validator begins unbonding
// Validator status is set to Unbonding prior to hook running
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	delegations, err := h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators, and also delegations to valAddr.
	// valAddr's status has just been set to Unbonding, but we want to include delegations to it in the sync.
	for _, delegation := range delegations {
		h.k.SynchronizeDelegatorRewards(sdkCtx, []byte(delegation.GetDelegatorAddr()), valAddr, true)
	}

	return nil
}

// AfterValidatorBonded is called after a validator is bonded
// Validator status is set to Bonded prior to hook running
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	fmt.Println("calling AfterValidatorBonded")
	delegations, err := h.k.stakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	fmt.Println("val addr", valAddr.String())
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators, except for delegations to valAddr.
	// valAddr's status has just been set to Bonded, but we don't want to include delegations to it in the sync
	fmt.Println("delegations: ", delegations)
	for _, delegation := range delegations {
		h.k.SynchronizeDelegatorRewards(sdkCtx, []byte(delegation.GetDelegatorAddr()), valAddr, false)
	}

	return nil
}

// NOTE: following hooks are just implemented to ensure StakingHooks interface compliance

// AfterDelegationModified runs after a delegation is modified
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved runs directly before a delegation is deleted. BeforeDelegationSharesModified is run prior to this.
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorCreated runs after a validator is created
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeValidatorModified runs before a validator is modified
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorRemoved runs after a validator is removed
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterUnbondingInitiated is called when an unbonding operation
// (validator unbonding, unbonding delegation, redelegation) was initiated
func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}

// ------------------- Swap Module Hooks -------------------

func (h Hooks) AfterPoolDepositCreated(ctx context.Context, poolID string, depositor sdk.AccAddress, _ sdkmath.Int) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.k.InitializeSwapReward(sdkCtx, poolID, depositor)
}

func (h Hooks) BeforePoolDepositModified(ctx context.Context, poolID string, depositor sdk.AccAddress, sharesOwned sdkmath.Int) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.k.SynchronizeSwapReward(sdkCtx, poolID, depositor, sharesOwned)
}

// ------------------- Savings Module Hooks -------------------

// AfterSavingsDepositCreated function that runs after a deposit is created
func (h Hooks) AfterSavingsDepositCreated(ctx context.Context, deposit savingstypes.Deposit) {
	// h.k.InitializeSavingsReward(ctx, deposit)
}

// BeforeSavingsDepositModified function that runs before a deposit is modified
func (h Hooks) BeforeSavingsDepositModified(ctx context.Context, deposit savingstypes.Deposit, incomingDenoms []string) {
	// h.k.SynchronizeSavingsReward(ctx, deposit, incomingDenoms)
}

// ------------------- Earn Module Hooks -------------------

// AfterVaultDepositCreated function that runs after a vault deposit is created
func (h Hooks) AfterVaultDepositCreated(
	ctx context.Context,
	vaultDenom string,
	depositor sdk.AccAddress,
	_ sdkmath.LegacyDec,
) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.k.InitializeEarnReward(sdkCtx, vaultDenom, depositor)
}

// BeforeVaultDepositModified function that runs before a vault deposit is modified
func (h Hooks) BeforeVaultDepositModified(
	ctx context.Context,
	vaultDenom string,
	depositor sdk.AccAddress,
	sharesOwned sdkmath.LegacyDec,
) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.k.SynchronizeEarnReward(sdkCtx, vaultDenom, depositor, sharesOwned)
}
