package store

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// SetRewardIndexes stores the global reward indexes that track total rewards of
// a given claim type and collateralType.
func (k IncentiveStore) SetRewardIndexes(
	ctx sdk.Context,
	claimType types.ClaimType,
	collateralType string,
	indexes types.RewardIndexes,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetRewardIndexesKeyPrefix(claimType))
	bz := k.cdc.MustMarshal(&types.TypedRewardIndexes{
		ClaimType:      claimType,
		CollateralType: collateralType,
		RewardIndexes:  indexes,
	})
	store.Set(types.GetKeyFromSourceID(collateralType), bz)
}

// GetRewardIndexesOfClaimType fetches the global reward indexes that track total rewards
// of a given claimType and collateralType.
func (k IncentiveStore) GetRewardIndexesOfClaimType(
	ctx sdk.Context,
	claimType types.ClaimType,
	collateralType string,
) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetRewardIndexesKeyPrefix(claimType))
	bz := store.Get(types.GetKeyFromSourceID(collateralType))
	if bz == nil {
		return types.RewardIndexes{}, false
	}

	var proto types.TypedRewardIndexes
	k.cdc.MustUnmarshal(bz, &proto)
	return proto.RewardIndexes, true
}

// IterateRewardIndexesByClaimType iterates over all reward index objects in the store of a
// given ClaimType and performs a callback function.
func (k IncentiveStore) IterateRewardIndexesByClaimType(
	ctx sdk.Context,
	claimType types.ClaimType,
	cb func(types.TypedRewardIndexes) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetRewardIndexesKeyPrefix(claimType))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var typedRewardIndexes types.TypedRewardIndexes
		k.cdc.MustUnmarshal(iterator.Value(), &typedRewardIndexes)

		if cb(typedRewardIndexes) {
			break
		}
	}
}

// IterateRewardIndexes iterates over all reward index objects in the store
// of all ClaimTypes and performs a callback function.
func (k IncentiveStore) IterateRewardIndexes(
	ctx sdk.Context,
	cb func(types.TypedRewardIndexes) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var typedRewardIndexes types.TypedRewardIndexes
		k.cdc.MustUnmarshal(iterator.Value(), &typedRewardIndexes)

		if cb(typedRewardIndexes) {
			break
		}
	}
}

// GetRewardIndexes returns all reward indexes of any claimType.
func (k IncentiveStore) GetRewardIndexes(ctx sdk.Context) types.TypedRewardIndexesList {
	var tril types.TypedRewardIndexesList
	k.IterateRewardIndexes(
		ctx,
		func(typedRewardIndexes types.TypedRewardIndexes) bool {
			tril = append(tril, typedRewardIndexes)
			return false
		},
	)

	return tril
}
