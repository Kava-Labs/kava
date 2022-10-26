package delegationsource

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/incentive/distributor"
	"github.com/kava-labs/kava/x/incentive/types"
)

var (
	_ distributor.SourceAdapter = &delegationAdapter{}
	_ stakingtypes.StakingHooks = &delegationAdapter{}
)

type delegationAdapter struct {
	stakingKeeper types.StakingKeeper
	listener      distributor.SharesUpdateListener
}

func New(stakingKeeper types.StakingKeeper) *delegationAdapter {
	sa := &delegationAdapter{
		stakingKeeper: stakingKeeper,
	}
	// keeper needs to be pointer, otherwise msgs won't trigger hooks
	// FIXME this doesn't work for staking hooks, as staking has other modules registered to it.
	// It would need an `AppendHooks` method, or the caller could register the hooks themselves.
	stakingKeeper.SetHooks(sa)
	return sa
}

func (da *delegationAdapter) GetTotalShares(ctx sdk.Context, sourceID string) sdk.Dec {
	totalBonded := da.stakingKeeper.TotalBondedTokens(ctx)

	return totalBonded.ToDec()
}

func (da *delegationAdapter) GetShares(ctx sdk.Context, sourceID string, owner sdk.AccAddress) sdk.Dec {
	return da.getTotalDelegated(ctx, owner, nil, false)
}

func (da *delegationAdapter) RegisterSharesUpdateListener(listener distributor.SharesUpdateListener) {
	da.listener = listener
}

// copied from rewards_delegator.go, probably a better way to do this
func (da *delegationAdapter) getTotalDelegated(ctx sdk.Context, delegator sdk.AccAddress, valAddr sdk.ValAddress, shouldIncludeValidator bool) sdk.Dec {
	totalDelegated := sdk.ZeroDec()

	delegations := da.stakingKeeper.GetDelegatorDelegations(ctx, delegator, 200)
	for _, delegation := range delegations {
		validator, found := da.stakingKeeper.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			continue
		}

		if validator.GetOperator().Equals(valAddr) {
			if shouldIncludeValidator {
				// do nothing, so the validator is included regardless of bonded status
			} else {
				// skip this validator
				continue
			}
		} else {
			// skip any not bonded validator
			if validator.GetStatus() != stakingtypes.Bonded {
				continue
			}
		}

		if validator.GetTokens().IsZero() {
			continue
		}

		delegatedTokens := validator.TokensFromShares(delegation.GetShares())
		if delegatedTokens.IsNegative() {
			continue
		}
		totalDelegated = totalDelegated.Add(delegatedTokens)
	}
	return totalDelegated
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
func (da *delegationAdapter) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	da.listener.SharesCreated(ctx, types.BondDenom, delAddr)
}

// BeforeDelegationSharesModified runs before an existing delegation is modified
func (da *delegationAdapter) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// Sync rewards based on total delegated to bonded validators.
	shares := da.getTotalDelegated(ctx, delAddr, nil, false)
	da.listener.SharesUpdated(ctx, types.BondDenom, delAddr, shares)
}

// BeforeValidatorSlashed is called before a validator is slashed
// Validator status is not updated when Slash or Jail is called
func (da *delegationAdapter) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators.
	for _, delegation := range da.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		shares := da.getTotalDelegated(ctx, delegation.GetDelegatorAddr(), nil, false)
		da.listener.SharesUpdated(ctx, types.BondDenom, delegation.GetDelegatorAddr(), shares)
	}
}

// AfterValidatorBeginUnbonding is called after a validator begins unbonding
// Validator status is set to Unbonding prior to hook running
func (da *delegationAdapter) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators, and also delegations to valAddr.
	// valAddr's status has just been set to Unbonding, but we want to include delegations to it in the sync.
	for _, delegation := range da.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		shares := da.getTotalDelegated(ctx, delegation.GetDelegatorAddr(), valAddr, true)
		da.listener.SharesUpdated(ctx, types.BondDenom, delegation.GetDelegatorAddr(), shares)
	}
}

// AfterValidatorBonded is called after a validator is bonded
// Validator status is set to Bonded prior to hook running
func (da *delegationAdapter) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	// Sync all claims for users delegated to this validator.
	// For each claim, sync based on the total delegated to bonded validators, except for delegations to valAddr.
	// valAddr's status has just been set to Bonded, but we don't want to include delegations to it in the sync
	for _, delegation := range da.stakingKeeper.GetValidatorDelegations(ctx, valAddr) {
		shares := da.getTotalDelegated(ctx, delegation.GetDelegatorAddr(), valAddr, false)
		da.listener.SharesUpdated(ctx, types.BondDenom, delegation.GetDelegatorAddr(), shares)
	}
}

// NOTE: following hooks are just implemented to ensure StakingHooks interface compliance

// AfterDelegationModified runs after a delegation is modified
func (da *delegationAdapter) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}

// BeforeDelegationRemoved runs directly before a delegation is deleted. BeforeDelegationSharesModified is run prior to this.
func (da *delegationAdapter) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}

// AfterValidatorCreated runs after a validator is created
func (da *delegationAdapter) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {}

// BeforeValidatorModified runs before a validator is modified
func (da *delegationAdapter) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}

// AfterValidatorRemoved runs after a validator is removed
func (da *delegationAdapter) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
