package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/bep3/internal/types"
)

// Keeper of the bep3 store
type Keeper struct {
	supplyKeeper  types.SupplyKeeper
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	codespace     sdk.CodespaceType
}

// NewKeeper creates a bep3 keeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, supplyKeeper types.SupplyKeeper, paramstore subspace.Subspace) Keeper {
	if addr := supplyKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	keeper := Keeper{
		supplyKeeper:  supplyKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetNextKHTLTID stores an ID to be used for the next created KHTLT
func (k Keeper) SetNextKHTLTID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextKHTLTIDKey, types.Uint64ToBytes(id))
}

// GetNextKHTLTID reads the next available global ID from store
func (k Keeper) GetNextKHTLTID(ctx sdk.Context) (uint64, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextKHTLTIDKey)
	if bz == nil {
		return 0, types.ErrInvalidInitialKHTLTID(k.codespace)
	}
	return types.Uint64FromBytes(bz), nil
}

// IncrementNextKHTLTID increments the next HTLT ID in the store by 1.
func (k Keeper) IncrementNextKHTLTID(ctx sdk.Context) sdk.Error {
	id, err := k.GetNextKHTLTID(ctx)
	if err != nil {
		return err
	}
	k.SetNextKHTLTID(ctx, id+1)
	return nil
}

// StoreNewKHTLT stores an KHTLT, adding a new ID
func (k Keeper) StoreNewKHTLT(ctx sdk.Context, khtlt types.KHTLT) (uint64, sdk.Error) {
	newKHTLTID, err := k.GetNextKHTLTID(ctx)
	if err != nil {
		return 0, err
	}
	khtlt = khtlt.WithID(newKHTLTID)

	k.SetKHTLT(ctx, khtlt)

	err = k.IncrementNextKHTLTID(ctx)
	if err != nil {
		return 0, err
	}
	return newKHTLTID, nil
}

// SetKHTLT puts the KHTLT into the store, and updates any indexes.
func (k Keeper) SetKHTLT(ctx sdk.Context, khtlt types.KHTLT) {
	existingKHTLT, found := k.GetKHTLT(ctx, khtlt.ID)
	if found {
		k.removeFromByTimeIndex(ctx, existingKHTLT.EndTime, existingKHTLT.ID)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KHTLTKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(khtlt)
	store.Set(types.GetKHTLTKey(khtlt.ID), bz)

	k.InsertIntoByTimeIndex(ctx, khtlt.EndTime, khtlt.ID)
}

// GetKHTLT gets an htlt from the store.
func (k Keeper) GetKHTLT(ctx sdk.Context, htltID uint64) (types.KHTLT, bool) {
	var khtlt types.KHTLT

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KHTLTKeyPrefix)
	bz := store.Get(types.GetKHTLTKey(khtlt.ID))
	if bz == nil {
		return khtlt, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &khtlt)
	return khtlt, true
}

// DeleteKHTLT removes a KHTLT from the store, and any indexes.
func (k Keeper) DeleteKHTLT(ctx sdk.Context, khtltID uint64) {
	khtlt, found := k.GetKHTLT(ctx, khtltID)
	if found {
		k.removeFromByTimeIndex(ctx, khtlt.EndTime, khtltID)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KHTLTKeyPrefix)
	store.Delete(types.GetKHTLTKey(khtltID))
}

// InsertIntoByTimeIndex adds a KHTLT ID and end time into the byTime index.
func (k Keeper) InsertIntoByTimeIndex(ctx sdk.Context, endTime time.Time, khtltID uint64) { // TODO make private, and find way to make tests work
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KHTLTByTimeKeyPrefix)
	store.Set(types.GetKHTLTByTimeKey(endTime, khtltID), types.Uint64ToBytes(khtltID))
}

// removeFromByTimeIndex removes an KHTLT ID and end time from the byTime index.
func (k Keeper) removeFromByTimeIndex(ctx sdk.Context, endTime time.Time, khtltID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KHTLTByTimeKeyPrefix)
	store.Delete(types.GetKHTLTByTimeKey(endTime, khtltID))
}

// IterateKHTLTsByTime provides an iterator over KHTLTs ordered by KHTLT.EndTime.
// For each KHTLT cb will be callled. If cb returns true the iterator will close and stop.
func (k Keeper) IterateKHTLTsByTime(ctx sdk.Context, inclusiveCutoffTime time.Time, cb func(khtltID uint64) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KHTLTByTimeKeyPrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(sdk.FormatTimeBytes(inclusiveCutoffTime)), // include any keys with times equal to inclusiveCutoffTime
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		khtltID := types.Uint64FromBytes(iterator.Value())

		if cb(khtltID) {
			break
		}
	}
}

// IterateKHTLTs provides an iterator over all stored KHTLTs.
// For each KHTLT, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateKHTLTs(ctx sdk.Context, cb func(htlt types.KHTLT) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KHTLTKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var khtlt types.KHTLT
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &khtlt)

		if cb(khtlt) {
			break
		}
	}
}
