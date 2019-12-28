package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/auction/types"
)

type Keeper struct {
	supplyKeeper  types.SupplyKeeper
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	// TODO codespace
}

// NewKeeper returns a new auction keeper.
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, supplyKeeper types.SupplyKeeper, paramstore subspace.Subspace) Keeper {
	return Keeper{
		supplyKeeper:  supplyKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
	}
}

// SetNextAuctionID stores an ID to be used for the next created auction
func (k Keeper) SetNextAuctionID(ctx sdk.Context, id types.ID) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextAuctionIDKey, id.Bytes())
}

// GetNextAuctionID reads the next available global ID from store
// TODO might be nicer to convert not found error to a panic, it's not an error that can be recovered from
func (k Keeper) GetNextAuctionID(ctx sdk.Context) (types.ID, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextAuctionIDKey)
	if bz == nil {
		//return 0, types.ErrInvalidGenesis(k.codespace, "initial auction ID hasn't been set") // TODO create error
		return 0, sdk.ErrInternal("initial auction ID hasn't been set")
	}
	return types.NewIDFromBytes(bz), nil
}

// incrementNextAuctionID increments the global ID in the store by 1
func (k Keeper) IncrementNextAuctionID(ctx sdk.Context) sdk.Error {
	id, err := k.GetNextAuctionID(ctx)
	if err != nil {
		return err
	}
	k.SetNextAuctionID(ctx, id+1)
	return nil
}

// storeNewAuction stores an auction, adding a new ID, and setting indexes
func (k Keeper) storeNewAuction(ctx sdk.Context, auction types.Auction) (types.ID, sdk.Error) {
	newAuctionID, err := k.GetNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}
	auction = auction.WithID(newAuctionID)

	k.SetAuction(ctx, auction)
	k.InsertIntoQueue(ctx, auction.GetEndTime(), auction.GetID())

	err = k.IncrementNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}
	return newAuctionID, nil
}

// TODO should get/set/delete be responsible for updating auctionByTime index?

// SetAuction puts the auction into the database and adds it to the queue
// it overwrites any pre-existing auction with same ID
func (k Keeper) SetAuction(ctx sdk.Context, auction types.Auction) {
	// remove the auction from the queue if it is already in there
	// existingAuction, found := k.GetAuction(ctx, auction.GetID())
	// if found {
	// 	k.removeFromQueue(ctx, existingAuction.GetEndTime(), existingAuction.GetID())
	// }

	// store auction
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auction)
	store.Set(types.GetAuctionKey(auction.GetID()), bz)

	// add to the queue
	//k.InsertIntoQueue(ctx, auction.GetEndTime(), auction.GetID())
}

// getAuction gets an auction from the store by auctionID
func (k Keeper) GetAuction(ctx sdk.Context, auctionID types.ID) (types.Auction, bool) {
	var auction types.Auction

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)
	bz := store.Get(types.GetAuctionKey(auctionID))
	if bz == nil {
		return auction, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auction)
	return auction, true
}

// DeleteAuction removes an auction from the store without any validation
func (k Keeper) DeleteAuction(ctx sdk.Context, auctionID types.ID) {
	// remove from queue
	//auction, found := k.GetAuction(ctx, auctionID)
	// if found {
	// 	k.removeFromQueue(ctx, auction.GetEndTime(), auctionID)
	// }

	// delete auction
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)
	store.Delete(types.GetAuctionKey(auctionID))
}

// Inserts a AuctionID into the queue at endTime
func (k Keeper) InsertIntoQueue(ctx sdk.Context, endTime time.Time, auctionID types.ID) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
	store.Set(types.GetAuctionByTimeKey(endTime, auctionID), auctionID.Bytes())
}

// removes an auctionID from the queue
func (k Keeper) RemoveFromQueue(ctx sdk.Context, endTime time.Time, auctionID types.ID) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
	store.Delete(types.GetAuctionByTimeKey(endTime, auctionID))
}

func (k Keeper) IterateAuctionsByTime(ctx sdk.Context, inclusiveCutoffTime time.Time, cb func(auctionID types.ID) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(sdk.FormatTimeBytes(inclusiveCutoffTime)), // include any keys with times equal to inclusiveCutoffTime
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		// TODO get the auction ID - either read from store, or extract from key
		auctionID := types.NewIDFromBytes(iterator.Value())

		if cb(auctionID) {
			break
		}
	}
}

// IterateAuctions provides an iterator over all stored auctions. For
// each auction, cb will be called. If the cb returns true, the iterator
// will close and stop.
func (k Keeper) IterateAuctions(ctx sdk.Context, cb func(auction types.Auction) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var auction types.Auction
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &auction)

		if cb(auction) {
			break
		}
	}
}
