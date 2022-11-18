package store

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// GetRewardAccrualTime fetches the last time rewards were accrued for the
// specified ClaimType and sourceID.
func (k IncentiveStore) GetRewardAccrualTime(
	ctx sdk.Context,
	claimType types.ClaimType,
	sourceID string,
) (time.Time, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetPreviousRewardAccrualTimeKeyPrefix(claimType))
	b := store.Get(types.GetKeyFromSourceID(sourceID))
	if b == nil {
		return time.Time{}, false
	}
	var accrualTime types.AccrualTime
	k.cdc.MustUnmarshal(b, &accrualTime)

	return accrualTime.PreviousAccumulationTime, true
}

// SetRewardAccrualTime stores the last time rewards were accrued for the
// specified ClaimType and sourceID.
func (k IncentiveStore) SetRewardAccrualTime(
	ctx sdk.Context,
	claimType types.ClaimType,
	sourceID string,
	blockTime time.Time,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetPreviousRewardAccrualTimeKeyPrefix(claimType))

	at := types.NewAccrualTime(claimType, sourceID, blockTime)
	bz := k.cdc.MustMarshal(&at)
	store.Set(types.GetKeyFromSourceID(sourceID), bz)
}

// IterateRewardAccrualTimesByClaimType iterates over all reward accrual times of a given
// claimType and performs a callback function.
func (k IncentiveStore) IterateRewardAccrualTimesByClaimType(
	ctx sdk.Context,
	claimType types.ClaimType,
	cb func(string, time.Time) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetPreviousRewardAccrualTimeKeyPrefix(claimType))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var accrualTime types.AccrualTime
		k.cdc.MustUnmarshal(iterator.Value(), &accrualTime)

		if cb(accrualTime.CollateralType, accrualTime.PreviousAccumulationTime) {
			break
		}
	}
}

// IterateRewardAccrualTimes iterates over all reward accrual times of any
// claimType and performs a callback function.
func (k IncentiveStore) IterateRewardAccrualTimes(
	ctx sdk.Context,
	cb func(types.AccrualTime) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousRewardAccrualTimeKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var accrualTime types.AccrualTime
		k.cdc.MustUnmarshal(iterator.Value(), &accrualTime)

		if cb(accrualTime) {
			break
		}
	}
}

// GetAllRewardAccrualTimes returns all reward accrual times of any claimType.
func (k IncentiveStore) GetAllRewardAccrualTimes(ctx sdk.Context) types.AccrualTimes {
	var ats types.AccrualTimes
	k.IterateRewardAccrualTimes(
		ctx,
		func(accrualTime types.AccrualTime) bool {
			ats = append(ats, accrualTime)
			return false
		},
	)

	return ats
}
