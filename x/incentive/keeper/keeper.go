package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// Keeper keeper for the incentive module
type Keeper struct {
	cdc           codec.Codec
	key           sdk.StoreKey
	paramSubspace types.ParamSubspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	cdpKeeper     types.CdpKeeper
	hardKeeper    types.HardKeeper
	stakingKeeper types.StakingKeeper
	swapKeeper    types.SwapKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc codec.Codec, key sdk.StoreKey, paramstore types.ParamSubspace, bk types.BankKeeper,
	cdpk types.CdpKeeper, hk types.HardKeeper, ak types.AccountKeeper, stk types.StakingKeeper,
	swpk types.SwapKeeper,
) Keeper {

	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		accountKeeper: ak,
		cdc:           cdc,
		key:           key,
		paramSubspace: paramstore,
		bankKeeper:    bk,
		cdpKeeper:     cdpk,
		hardKeeper:    hk,
		stakingKeeper: stk,
		swapKeeper:    swpk,
	}
}

// GetUSDXMintingClaim returns the claim in the store corresponding the the input address collateral type and id and a boolean for if the claim was found
func (k Keeper) GetUSDXMintingClaim(ctx sdk.Context, addr sdk.AccAddress) (types.USDXMintingClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.USDXMintingClaim{}, false
	}
	var c types.USDXMintingClaim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetUSDXMintingClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetUSDXMintingClaim(ctx sdk.Context, c types.USDXMintingClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)

}

// DeleteUSDXMintingClaim deletes the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) DeleteUSDXMintingClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	store.Delete(owner)
}

// IterateUSDXMintingClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateUSDXMintingClaims(ctx sdk.Context, cb func(c types.USDXMintingClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.USDXMintingClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.USDXMintingClaim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllUSDXMintingClaims returns all Claim objects in the store
func (k Keeper) GetAllUSDXMintingClaims(ctx sdk.Context) types.USDXMintingClaims {
	cs := types.USDXMintingClaims{}
	k.IterateUSDXMintingClaims(ctx, func(c types.USDXMintingClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetUSDXMintingRewardFactor returns the current reward factor for an individual collateral type
func (k Keeper) GetUSDXMintingRewardFactor(ctx sdk.Context, ctype string) (sdk.Dec, bool) {
	indexes, found := k.GetRewardIndexes(ctx, types.USDXMinting, ctype)
	if !found {
		return sdk.ZeroDec(), false
	}
	factor, found := indexes.Get(types.USDXMintingRewardDenom)
	if !found || len(indexes) != 1 {
		panic(fmt.Sprintf("USDX Minting reward factors must only have denom %s", types.USDXMintingRewardDenom))
	}
	return factor, true
}

// SetUSDXMintingRewardFactor sets the current reward factor for an individual collateral type
func (k Keeper) SetUSDXMintingRewardFactor(ctx sdk.Context, ctype string, factor sdk.Dec) {
	indexes := types.RewardIndexes{types.NewRewardIndex(types.USDXMintingRewardDenom, factor)}
	k.SetRewardIndexes(ctx, types.USDXMinting, ctype, indexes)
}

// IterateUSDXMintingRewardFactors iterates over all USDX Minting reward factor objects in the store and preforms a callback function
func (k Keeper) IterateUSDXMintingRewardFactors(ctx sdk.Context, cb func(denom string, factor sdk.Dec) (stop bool)) {
	k.IterateRewardIndexes(ctx, types.USDXMinting, func(denom string, indexes types.RewardIndexes) bool {
		factor, found := indexes.Get(types.USDXMintingRewardDenom)
		if !found || len(indexes) != 1 {
			panic(fmt.Sprintf("USDX Minting reward factors must only have denom %s", types.USDXMintingRewardDenom))
		}
		return cb(denom, factor)
	})
}

// GetHardLiquidityProviderClaim returns the claim in the store corresponding the the input address collateral type and id and a boolean for if the claim was found
func (k Keeper) GetHardLiquidityProviderClaim(ctx sdk.Context, addr sdk.AccAddress) (types.HardLiquidityProviderClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.HardLiquidityProviderClaim{}, false
	}
	var c types.HardLiquidityProviderClaim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetHardLiquidityProviderClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetHardLiquidityProviderClaim(ctx sdk.Context, c types.HardLiquidityProviderClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)
}

// DeleteHardLiquidityProviderClaim deletes the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) DeleteHardLiquidityProviderClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	store.Delete(owner)
}

// IterateHardLiquidityProviderClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateHardLiquidityProviderClaims(ctx sdk.Context, cb func(c types.HardLiquidityProviderClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HardLiquidityClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.HardLiquidityProviderClaim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllHardLiquidityProviderClaims returns all Claim objects in the store
func (k Keeper) GetAllHardLiquidityProviderClaims(ctx sdk.Context) types.HardLiquidityProviderClaims {
	cs := types.HardLiquidityProviderClaims{}
	k.IterateHardLiquidityProviderClaims(ctx, func(c types.HardLiquidityProviderClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetDelegatorClaim returns the claim in the store corresponding the the input address collateral type and id and a boolean for if the claim was found
func (k Keeper) GetDelegatorClaim(ctx sdk.Context, addr sdk.AccAddress) (types.DelegatorClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.DelegatorClaim{}, false
	}
	var c types.DelegatorClaim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetDelegatorClaim sets the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) SetDelegatorClaim(ctx sdk.Context, c types.DelegatorClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)
}

// DeleteDelegatorClaim deletes the claim in the store corresponding to the input address, collateral type, and id
func (k Keeper) DeleteDelegatorClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	store.Delete(owner)
}

// IterateDelegatorClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateDelegatorClaims(ctx sdk.Context, cb func(c types.DelegatorClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegatorClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.DelegatorClaim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllDelegatorClaims returns all DelegatorClaim objects in the store
func (k Keeper) GetAllDelegatorClaims(ctx sdk.Context) types.DelegatorClaims {
	cs := types.DelegatorClaims{}
	k.IterateDelegatorClaims(ctx, func(c types.DelegatorClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetSwapClaim returns the claim in the store corresponding the the input address.
func (k Keeper) GetSwapClaim(ctx sdk.Context, addr sdk.AccAddress) (types.SwapClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.SwapClaim{}, false
	}
	var c types.SwapClaim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetSwapClaim sets the claim in the store corresponding to the input address.
func (k Keeper) SetSwapClaim(ctx sdk.Context, c types.SwapClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)
}

// DeleteSwapClaim deletes the claim in the store corresponding to the input address.
func (k Keeper) DeleteSwapClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	store.Delete(owner)
}

// IterateSwapClaims iterates over all claim  objects in the store and preforms a callback function
func (k Keeper) IterateSwapClaims(ctx sdk.Context, cb func(c types.SwapClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SwapClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.SwapClaim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllSwapClaims returns all Claim objects in the store
func (k Keeper) GetAllSwapClaims(ctx sdk.Context) types.SwapClaims {
	cs := types.SwapClaims{}
	k.IterateSwapClaims(ctx, func(c types.SwapClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// GetSavingsClaim returns the claim in the store corresponding the the input address.
func (k Keeper) GetSavingsClaim(ctx sdk.Context, addr sdk.AccAddress) (types.SavingsClaim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	bz := store.Get(addr)
	if bz == nil {
		return types.SavingsClaim{}, false
	}
	var c types.SavingsClaim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetSavingsClaim sets the claim in the store corresponding to the input address.
func (k Keeper) SetSavingsClaim(ctx sdk.Context, c types.SavingsClaim) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)
}

// DeleteSavingsClaim deletes the claim in the store corresponding to the input address.
func (k Keeper) DeleteSavingsClaim(ctx sdk.Context, owner sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	store.Delete(owner)
}

// IterateSavingsClaims iterates over all savings claim objects in the store and preforms a callback function
func (k Keeper) IterateSavingsClaims(ctx sdk.Context, cb func(c types.SavingsClaim) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.SavingsClaimKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.SavingsClaim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllSavingsClaims returns all savings claim objects in the store
func (k Keeper) GetAllSavingsClaims(ctx sdk.Context) types.SavingsClaims {
	cs := types.SavingsClaims{}
	k.IterateSavingsClaims(ctx, func(c types.SavingsClaim) (stop bool) {
		cs = append(cs, c)
		return false
	})
	return cs
}

// SetRewardIndexes stores the global reward indexes that track total rewards to a source.
func (k Keeper) SetRewardIndexes(ctx sdk.Context, sourceID types.SourceID, sourceSubID string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardIndexesKeyPrefix)
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set(types.GetFullSourceIDKey(sourceID, sourceSubID), bz)
}

// GetRewardIndexes fetches the global reward indexes that track total rewards to a source
func (k Keeper) GetRewardIndexes(ctx sdk.Context, sourceID types.SourceID, sourceSubID string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardIndexesKeyPrefix)
	bz := store.Get(types.GetFullSourceIDKey(sourceID, sourceSubID))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)
	return proto.RewardIndexes, true
}

// IterateRewardIndexes iterates over all reward index objects in the store and preforms a callback function
func (k Keeper) IterateRewardIndexes(ctx sdk.Context, sourceID types.SourceID, cb func(sourceSubID string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.RewardIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, types.GetSourceIDKey(sourceID))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		k.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()), proto.RewardIndexes) {
			break
		}
	}
}

// GetLastAccrual fetches the last time rewards were accrued for a source.
func (k Keeper) GetLastAccrual(ctx sdk.Context, sourceID types.SourceID, sourceSubID string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.LastAccrualKeyPrefix)
	b := store.Get(types.GetFullSourceIDKey(sourceID, sourceSubID))
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

// SetLastAccrual stores the last time rewards were accrued for a swap pool.
func (k Keeper) SetLastAccrual(ctx sdk.Context, sourceID types.SourceID, sourceSubID string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.LastAccrualKeyPrefix)
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set(types.GetFullSourceIDKey(sourceID, sourceSubID), bz)
}

// IterateLastAccruals steps through previous accrual times for each source.
func (k Keeper) IterateLastAccruals(ctx sdk.Context, sourceID types.SourceID, cb func(string, time.Time) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.LastAccrualKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, types.GetSourceIDKey(sourceID))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		sourceSubID := string(iterator.Key())
		var accrualTime time.Time
		if err := accrualTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}
		if cb(sourceSubID, accrualTime) {
			break
		}
	}
}
