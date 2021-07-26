package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateDelegatorRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateDelegatorRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {

	previousAccrualTime, found := k.GetPreviousDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetDelegatorRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSource := k.getDelegatorTotalSourceShares(ctx, rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, totalSource, ctx.BlockTime())

	k.SetPreviousDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)
	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetDelegatorRewardIndexes(ctx, rewardPeriod.CollateralType, acc.Indexes)
	}
}

// getDelegatorTotalSourceShares fetches the sum of all source shares for a delegator reward.
// In the case of delegation, this is the total tokens staked to bonded validators.
func (k Keeper) getDelegatorTotalSourceShares(ctx sdk.Context, denom string) sdk.Dec {
	totalBonded := k.stakingKeeper.TotalBondedTokens(ctx)

	return totalBonded.ToDec()
}

// InitializeDelegatorReward initializes the reward index of a delegator claim
func (k Keeper) InitializeDelegatorReward(ctx sdk.Context, delegator sdk.AccAddress) {
	claim, found := k.GetDelegatorClaim(ctx, delegator)
	if !found {
		claim = types.NewDelegatorClaim(delegator, sdk.Coins{}, nil)
	} else {
		k.SynchronizeDelegatorRewards(ctx, delegator, nil, false)
		claim, _ = k.GetDelegatorClaim(ctx, delegator)
	}

	var rewardIndexes types.MultiRewardIndexes
	globalRewardIndexes, found := k.GetDelegatorRewardIndexes(ctx, types.BondDenom)
	if !found {
		globalRewardIndexes = types.RewardIndexes{}
	}
	rewardIndexes = rewardIndexes.With(types.BondDenom, globalRewardIndexes)
	claim.RewardIndexes = rewardIndexes
	k.SetDelegatorClaim(ctx, claim)
}

// SynchronizeDelegatorClaim is a wrapper around SynchronizeDelegatorRewards that returns the synced claim
func (k Keeper) SynchronizeDelegatorClaim(ctx sdk.Context, claim types.DelegatorClaim) (types.DelegatorClaim, error) {
	k.SynchronizeDelegatorRewards(ctx, claim.Owner, nil, false)

	claim, found := k.GetDelegatorClaim(ctx, claim.Owner)
	if !found {
		return claim, types.ErrClaimNotFound
	}
	return claim, nil
}

// SynchronizeDelegatorRewards updates the claim object by adding any accumulated rewards, and setting the reward indexes to the global values.
// valAddr and shouldIncludeValidator are used to ignore or include delegations to a particular validator when summing up the total delegation.
// Normally only delegations to Bonded validators are included in the total. This is needed as staking hooks are sometimes called on the wrong
// side of a validator's state update (from this module's perspective).
func (k Keeper) SynchronizeDelegatorRewards(ctx sdk.Context, delegator sdk.AccAddress, valAddr sdk.ValAddress, shouldIncludeValidator bool) {
	claim, found := k.GetDelegatorClaim(ctx, delegator)
	if !found {
		return
	}

	globalRewardIndexes, found := k.GetDelegatorRewardIndexes(ctx, types.BondDenom)
	if !found {
		// The global factor is only not found if
		// - the bond denom has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded denoms.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return
	}

	userRewardIndexes, found := claim.RewardIndexes.Get(types.BondDenom)
	if !found {
		// Normally the reward indexes should always be found.
		// However if there were no delegator rewards (ie no reward period in params) then a reward period is added, existing claims will not have the factor.
		// So given the reward period was just added, assume the starting value for any global reward indexes, which is an empty slice.
		userRewardIndexes = types.RewardIndexes{}
	}

	totalDelegated := k.GetTotalDelegated(ctx, delegator, valAddr, shouldIncludeValidator)

	rewardsEarned, err := k.CalculateRewards(userRewardIndexes, globalRewardIndexes, totalDelegated)
	if err != nil {
		// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
		// This panics if a global reward factor decreases or disappears between the old and new indexes.
		panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
	}

	claim.Reward = claim.Reward.Add(rewardsEarned...)
	claim.RewardIndexes = claim.RewardIndexes.With(types.BondDenom, globalRewardIndexes)
	k.SetDelegatorClaim(ctx, claim)
}

func (k Keeper) GetTotalDelegated(ctx sdk.Context, delegator sdk.AccAddress, valAddr sdk.ValAddress, shouldIncludeValidator bool) sdk.Dec {
	totalDelegated := sdk.ZeroDec()

	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delegator, 200)
	for _, delegation := range delegations {
		validator, found := k.stakingKeeper.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			continue
		}

		if validator.OperatorAddress.Equals(valAddr) {
			if shouldIncludeValidator {
				// do nothing, so the validator is included regardless of bonded status
			} else {
				// skip this validator
				continue
			}
		} else {
			// skip any not bonded validator
			if validator.GetStatus() != sdk.Bonded {
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

// SimulateDelegatorSynchronization calculates a user's outstanding delegator rewards by simulating reward synchronization
func (k Keeper) SimulateDelegatorSynchronization(ctx sdk.Context, claim types.DelegatorClaim) types.DelegatorClaim {
	for _, ri := range claim.RewardIndexes {
		// For each Delegator reward index (there's only one: the bond denom 'ukava')
		globalRewardIndexes, foundGlobalRewardIndexes := k.GetDelegatorRewardIndexes(ctx, ri.CollateralType)
		if !foundGlobalRewardIndexes {
			continue
		}

		userRewardIndexes, foundUserRewardIndexes := claim.RewardIndexes.GetRewardIndex(ri.CollateralType)
		if !foundUserRewardIndexes {
			continue
		}

		userRewardIndexIndex, foundUserRewardIndexIndex := claim.RewardIndexes.GetRewardIndexIndex(ri.CollateralType)
		if !foundUserRewardIndexIndex {
			continue
		}

		amtDelegated := k.GetTotalDelegated(ctx, claim.GetOwner(), sdk.ValAddress(claim.Owner.String()), true)

		for _, globalRewardIndex := range globalRewardIndexes {
			userRewardIndex, foundUserRewardIndex := userRewardIndexes.RewardIndexes.GetRewardIndex(globalRewardIndex.CollateralType)
			if !foundUserRewardIndex {
				userRewardIndex = types.NewRewardIndex(globalRewardIndex.CollateralType, sdk.ZeroDec())
				userRewardIndexes.RewardIndexes = append(userRewardIndexes.RewardIndexes, userRewardIndex)
				claim.RewardIndexes[userRewardIndexIndex].RewardIndexes = append(claim.RewardIndexes[userRewardIndexIndex].RewardIndexes, userRewardIndex)
			}

			globalRewardFactor := globalRewardIndex.RewardFactor
			userRewardFactor := userRewardIndex.RewardFactor
			rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
			if rewardsAccumulatedFactor.IsZero() {
				continue
			}

			rewardsEarned := rewardsAccumulatedFactor.Mul(amtDelegated).RoundInt()
			if rewardsEarned.IsZero() || rewardsEarned.IsNegative() {
				continue
			}

			factorIndex, foundFactorIndex := userRewardIndexes.RewardIndexes.GetFactorIndex(globalRewardIndex.CollateralType)
			if !foundFactorIndex {
				continue
			}
			claim.RewardIndexes[userRewardIndexIndex].RewardIndexes[factorIndex].RewardFactor = globalRewardIndex.RewardFactor
			newRewardsCoin := sdk.NewCoin(userRewardIndex.CollateralType, rewardsEarned)
			claim.Reward = claim.Reward.Add(newRewardsCoin)
		}
	}
	return claim
}
