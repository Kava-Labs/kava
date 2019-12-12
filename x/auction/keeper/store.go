package keeper

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction/types"
)

// set an auction in the store, adding a new ID, and setting indexes
func (k Keeper) storeNewAuction(ctx sdk.Context, auction types.Auction) (types.ID, sdk.Error) {
	// get ID
	newAuctionID, err := k.getNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}
	// set ID
	auction.SetID(newAuctionID)

	// store auction
	k.SetAuction(ctx, auction)
	k.incrementNextAuctionID(ctx)
	return newAuctionID, nil
}

// getNextAuctionID gets the next available global AuctionID
func (k Keeper) getNextAuctionID(ctx sdk.Context) (types.ID, sdk.Error) {
	// get next ID from store
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		// if not found, set the id at 0
		bz = k.cdc.MustMarshalBinaryLengthPrefixed(types.ID(0))
		store.Set(k.getNextAuctionIDKey(), bz)
		// TODO Set auction ID in genesis
		//return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	var auctionID types.ID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auctionID)
	return auctionID, nil
}

// incrementNextAuctionID increments the global ID in the store by 1
func (k Keeper) incrementNextAuctionID(ctx sdk.Context) sdk.Error {
	// get next ID from store
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		panic("initial auctionID never set in genesis")
		//return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set") // TODO
	}
	var auctionID types.ID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auctionID)

	// increment the stored next ID
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(auctionID + 1)
	store.Set(k.getNextAuctionIDKey(), bz)

	return nil
}

// SetAuction puts the auction into the database and adds it to the queue
// it overwrites any pre-existing auction with same ID
func (k Keeper) SetAuction(ctx sdk.Context, auction types.Auction) {
	// remove the auction from the queue if it is already in there
	existingAuction, found := k.GetAuction(ctx, auction.GetID())
	if found {
		k.removeFromQueue(ctx, existingAuction.GetEndTime(), existingAuction.GetID())
	}

	// store auction
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auction)
	store.Set(k.getAuctionKey(auction.GetID()), bz)

	// add to the queue
	k.InsertIntoQueue(ctx, auction.GetEndTime(), auction.GetID())
}

// getAuction gets an auction from the store by auctionID
func (k Keeper) GetAuction(ctx sdk.Context, auctionID types.ID) (types.Auction, bool) {
	var auction types.Auction

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getAuctionKey(auctionID))
	if bz == nil {
		return auction, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auction)
	return auction, true
}

// DeleteAuction removes an auction from the store without any validation
func (k Keeper) DeleteAuction(ctx sdk.Context, auctionID types.ID) {
	// remove from queue
	auction, found := k.GetAuction(ctx, auctionID)
	if found {
		k.removeFromQueue(ctx, auction.GetEndTime(), auctionID)
	}

	// delete auction
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getAuctionKey(auctionID))
}

// ---------- Queue and key methods ----------
// These are lower level function used by the store methods above.

func (k Keeper) getNextAuctionIDKey() []byte {
	return []byte("nextAuctionID")
}
func (k Keeper) getAuctionKey(auctionID types.ID) []byte {
	return []byte(fmt.Sprintf("auctions:%d", auctionID))
}

// Inserts a AuctionID into the queue at endTime
func (k Keeper) InsertIntoQueue(ctx sdk.Context, endTime time.Time, auctionID types.ID) {
	// get the store
	store := ctx.KVStore(k.storeKey)
	// marshal thing to be inserted
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auctionID)
	// store it
	store.Set(
		getQueueElementKey(endTime, auctionID),
		bz,
	)
}

// removes an auctionID from the queue
func (k Keeper) removeFromQueue(ctx sdk.Context, endTime time.Time, auctionID types.ID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getQueueElementKey(endTime, auctionID))
}

// Returns an iterator for all the auctions in the queue that expire by endTime
func (k Keeper) GetQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator { // TODO rename to "getAuctionsByExpiry" ?
	// get store
	store := ctx.KVStore(k.storeKey)
	// get an interator
	return store.Iterator(
		queueKeyPrefix, // start key
		sdk.PrefixEndBytes(getQueueElementKeyPrefix(endTime)), // end key (apparently exclusive but tests suggested otherwise)
	)
}

// GetAuctionIterator returns an iterator over all auctions in the store
func (k Keeper) GetAuctionIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, nil)
}

var queueKeyPrefix = []byte("queue")
var keyDelimiter = []byte(":")

// Returns half a key for an auctionID in the queue, it missed the id off the end
func getQueueElementKeyPrefix(endTime time.Time) []byte {
	return bytes.Join([][]byte{
		queueKeyPrefix,
		sdk.Uint64ToBigEndian(uint64(endTime)), // TODO check this gives correct ordering
	}, keyDelimiter)
}

// Returns the key for an auctionID in the queue
func getQueueElementKey(endTime time.Time, auctionID types.ID) []byte {
	return bytes.Join([][]byte{
		queueKeyPrefix,
		sdk.Uint64ToBigEndian(uint64(endTime)), // TODO check this gives correct ordering
		sdk.Uint64ToBigEndian(uint64(auctionID)),
	}, keyDelimiter)
}

// GetAuctionID returns the id from an input Auction
func (k Keeper) DecodeAuctionID(ctx sdk.Context, idBytes []byte) types.ID {
	var auctionID types.ID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(idBytes, &auctionID)
	return auctionID
}

func (k Keeper) DecodeAuction(ctx sdk.Context, auctionBytes []byte) types.Auction {
	var auction types.Auction
	k.cdc.MustUnmarshalBinaryBare(auctionBytes, &auction)
	return auction
}
