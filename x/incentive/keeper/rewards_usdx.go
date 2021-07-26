package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateUSDXMintingRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateUSDXMintingRewards(ctx sdk.Context, rewardPeriod types.RewardPeriod) {
	previousAccrualTime, found := k.GetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	factor, found := k.GetUSDXMintingRewardFactor(ctx, rewardPeriod.CollateralType)
	if !found {
		factor = sdk.ZeroDec()
	}
	// wrap in RewardIndexes for compatibility with Accumulator
	indexes := types.RewardIndexes{}.With(types.USDXMintingRewardDenom, factor)

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSource := k.getUSDXTotalSourceShares(ctx, rewardPeriod.CollateralType)

	acc.Accumulate(types.NewMultiRewardPeriodFromRewardPeriod(rewardPeriod), totalSource, ctx.BlockTime())

	k.SetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)

	factor, found = acc.Indexes.Get(types.USDXMintingRewardDenom)
	if !found {
		panic("could not find factor that should never be missing when accumulating usdx rewards")
	}
	k.SetUSDXMintingRewardFactor(ctx, rewardPeriod.CollateralType, factor)
}

// getUSDXTotalSourceShares fetches the sum of all source shares for a usdx minting reward.
// In the case of usdx minting, this is the total debt from all cdps of a particular type, divided by the cdp interest factor.
// This gives the "pre interest" value of the total debt.
func (k Keeper) getUSDXTotalSourceShares(ctx sdk.Context, collateralType string) sdk.Dec {
	totalPrincipal := k.cdpKeeper.GetTotalPrincipal(ctx, collateralType, cdptypes.DefaultStableDenom)

	cdpFactor, found := k.cdpKeeper.GetInterestFactor(ctx, collateralType)
	if !found {
		// assume nothing has been borrowed so the factor starts at it's default value
		cdpFactor = sdk.OneDec()
	}
	// return debt/factor to get the "pre interest" value of the current total debt
	return totalPrincipal.ToDec().Quo(cdpFactor)
}

// InitializeUSDXMintingClaim creates or updates a claim such that no new rewards are accrued, but any existing rewards are not lost.
// this function should be called after a cdp is created. If a user previously had a cdp, then closed it, they shouldn't
// accrue rewards during the period the cdp was closed. By setting the reward factor to the current global reward factor,
// any unclaimed rewards are preserved, but no new rewards are added.
func (k Keeper) InitializeUSDXMintingClaim(ctx sdk.Context, cdp cdptypes.CDP) {
	claim, found := k.GetUSDXMintingClaim(ctx, cdp.Owner)
	if !found { // this is the owner's first usdx minting reward claim
		claim = types.NewUSDXMintingClaim(cdp.Owner, sdk.NewCoin(types.USDXMintingRewardDenom, sdk.ZeroInt()), types.RewardIndexes{})
	}

	globalRewardFactor, found := k.GetUSDXMintingRewardFactor(ctx, cdp.Type)
	if !found {
		globalRewardFactor = sdk.ZeroDec()
	}
	claim.RewardIndexes = claim.RewardIndexes.With(cdp.Type, globalRewardFactor)

	k.SetUSDXMintingClaim(ctx, claim)
}

// SynchronizeUSDXMintingReward updates the claim object by adding any accumulated rewards and updating the reward index value.
// this should be called before a cdp is modified.
func (k Keeper) SynchronizeUSDXMintingReward(ctx sdk.Context, cdp cdptypes.CDP) {

	claim, found := k.GetUSDXMintingClaim(ctx, cdp.Owner)
	if !found {
		return
	}

	sourceShares, err := cdp.GetNormalizedPrincipal()
	if err != nil {
		panic(fmt.Sprintf("during usdx reward sync, could not get normalized principal for %s: %s", cdp.Owner, err.Error()))
	}

	claim = k.synchronizeSingleUSDXMintingReward(ctx, claim, cdp.Type, sourceShares)

	k.SetUSDXMintingClaim(ctx, claim)
}

// synchronizeSingleUSDXMintingReward synchronizes a single rewarded cdp collateral type in a usdx minting claim.
// It returns the claim without setting in the store.
// The public methods for accessing and modifying claims are preferred over this one. Direct modification of claims is easy to get wrong.
func (k Keeper) synchronizeSingleUSDXMintingReward(ctx sdk.Context, claim types.USDXMintingClaim, ctype string, sourceShares sdk.Dec) types.USDXMintingClaim {

	globalRewardFactor, found := k.GetUSDXMintingRewardFactor(ctx, ctype)
	if !found {
		// The global factor is only not found if
		// - the cdp collateral type has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded denoms.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return claim
	}

	userRewardFactor, found := claim.RewardIndexes.Get(ctype)
	if !found {
		// Normally the factor should always be found, as it is added when the cdp is created in InitializeUSDXMintingClaim.
		// However if a cdp type is not rewarded then becomes rewarded (ie a reward period is added to params), existing cdps will not have the factor in their claims.
		// So assume the factor is the starting value for any global factor: 0.
		userRewardFactor = sdk.ZeroDec()
	}

	newRewardsAmount, err := k.CalculateSingleReward(userRewardFactor, globalRewardFactor, sourceShares)
	if err != nil {
		// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
		// This panics if a global reward factor decreases or disappears between the old and new indexes.
		panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
	}
	newRewardsCoin := sdk.NewCoin(types.USDXMintingRewardDenom, newRewardsAmount)

	claim.Reward = claim.Reward.Add(newRewardsCoin)
	claim.RewardIndexes = claim.RewardIndexes.With(ctype, globalRewardFactor)

	return claim
}

// SimulateUSDXMintingSynchronization calculates a user's outstanding USDX minting rewards by simulating reward synchronization
func (k Keeper) SimulateUSDXMintingSynchronization(ctx sdk.Context, claim types.USDXMintingClaim) types.USDXMintingClaim {
	for _, ri := range claim.RewardIndexes {
		_, found := k.GetUSDXMintingRewardPeriod(ctx, ri.CollateralType)
		if !found {
			continue
		}

		globalRewardFactor, found := k.GetUSDXMintingRewardFactor(ctx, ri.CollateralType)
		if !found {
			globalRewardFactor = sdk.ZeroDec()
		}

		// the owner has an existing usdx minting reward claim
		index, hasRewardIndex := claim.HasRewardIndex(ri.CollateralType)
		if !hasRewardIndex { // this is the owner's first usdx minting reward for this collateral type
			claim.RewardIndexes = append(claim.RewardIndexes, types.NewRewardIndex(ri.CollateralType, globalRewardFactor))
		}
		userRewardFactor := claim.RewardIndexes[index].RewardFactor
		rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
		if rewardsAccumulatedFactor.IsZero() {
			continue
		}

		claim.RewardIndexes[index].RewardFactor = globalRewardFactor

		cdp, found := k.cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, claim.GetOwner(), ri.CollateralType)
		if !found {
			continue
		}
		newRewardsAmount := rewardsAccumulatedFactor.Mul(cdp.GetTotalPrincipal().Amount.ToDec()).RoundInt()
		if newRewardsAmount.IsZero() {
			continue
		}
		newRewardsCoin := sdk.NewCoin(types.USDXMintingRewardDenom, newRewardsAmount)
		claim.Reward = claim.Reward.Add(newRewardsCoin)
	}

	return claim
}

// SynchronizeUSDXMintingClaim updates the claim object by adding any rewards that have accumulated.
// Returns the updated claim object
func (k Keeper) SynchronizeUSDXMintingClaim(ctx sdk.Context, claim types.USDXMintingClaim) (types.USDXMintingClaim, error) {
	for _, ri := range claim.RewardIndexes {
		cdp, found := k.cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, claim.Owner, ri.CollateralType)
		if !found {
			// if the cdp for this collateral type has been closed, no updates are needed
			continue
		}
		claim = k.synchronizeRewardAndReturnClaim(ctx, cdp)
	}
	return claim, nil
}

// this function assumes a claim already exists, so don't call it if that's not the case
func (k Keeper) synchronizeRewardAndReturnClaim(ctx sdk.Context, cdp cdptypes.CDP) types.USDXMintingClaim {
	k.SynchronizeUSDXMintingReward(ctx, cdp)
	claim, _ := k.GetUSDXMintingClaim(ctx, cdp.Owner)
	return claim
}

// ZeroUSDXMintingClaim zeroes out the claim object's rewards and returns the updated claim object
func (k Keeper) ZeroUSDXMintingClaim(ctx sdk.Context, claim types.USDXMintingClaim) types.USDXMintingClaim {
	claim.Reward = sdk.NewCoin(claim.Reward.Denom, sdk.ZeroInt())
	k.SetUSDXMintingClaim(ctx, claim)
	return claim
}
