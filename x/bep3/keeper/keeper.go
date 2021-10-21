package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/bep3/types"
)

// Keeper of the bep3 store
type Keeper struct {
	key           sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	supplyKeeper  types.SupplyKeeper
	accountKeeper types.AccountKeeper
	Maccs         map[string]bool
}

// NewKeeper creates a bep3 keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, sk types.SupplyKeeper, ak types.AccountKeeper,
	paramstore subspace.Subspace, maccs map[string]bool) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	keeper := Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore,
		supplyKeeper:  sk,
		accountKeeper: ak,
		Maccs:         maccs,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// EnsureModuleAccountPermissions syncs the bep3 module account's permissions with those in the supply keeper.
func (k Keeper) EnsureModuleAccountPermissions(ctx sdk.Context) error {
	maccI := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
	macc, ok := maccI.(*supply.ModuleAccount)
	if !ok {
		return fmt.Errorf("expected %s account to be a module account type", types.ModuleName)
	}
	_, perms := k.supplyKeeper.GetModuleAddressAndPermissions(types.ModuleName)
	macc.Permissions = perms
	k.supplyKeeper.SetModuleAccount(ctx, macc)
	return nil
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
		sdk.PrefixEndBytes(sdk.Uint64ToBigEndian(inclusiveCutoffTime)), // end of range
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
// ------------------------------------------

// InsertIntoLongtermStorage adds a swap ID and deletion time into the longterm storage index.
// Completed swaps are stored for 1 week.
func (k Keeper) InsertIntoLongtermStorage(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapLongtermStoragePrefix)
	deletionHeight := uint64(atomicSwap.ClosedBlock) + types.DefaultLongtermStorageDuration
	store.Set(types.GetAtomicSwapByHeightKey(deletionHeight, atomicSwap.GetSwapID()), atomicSwap.GetSwapID())
}

// RemoveFromLongtermStorage removes a swap from the into the longterm storage index
func (k Keeper) RemoveFromLongtermStorage(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapLongtermStoragePrefix)
	deletionHeight := uint64(atomicSwap.ClosedBlock) + types.DefaultLongtermStorageDuration
	store.Delete(types.GetAtomicSwapByHeightKey(deletionHeight, atomicSwap.GetSwapID()))
}

// IterateAtomicSwapsLongtermStorage provides an iterator over AtomicSwaps ordered by deletion height.
// For each AtomicSwap cb will be called. If cb returns true the iterator will close and stop.
func (k Keeper) IterateAtomicSwapsLongtermStorage(ctx sdk.Context, inclusiveCutoffTime uint64,
	cb func(swapID []byte) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapLongtermStoragePrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(sdk.Uint64ToBigEndian(inclusiveCutoffTime)), // end of range
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

// GetAssetSupply gets an asset's current supply from the store.
func (k Keeper) GetAssetSupply(ctx sdk.Context, denom string) (types.AssetSupply, bool) {
	var assetSupply types.AssetSupply
	store := prefix.NewStore(ctx.KVStore(k.key), types.AssetSupplyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.AssetSupply{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &assetSupply)
	return assetSupply, true
}

// SetAssetSupply updates an asset's supply
func (k Keeper) SetAssetSupply(ctx sdk.Context, supply types.AssetSupply, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AssetSupplyPrefix)
	store.Set([]byte(denom), k.cdc.MustMarshalBinaryBare(supply))
}

// IterateAssetSupplies provides an iterator over all stored AssetSupplies.
func (k Keeper) IterateAssetSupplies(ctx sdk.Context, cb func(supply types.AssetSupply) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.AssetSupplyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var supply types.AssetSupply
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &supply)

		if cb(supply) {
			break
		}
	}
}

// GetAllAssetSupplies returns all asset supplies from the store
func (k Keeper) GetAllAssetSupplies(ctx sdk.Context) (supplies types.AssetSupplies) {
	k.IterateAssetSupplies(ctx, func(supply types.AssetSupply) bool {
		supplies = append(supplies, supply)
		return false
	})
	return
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	b := store.Get([]byte{})
	if b == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &blockTime)
	return blockTime, true
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryLengthPrefixed(blockTime))
}
