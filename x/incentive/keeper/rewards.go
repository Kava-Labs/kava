package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// IterateRewardPeriods iterates over all reward period objects in the store and preforms a callback function
func (k Keeper) IterateRewardPeriods(ctx sdk.Context, cb func(rp types.RewardPeriod) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var rp types.RewardPeriod
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &rp)
		if cb(rp) {
			break
		}
	}
}

// HandleRewardPeriodExpiry deletes expired RewardPeriods from the store and creates a ClaimPeriod in the store for each expired RewardPeriod
func (k Keeper) HandleRewardPeriodExpiry(ctx sdk.Context, rp types.RewardPeriod) {
	k.CreateClaimPeriod(ctx, rp.Denom, rp.ClaimEnd, rp.ClaimTimeLock)
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	store.Delete(types.GetDenomBytes(rp.Denom))
	return
}

// GetRewardPeriod returns the reward period from the store for the input denom and a boolean for if it was found
func (k Keeper) GetRewardPeriod(ctx sdk.Context, denom string) (types.RewardPeriod, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	bz := store.Get(types.GetDenomBytes(denom))
	if bz == nil {
		return types.RewardPeriod{}, false
	}
	var rp types.RewardPeriod
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &rp)
	return rp, true
}

// SetRewardPeriod sets the reward period in the store for the input deno,
func (k Keeper) SetRewardPeriod(ctx sdk.Context, rp types.RewardPeriod) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(rp)
	store.Set(types.GetDenomBytes(rp.Denom), bz)
}

// DeleteRewardPeriod deletes the reward period in the store for the input denom,
func (k Keeper) DeleteRewardPeriod(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	store.Delete(types.GetDenomBytes(denom))
}

// CreateNewRewardPeriod creates a new reward period from the input reward
func (k Keeper) CreateNewRewardPeriod(ctx sdk.Context, reward types.Reward) {
	// reward periods store the amount of rewards payed PER SECOND
	rewardsPerSecond := sdk.NewDecFromInt(reward.Reward.Amount).Quo(sdk.NewDecFromInt(sdk.NewInt(int64(reward.Duration.Seconds())))).TruncateInt()
	rewardCoinPerSecond := sdk.NewCoin(reward.Reward.Denom, rewardsPerSecond)
	rp := types.RewardPeriod{
		Denom:         reward.Denom,
		Start:         ctx.BlockTime(),
		End:           ctx.BlockTime().Add(reward.Duration),
		Reward:        rewardCoinPerSecond,
		ClaimEnd:      ctx.BlockTime().Add(reward.Duration).Add(reward.ClaimDuration),
		ClaimTimeLock: reward.TimeLock,
	}
	k.SetRewardPeriod(ctx, rp)
}

// CreateAndDeleteRewardPeriods creates reward periods for active rewards that don't already have a reward period and deletes reward periods for inactive rewards that currently have a reward period
func (k Keeper) CreateAndDeleteRewardPeriods(ctx sdk.Context) {
	params := k.GetParams(ctx)

	for _, r := range params.Rewards {
		_, found := k.GetRewardPeriod(ctx, r.Denom)
		// if governance has made a reward inactive, delete the current period
		if !r.Active {
			k.DeleteRewardPeriod(ctx, r.Denom)
		}
		// if a reward period for an active reward is not found, create one
		if !found {
			k.CreateNewRewardPeriod(ctx, r)
		}
	}
}

// ApplyRewardsToCdps iterates over the reward periods and creates a claim for each cdp owner that created usdx with the collateral specified in the reward period
func (k Keeper) ApplyRewardsToCdps(ctx sdk.Context) {
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
		k.SetPreviousBlockTime(ctx, previousBlockTime)
		return
	}
	k.IterateRewardPeriods(ctx, func(rp types.RewardPeriod) bool {
		expired := false
		// the total amount of usdx created with the collateral type being incentivized
		totalPrincipal := k.cdpKeeper.GetTotalPrincipal(ctx, rp.Denom, types.PrincipalDenom)
		// the number of seconds since last payout
		timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())
		if rp.End.After(ctx.BlockTime()) {
			timeElapsed = sdk.NewInt(rp.End.Unix() - previousBlockTime.Unix())
			expired = true
		}
		// the amount of rewards to pay (rewardAmount * timeElapsed)
		rewardsThisPeriod := rp.Reward.Amount.Mul(timeElapsed)
		id := k.GetNextClaimPeriodID(ctx, rp.Denom)
		k.cdpKeeper.IterateCdpsByDenom(ctx, rp.Denom, func(cdp cdptypes.CDP) bool {
			rewardsShare := sdk.NewDecFromInt(cdp.Principal.AmountOf(types.PrincipalDenom).Add(cdp.AccumulatedFees.AmountOf(types.PrincipalDenom))).Quo(sdk.NewDecFromInt(totalPrincipal))
			if rewardsShare.IsZero() {
				return false
			}
			rewardsEarned := rewardsShare.Mul(sdk.NewDecFromInt(rewardsThisPeriod)).TruncateInt()
			k.AddToClaim(ctx, cdp.Owner, types.GovDenom, id, sdk.NewCoin(types.GovDenom, rewardsEarned))
			return false
		})
		if expired {
			k.HandleRewardPeriodExpiry(ctx, rp)
		}
		return false
	})
}

// IterateClaimPeriodIDKeysAndValues iterates over the claim period id (value) and denom (key) of each claim period id in the store and performs a callback function
func (k Keeper) IterateClaimPeriodIDKeysAndValues(ctx sdk.Context, cb func(denom string, id uint64) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.NextClaimPeriodIDPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		id := types.BytesToUint64(iterator.Value())
		denom := types.GetDenomFromBytes(iterator.Key())
		if cb(denom, id) {
			break
		}
	}
}

// GetNextClaimPeriodID returns the highest claim period id in the store for the input denom
func (k Keeper) GetNextClaimPeriodID(ctx sdk.Context, denom string) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.key), types.NextClaimPeriodIDPrefix)
	bz := store.Get(types.GetDenomBytes(denom))
	id := types.BytesToUint64(bz)
	return id
}

// SetNextClaimPeriodID sets the highest claim period id in the store for the input denom
func (k Keeper) SetNextClaimPeriodID(ctx sdk.Context, denom string, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.NextClaimPeriodIDPrefix)
	store.Set(types.GetDenomBytes(denom), types.GetIDBytes(id))
}

// CreateClaimPeriod creates a new claim period in the store and updates the highest claim period id
func (k Keeper) CreateClaimPeriod(ctx sdk.Context, denom string, end time.Time, timeLock time.Duration) {
	// could call this CreateUniquePeriod, since it's a wrapper around NewClaimPeriod that ensures uniqueness
	id := k.GetNextClaimPeriodID(ctx, denom)
	claimPeriod := types.NewClaimPeriod(denom, id, end, timeLock)
	k.SetClaimPeriod(ctx, claimPeriod)
	k.SetNextClaimPeriodID(ctx, denom, id+1)
}

// IterateClaimPeriods iterates over all claim period objects in the store and preforms a callback function
func (k Keeper) IterateClaimPeriods(ctx sdk.Context, cb func(cp types.ClaimPeriod) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cp types.ClaimPeriod
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &cp)
		if cb(cp) {
			break
		}
	}
}

// GetClaimPeriod returns claim period in the store for the input ID and denom and a boolean for if it was found
func (k Keeper) GetClaimPeriod(ctx sdk.Context, id uint64, denom string) (types.ClaimPeriod, bool) {
	var cp types.ClaimPeriod
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	bz := store.Get(types.GetClaimPeriodPrefix(denom, id))
	if bz == nil {
		return types.ClaimPeriod{}, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cp)
	return cp, true
}

// SetClaimPeriod sets the claim period in the store for the input ID and denom
func (k Keeper) SetClaimPeriod(ctx sdk.Context, cp types.ClaimPeriod) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cp)
	store.Set(types.GetClaimPeriodPrefix(cp.Denom, cp.ID), bz)
}

// DeleteClaimPeriod deletes the claim period in the store for the input ID and denom
func (k Keeper) DeleteClaimPeriod(ctx sdk.Context, id uint64, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	store.Delete(types.GetClaimPeriodPrefix(denom, id))
}

// IterateClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateClaims(ctx sdk.Context, cb func(c types.Claim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.Claim
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetClaim returns the claim in the store corresponding the the input address denom and id and a boolean for if the claim was found
func (k Keeper) GetClaim(ctx sdk.Context, addr sdk.AccAddress, denom string, id uint64) (types.Claim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	bz := store.Get(types.GetClaimPrefix(addr, denom, id))
	if bz == nil {
		return types.Claim{}, false
	}
	var c types.Claim
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &c)
	return c, true
}

// SetClaim sets the claim in the store corresponding to the input address, denom, and id
func (k Keeper) SetClaim(ctx sdk.Context, c types.Claim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(c)
	store.Set(types.GetClaimPrefix(c.Owner, c.Denom, c.ID), bz)

}

// DeleteClaim deletes the claim in the store corresponding to the input address, denom, and id
func (k Keeper) DeleteClaim(ctx sdk.Context, owner sdk.AccAddress, denom string, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	store.Delete(types.GetClaimPrefix(owner, denom, id))
}

// AddToClaim adds the amount to an existing claim or creates a new on for that amount
func (k Keeper) AddToClaim(ctx sdk.Context, addr sdk.AccAddress, denom string, id uint64, amount sdk.Coin) {
	claim, found := k.GetClaim(ctx, addr, denom, id)
	if found {
		claim.Reward = claim.Reward.Add(amount)
		k.SetClaim(ctx, claim)
		return
	}
	claim = types.NewClaim(addr, amount, denom, id)
	k.SetClaim(ctx, claim)
}
