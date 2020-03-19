package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the bep3 store
type Keeper struct {
	key           sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	supplyKeeper  types.SupplyKeeper
	codespace     sdk.CodespaceType
}

// NewKeeper creates a bep3 keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, sk types.SupplyKeeper, paramstore subspace.Subspace, codespace sdk.CodespaceType) Keeper {
	if addr := sk.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	keeper := Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
		supplyKeeper:  sk,
		codespace:     codespace,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ------------------------------------------
//				Atomic Swaps
// ------------------------------------------

// SetAtomicSwap puts the AtomicSwap into the store, and updates any indexes.
func (k Keeper) SetAtomicSwap(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(atomicSwap)
	store.Set(atomicSwap.GetSwapID(), bz)
}

// GetAtomicSwap gets an AtomicSwap from the store.
func (k Keeper) GetAtomicSwap(ctx sdk.Context, swapID []byte) (types.AtomicSwap, bool) {
	var atomicSwap types.AtomicSwap

	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapKeyPrefix)
	bz := store.Get(swapID)
	if bz == nil {
		return atomicSwap, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &atomicSwap)
	return atomicSwap, true
}

// RemoveAtomicSwap removes an AtomicSwap from the AtomicSwapKeyPrefix.
func (k Keeper) RemoveAtomicSwap(ctx sdk.Context, swapID []byte) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapKeyPrefix)
	store.Delete(swapID)
}

// IterateAtomicSwaps provides an iterator over all stored AtomicSwaps.
// For each AtomicSwap, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateAtomicSwaps(ctx sdk.Context, cb func(atomicSwap types.AtomicSwap) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.AtomicSwapKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var atomicSwap types.AtomicSwap
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &atomicSwap)

		if cb(atomicSwap) {
			break
		}
	}
}

// GetAllAtomicSwaps returns all AtomicSwaps from the store
func (k Keeper) GetAllAtomicSwaps(ctx sdk.Context) (atomicSwaps types.AtomicSwaps) {
	k.IterateAtomicSwaps(ctx, func(atomicSwap types.AtomicSwap) bool {
		atomicSwaps = append(atomicSwaps, atomicSwap)
		return false
	})
	return
}

// ------------------------------------------
//			Atomic Swap Block Index
// ------------------------------------------

// InsertIntoByBlockIndex adds a swap ID and expiration time into the byBlock index.
func (k Keeper) InsertIntoByBlockIndex(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapByBlockPrefix)
	store.Set(types.GetAtomicSwapByHeightKey(atomicSwap.ExpireHeight, atomicSwap.GetSwapID()), atomicSwap.GetSwapID())
}

// RemoveFromByBlockIndex removes an AtomicSwap from the byBlock index.
func (k Keeper) RemoveFromByBlockIndex(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapByBlockPrefix)
	store.Delete(types.GetAtomicSwapByHeightKey(atomicSwap.ExpireHeight, atomicSwap.GetSwapID()))
}

// IterateAtomicSwapsByBlock provides an iterator over AtomicSwaps ordered by AtomicSwap expiration block
// For each AtomicSwap cb will be called. If cb returns true the iterator will close and stop.
func (k Keeper) IterateAtomicSwapsByBlock(ctx sdk.Context, inclusiveCutoffTime uint64, cb func(swapID []byte) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapByBlockPrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(types.Uint64ToBytes(inclusiveCutoffTime)), // end of range
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		id := iterator.Value()

		if cb(id) {
			break
		}
	}
}

// ------------------------------------------
//		Atomic Swap Longterm Storage Index
// TODO: make block deletion variable a param genesis value
// ------------------------------------------

// InsertIntoLongtermStorage adds a swap ID and deletion time into the longterm storage index.
// Completed swaps are stored for 1 week.
func (k Keeper) InsertIntoLongtermStorage(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapLongtermStoragePrefix)
	// Assuming block time of 7 seconds, there are 86,400 blocks in a week
	store.Set(types.GetAtomicSwapByHeightKey(atomicSwap.ClosedBlock+int64(86400), atomicSwap.GetSwapID()), atomicSwap.GetSwapID())
}

// RemoveFromLongtermStorage removes a swap from the into the longterm storage index
func (k Keeper) RemoveFromLongtermStorage(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapLongtermStoragePrefix)
	store.Delete(types.GetAtomicSwapByHeightKey(atomicSwap.ClosedBlock+int64(86400), atomicSwap.GetSwapID()))
}

// IterateAtomicSwapsLongtermStorage provides an iterator over AtomicSwaps ordered by deletion height.
// For each AtomicSwap cb will be called. If cb returns true the iterator will close and stop.
func (k Keeper) IterateAtomicSwapsLongtermStorage(ctx sdk.Context, inclusiveCutoffTime uint64, cb func(swapID []byte) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapLongtermStoragePrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(types.Uint64ToBytes(inclusiveCutoffTime)), // end of range
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		id := iterator.Value()

		if cb(id) {
			break
		}
	}
}

// ------------------------------------------
//				Asset Supplies
// ------------------------------------------

// SetAssetSupply updates an asset's current active supply
func (k Keeper) SetAssetSupply(ctx sdk.Context, asset sdk.Coin, denom []byte) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AssetSupplyKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(asset)
	store.Set(denom, bz)
}

// GetAssetSupply gets an asset's current supply from the store.
func (k Keeper) GetAssetSupply(ctx sdk.Context, denom []byte) (sdk.Coin, bool) {
	var asset sdk.Coin

	store := prefix.NewStore(ctx.KVStore(k.key), types.AssetSupplyKeyPrefix)
	bz := store.Get(denom)
	if bz == nil {
		return sdk.Coin{}, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &asset)
	return asset, true
}

// IterateAssetSupplies provides an iterator over all stored AssetSupplies.
// For each AssetSupply, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateAssetSupplies(ctx sdk.Context, cb func(asset sdk.Coin) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.AssetSupplyKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var asset sdk.Coin
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &asset)

		if cb(asset) {
			break
		}
	}
}

// GetAllAssetSupplies returns all asset supplies from the store as an array of sdk.Coin
func (k Keeper) GetAllAssetSupplies(ctx sdk.Context) (assets []sdk.Coin) {
	k.IterateAssetSupplies(ctx, func(asset sdk.Coin) bool {
		assets = append(assets, asset)
		return false
	})
	return
}

// ------------------------------------------
//				In Swap Supply
// ------------------------------------------

// SetInSwapSupply sets an asset's total supply currently in swaps from the store.
func (k Keeper) SetInSwapSupply(ctx sdk.Context, asset sdk.Coin, denom []byte) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.InSwapSupplyKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(asset)
	store.Set(denom, bz)
}

// GetInSwapSupply gets an asset's total supply currently in swaps from the store.
func (k Keeper) GetInSwapSupply(ctx sdk.Context, denom []byte) (sdk.Coin, bool) {
	var asset sdk.Coin

	store := prefix.NewStore(ctx.KVStore(k.key), types.InSwapSupplyKeyPrefix)
	bz := store.Get(denom)
	if bz == nil {
		return sdk.Coin{}, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &asset)
	return asset, true
}
