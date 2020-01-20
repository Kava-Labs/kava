package keeper

import (
	"fmt"

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

// SetNextHTLTID stores an ID to be used for the next created auction
func (k Keeper) SetNextHTLTID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextHTLTIDKey, types.Uint64ToBytes(id))
}

// GetNextHTLTID reads the next available global ID from store
func (k Keeper) GetNextHTLTID(ctx sdk.Context) (uint64, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextHTLTIDKey)
	if bz == nil {
		return 0, types.ErrInvalidInitialHTLTID(k.codespace)
	}
	return types.Uint64FromBytes(bz), nil
}

// IncrementNextHTLTID increments the next HTLT ID in the store by 1.
func (k Keeper) IncrementNextHTLTID(ctx sdk.Context) sdk.Error {
	id, err := k.GetNextHTLTID(ctx)
	if err != nil {
		return err
	}
	k.SetNextHTLTID(ctx, id+1)
	return nil
}

// StoreNewHTLT stores an HTLT, adding a new ID
func (k Keeper) StoreNewHTLT(ctx sdk.Context, htlt types.KavaHTLT) (uint64, sdk.Error) {
	newHTLTID, err := k.GetNextHTLTID(ctx)
	if err != nil {
		return 0, err
	}
	htlt = htlt.WithID(newHTLTID)

	k.SetHTLT(ctx, htlt)

	err = k.IncrementNextHTLTID(ctx)
	if err != nil {
		return 0, err
	}
	return newHTLTID, nil
}

// SetHTLT puts the HTLT into the store, and updates any indexes.
func (k Keeper) SetHTLT(ctx sdk.Context, htlt types.KavaHTLT) {
	// TODO: remove the HTLT from the byTime index if it is already in there
	// existingHTLT, found := k.GetHTLT(ctx, htlt.ID)

	// if found {
	// 	k.removeFromByTimeIndex(ctx, existingHTLT.GetEndTime(), existingHTLT.GetID())
	// }

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.HTLTKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(htlt)
	store.Set(types.GetHTLTKey(htlt.ID), bz)

	// k.InsertIntoByTimeIndex(ctx, htlt.GetEndTime(), htlt.ID)
}

// GetHTLT gets an htlt from the store.
func (k Keeper) GetHTLT(ctx sdk.Context, htltID uint64) (types.KavaHTLT, bool) {
	var htlt types.KavaHTLT

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.HTLTKeyPrefix)
	bz := store.Get(types.GetHTLTKey(htltID))
	if bz == nil {
		return htlt, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &htlt)
	return htlt, true
}

// // DeleteAuction removes an auction from the store, and any indexes.
// func (k Keeper) DeleteAuction(ctx sdk.Context, auctionID uint64) {
// 	auction, found := k.GetAuction(ctx, auctionID)
// 	if found {
// 		k.removeFromByTimeIndex(ctx, auction.GetEndTime(), auctionID)
// 	}

// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)
// 	store.Delete(types.GetAuctionKey(auctionID))
// }

// // InsertIntoByTimeIndex adds an auction ID and end time into the byTime index.
// func (k Keeper) InsertIntoByTimeIndex(ctx sdk.Context, endTime time.Time, auctionID uint64) { // TODO make private, and find way to make tests work
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
// 	store.Set(types.GetAuctionByTimeKey(endTime, auctionID), types.Uint64ToBytes(auctionID))
// }

// // removeFromByTimeIndex removes an auction ID and end time from the byTime index.
// func (k Keeper) removeFromByTimeIndex(ctx sdk.Context, endTime time.Time, auctionID uint64) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
// 	store.Delete(types.GetAuctionByTimeKey(endTime, auctionID))
// }

// // IterateAuctionByTime provides an iterator over auctions ordered by auction.EndTime.
// // For each auction cb will be callled. If cb returns true the iterator will close and stop.
// func (k Keeper) IterateAuctionsByTime(ctx sdk.Context, inclusiveCutoffTime time.Time, cb func(auctionID uint64) (stop bool)) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
// 	iterator := store.Iterator(
// 		nil, // start at the very start of the prefix store
// 		sdk.PrefixEndBytes(sdk.FormatTimeBytes(inclusiveCutoffTime)), // include any keys with times equal to inclusiveCutoffTime
// 	)

// 	defer iterator.Close()
// 	for ; iterator.Valid(); iterator.Next() {

// 		auctionID := types.Uint64FromBytes(iterator.Value())

// 		if cb(auctionID) {
// 			break
// 		}
// 	}
// }

// IterateHTLTs provides an iterator over all stored HTLTs.
// For each HTLT, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateHTLTs(ctx sdk.Context, cb func(htlt types.KavaHTLT) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.HTLTKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var htlt types.KavaHTLT
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &htlt)

		if cb(htlt) {
			break
		}
	}
}
