package keeper

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateRewards updates the rewards accumulated for the input reward period
func (k Keeper) AccumulateRewards(ctx sdk.Context, rewardPeriod types.RewardPeriod) error {
	previousAccrualTime, found := k.GetPreviousAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	timeElapsed := sdk.NewInt(int64(math.RoundToEven(
		ctx.BlockTime().Sub(previousAccrualTime).Seconds(),
	)))
	if timeElapsed.IsZero() {
		return nil
	}
	newRewards := timeElapsed.Mul(rewardPeriod.RewardsPerSecond.Amount)
	rewardFactor := newRewards.ToDec().Quo(k.cdpKeeper.GetTotalPrincipal(ctx, rewardPeriod.CollateralType, types.PrincipalDenom).ToDec())

	previousRewardFactor, found := k.GetRewardFactor(ctx, rewardPeriod.CollateralType)
	if !found {
		previousRewardFactor = sdk.ZeroDec()
	}
	newRewardFactor := previousRewardFactor.Add(rewardFactor)
	k.SetRewardFactor(ctx, rewardPeriod.CollateralType, newRewardFactor)
	return nil
}

// InitializeClaim creates or updates a claim such that no new rewards are accrued, but any existing rewards are not lost.
// this function should be called after a cdp is created. If a user previously had a cdp, then closed it, they shouldn't
// accrue rewards during the period the cdp was closed. By setting the reward factor to the current global reward factor,
// any unclaimed rewards are preserved, but no new rewards are added.
func (k Keeper) InitializeClaim(ctx sdk.Context, cdp cdptypes.CDP) {
	rewardFactor, found := k.GetRewardFactor(ctx, cdp.Type)
	if !found {
		rewardFactor = sdk.ZeroDec()
	}
	rewardIndex := types.NewRewardIndex("ukava", rewardFactor)
	claim, found := k.GetClaim(ctx, cdp.Owner, cdp.Type)
	if !found {
		claim = types.NewClaim(cdp.Owner, sdk.NewCoin("ukava", sdk.ZeroInt()), cdp.Type, rewardIndex)
	} else {
		claim.RewardIndex = rewardIndex
	}
	k.SetClaim(ctx, claim)
}

// SynchronizeReward updates the claim object by adding any accumulated rewards and updating the reward index value.
// this should be called before a cdp is modified, immediately after the 'SynchronizeInterest' method is called in the cdp module
func (k Keeper) SynchronizeReward(ctx sdk.Context, cdp cdptypes.CDP) {
	// User creates CDP, claims reward, which then deletes reward object
	// user modifies cdp or goes to claim rewards again, no existing claim. NOT SAFE!
	// ---> Claims CANNOT be deleted unless they are expired --> requires modification to Claim function
	globalRewardFactor, found := k.GetRewardFactor(ctx, cdp.Type)
	if !found {
		globalRewardFactor = sdk.ZeroDec()
	}
	rewardIndex := types.NewRewardIndex("ukava", globalRewardFactor)
	claim, found := k.GetClaim(ctx, cdp.Owner, cdp.Type)
	if !found {
		claim = types.NewClaim(cdp.Owner, sdk.NewCoin("ukava", sdk.ZeroInt()), cdp.Type, rewardIndex)
		k.SetClaim(ctx, claim)
		return
	}
	rewardsAccumulatedFactor := globalRewardFactor.Sub(claim.RewardIndex.Value)
	if rewardsAccumulatedFactor.IsZero() {
		return
	}
	claim.RewardIndex = rewardIndex
	newRewardsAmount := rewardsAccumulatedFactor.Mul(cdp.GetTotalPrincipal().Amount.ToDec()).RoundInt()
	if newRewardsAmount.IsZero() {
		k.SetClaim(ctx, claim)
		return
	}
	newRewardsCoin := sdk.NewCoin("ukava", newRewardsAmount)
	claim.Reward = claim.Reward.Add(newRewardsCoin)
	k.SetClaim(ctx, claim)
	return
}

// SynchronizeClaim updates the claim object by adding any rewards that have accumulated.
// Returns the updated claim object
func (k Keeper) SynchronizeClaim(ctx sdk.Context, claim types.Claim) (types.Claim, error) {
	cdp, found := k.cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, claim.Owner, claim.CollateralType)
	if !found {
		// if the cdp has been closed, no updates are needed
		return claim, nil
	}
	claim = k.synchronizeRewardAndReturnClaim(ctx, cdp)
	return claim, nil
}

// ZeroClaim zeroes out the claim object's rewards and returns the updated claim object
func (k Keeper) ZeroClaim(ctx sdk.Context, claim types.Claim) types.Claim {
	claim.Reward = sdk.NewCoin(claim.Reward.Denom, sdk.ZeroInt())
	k.SetClaim(ctx, claim)
	return claim
}

func (k Keeper) synchronizeRewardAndReturnClaim(ctx sdk.Context, cdp cdptypes.CDP) types.Claim {
	k.SynchronizeReward(ctx, cdp)
	claim, _ := k.GetClaim(ctx, cdp.Owner, cdp.Type)
	return claim
}
