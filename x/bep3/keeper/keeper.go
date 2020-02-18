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

// StoreNewHTLT stores an HTLT
func (k Keeper) StoreNewHTLT(ctx sdk.Context, htlt types.HTLT) (string, error) {
	swapID, err := types.CalculateSwapID(htlt.RandomNumberHash, htlt.From, htlt.SenderOtherChain)
	if err != nil {
		return "", err
	}
	k.SetHTLT(ctx, htlt, swapID)
	return types.BytesToHexEncodedString(swapID), nil
}

// SetHTLT puts the HTLT into the store, and updates any indexes.
func (k Keeper) SetHTLT(ctx sdk.Context, htlt types.HTLT, swapID []byte) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HTLTKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(htlt)
	store.Set(swapID, bz)
}

// GetHTLT gets an htlt from the store.
func (k Keeper) GetHTLT(ctx sdk.Context, swapID []byte) (types.HTLT, bool) {
	var htlt types.HTLT

	store := prefix.NewStore(ctx.KVStore(k.key), types.HTLTKeyPrefix)
	bz := store.Get(swapID)
	if bz == nil {
		return htlt, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &htlt)
	return htlt, true
}

// DeleteHTLT removes a HTLT from the store, and any indexes.
func (k Keeper) DeleteHTLT(ctx sdk.Context, swapID []byte) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HTLTKeyPrefix)
	store.Delete(swapID)
}

// IterateHTLTs provides an iterator over all stored HTLTs.
// For each HTLT, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateHTLTs(ctx sdk.Context, cb func(htlt types.HTLT) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.HTLTKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var htlt types.HTLT
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &htlt)

		if cb(htlt) {
			break
		}
	}
}

// InsertIntoByTimeIndex adds a htlt ID and expiration time into the byTime index.
func (k Keeper) InsertIntoByTimeIndex(ctx sdk.Context, expirationTime uint64, htltID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HTLTByTimeKeyPrefix)
	store.Set(types.GetHTLTByTimeKey(expirationTime, htltID), types.Uint64ToBytes(htltID))
}

// removeFromByTimeIndex removes a htlt ID and expiration time from the byTime index.
func (k Keeper) removeFromByTimeIndex(ctx sdk.Context, expirationTime uint64, htltID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HTLTByTimeKeyPrefix)
	store.Delete(types.GetHTLTByTimeKey(expirationTime, htltID))
}

// IterateHTLTsByTime provides an iterator over HTLTs ordered by HTLT expiration time,
// where expiration time = HTLT.timestamp + (HTLT.heightspan*millisecondsPerBlock).
// TODO: HTLT.timestamp does not always equal the time at which the HTLT was submitted
// For each HTLT cb will be called. If cb returns true the iterator will close and stop.
func (k Keeper) IterateHTLTsByTime(ctx sdk.Context, inclusiveCutoffTime uint64, cb func(htltID uint64) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.HTLTByTimeKeyPrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(types.Uint64ToBytes(inclusiveCutoffTime)),
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		htltID := types.Uint64FromBytes(iterator.Value())

		if cb(htltID) {
			break
		}
	}
}
