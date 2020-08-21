package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/incentive/types"
)

// Keeper keeper for the incentive module
type Keeper struct {
	accountKeeper types.AccountKeeper
	cdc           *codec.Codec
	cdpKeeper     types.CdpKeeper
	key           sdk.StoreKey
	paramSubspace subspace.Subspace
	supplyKeeper  types.SupplyKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, sk types.SupplyKeeper,
	cdpk types.CdpKeeper, ak types.AccountKeeper,
) Keeper {

	return Keeper{
		accountKeeper: ak,
		cdc:           cdc,
		cdpKeeper:     cdpk,
		key:           key,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
		supplyKeeper:  sk,
	}
}

// GetRewardPeriod returns the reward period from the store for the input collateral type and a boolean for if it was found
func (k Keeper) GetRewardPeriod(ctx sdk.Context, collateralType string) (types.RewardPeriod, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	bz := store.Get([]byte(collateralType))
	if bz == nil {
		return types.RewardPeriod{}, false
	}
	var rp types.RewardPeriod
	k.cdc.MustUnmarshalBinaryBare(bz, &rp)
	return rp, true
}

// SetRewardPeriod sets the reward period in the store for the input deno,
func (k Keeper) SetRewardPeriod(ctx sdk.Context, rp types.RewardPeriod) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(rp)
	store.Set([]byte(rp.CollateralType), bz)
}

// DeleteRewardPeriod deletes the reward period in the store for the input collateral type,
func (k Keeper) DeleteRewardPeriod(ctx sdk.Context, collateralType string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	store.Delete([]byte(collateralType))
}

// IterateRewardPeriods iterates over all reward period objects in the store and preforms a callback function
func (k Keeper) IterateRewardPeriods(ctx sdk.Context, cb func(rp types.RewardPeriod) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardPeriodKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var rp types.RewardPeriod
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &rp)
		if cb(rp) {
			break
		}
	}
}

// GetAllRewardPeriods returns all reward periods in the store
func (k Keeper) GetAllRewardPeriods(ctx sdk.Context) types.RewardPeriods {
	rps := types.RewardPeriods{}
	k.IterateRewardPeriods(ctx, func(rp types.RewardPeriod) (stop bool) {
		rps = append(rps, rp)
		return false
	})
	return rps
}

// GetNextClaimPeriodID returns the highest claim period id in the store for the input collateral type
func (k Keeper) GetNextClaimPeriodID(ctx sdk.Context, collateralType string) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.key), types.NextClaimPeriodIDPrefix)
	bz := store.Get([]byte(collateralType))
	if bz == nil {
		k.SetNextClaimPeriodID(ctx, collateralType, 1)
		return uint64(1)
	}
	return types.BytesToUint64(bz)
}

// SetNextClaimPeriodID sets the highest claim period id in the store for the input collateral type
func (k Keeper) SetNextClaimPeriodID(ctx sdk.Context, collateralType string, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.NextClaimPeriodIDPrefix)
	store.Set([]byte(collateralType), sdk.Uint64ToBigEndian(id))
}

// IterateClaimPeriodIDKeysAndValues iterates over the claim period id (value) and collateral type (key) of each claim period id in the store and performs a callback function
func (k Keeper) IterateClaimPeriodIDKeysAndValues(ctx sdk.Context, cb func(collateralType string, id uint64) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.NextClaimPeriodIDPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		id := types.BytesToUint64(iterator.Value())
		collateralType := string(iterator.Key())
		if cb(collateralType, id) {
			break
		}
	}
}

// GetAllClaimPeriodIDPairs returns all collateralType:nextClaimPeriodID pairs in the store
func (k Keeper) GetAllClaimPeriodIDPairs(ctx sdk.Context) types.GenesisClaimPeriodIDs {
	ids := types.GenesisClaimPeriodIDs{}
	k.IterateClaimPeriodIDKeysAndValues(ctx, func(collateralType string, id uint64) (stop bool) {
		genID := types.GenesisClaimPeriodID{
			CollateralType: collateralType,
			ID:             id,
		}
		ids = append(ids, genID)
		return false
	})
	return ids
}

// GetClaimPeriod returns claim period in the store for the input ID and collateral type and a boolean for if it was found
func (k Keeper) GetClaimPeriod(ctx sdk.Context, id uint64, collateralType string) (types.ClaimPeriod, bool) {
	var cp types.ClaimPeriod
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	bz := store.Get(types.GetClaimPeriodPrefix(collateralType, id))
	if bz == nil {
		return types.ClaimPeriod{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &cp)
	return cp, true
}

// SetClaimPeriod sets the claim period in the store for the input ID and collateral type
func (k Keeper) SetClaimPeriod(ctx sdk.Context, cp types.ClaimPeriod) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(cp)
	store.Set(types.GetClaimPeriodPrefix(cp.CollateralType, cp.ID), bz)
}

// DeleteClaimPeriod deletes the claim period in the store for the input ID and collateral type
func (k Keeper) DeleteClaimPeriod(ctx sdk.Context, id uint64, collateralType string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	store.Delete(types.GetClaimPeriodPrefix(collateralType, id))
}

// IterateClaimPeriods iterates over all claim period objects in the store and preforms a callback function
func (k Keeper) IterateClaimPeriods(ctx sdk.Context, cb func(cp types.ClaimPeriod) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimPeriodKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cp types.ClaimPeriod
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &cp)
		if cb(cp) {
			break
		}
	}
}

// GetAllClaimPeriods returns all ClaimPeriod objects in the store
func (k Keeper) GetAllClaimPeriods(ctx sdk.Context) types.ClaimPeriods {
	cps := types.ClaimPeriods{}
	k.IterateClaimPeriods(ctx, func(cp types.ClaimPeriod) (stop bool) {
		cps = append(cps, cp)
		return false
	})
	return cps
}

// GetClaim returns the claim in the store corresponding the the input address collateral type and id and a boolean for if the claim was found
func (k Keeper) GetClaim(ctx sdk.Context, addr sdk.AccAddress, collateralType string, id uint64) (types.Claim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	bz := store.Get(types.GetClaimPrefix(addr, collateralType, id))
	if bz == nil {
		return types.Claim{}, false
	}
	var c types.Claim
	k.cdc.MustUnmarshalBinaryBare(bz, &c)
	return c, true
}

// SetClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetClaim(ctx sdk.Context, c types.Claim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(c)
	store.Set(types.GetClaimPrefix(c.Owner, c.CollateralType, c.ClaimPeriodID), bz)

}

// DeleteClaim deletes the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) DeleteClaim(ctx sdk.Context, owner sdk.AccAddress, collateralType string, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	store.Delete(types.GetClaimPrefix(owner, collateralType, id))
}

// IterateClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateClaims(ctx sdk.Context, cb func(c types.Claim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.Claim
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllClaims returns all Claim objects in the store
func (k Keeper) GetAllClaims(ctx sdk.Context) types.Claims {
	cs := types.Claims{}
	k.IterateClaims(ctx, func(c types.Claim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	b := store.Get([]byte{})
	if b == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &blockTime)
	return blockTime, true
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryBare(blockTime))
}
