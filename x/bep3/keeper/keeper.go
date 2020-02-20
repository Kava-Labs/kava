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

// StoreNewAtomicSwap stores an AtomicSwap
func (k Keeper) StoreNewAtomicSwap(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	k.SetAtomicSwap(ctx, atomicSwap)
	k.InsertIntoByBlockIndex(ctx, atomicSwap)
}

// SetAtomicSwap puts the AtomicSwap into the store, and updates any indexes.
func (k Keeper) SetAtomicSwap(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(atomicSwap)
	store.Set(atomicSwap.SwapID, bz)
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

// DeleteAtomicSwap removes an AtomicSwap from the store, and any indexes.
func (k Keeper) DeleteAtomicSwap(ctx sdk.Context, swapID []byte) {
	// Remove AtomicSwap from byTime index
	atomicSwap, found := k.GetAtomicSwap(ctx, swapID)
	if found {
		k.removeFromByBlockIndex(ctx, atomicSwap.ExpirationBlock, atomicSwap.SwapID)
	}

	// Remove AtomicSwap from store
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

// InsertIntoByBlockIndex adds a swap ID and expiration time into the byTime index.
func (k Keeper) InsertIntoByBlockIndex(ctx sdk.Context, atomicSwap types.AtomicSwap) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapByBlockPrefix)
	store.Set(types.GetAtomicSwapByBlockKey(atomicSwap.ExpirationBlock, atomicSwap.SwapID), atomicSwap.SwapID)
}

// removeFromByBlockIndex removes a swap ID and expiration time from the byTime index.
func (k Keeper) removeFromByBlockIndex(ctx sdk.Context, expirationBlock uint64, swapID []byte) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapByBlockPrefix)
	store.Delete(types.GetAtomicSwapByBlockKey(expirationBlock, swapID))
}

// IterateAtomicSwapsByBlock provides an iterator over AtomicSwaps ordered by AtomicSwap expiration block
// For each AtomicSwap cb will be called. If cb returns true the iterator will close and stop.
func (k Keeper) IterateAtomicSwapsByBlock(ctx sdk.Context, inclusiveCutoffTime uint64, cb func(atomicSwapID []byte) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AtomicSwapByBlockPrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(types.Uint64ToBytes(inclusiveCutoffTime)), // end of range
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		atomicSwapID := iterator.Value()

		if cb(atomicSwapID) {
			break
		}
	}
}
