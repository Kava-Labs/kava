package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/kava-labs/kava/x/auction/types"
)

type Keeper struct {
	bankKeeper    types.BankKeeper
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	// TODO codespace
}

// NewKeeper returns a new auction keeper.
func NewKeeper(cdc *codec.Codec, bankKeeper types.BankKeeper, storeKey sdk.StoreKey, paramstore subspace.Subspace) Keeper {
	return Keeper{
		bankKeeper:    bankKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
	}
}

// TODO these 3 start functions be combined or abstracted away?

// StartForwardAuction starts a normal auction. Known as flap in maker.
func (k Keeper) StartForwardAuction(ctx sdk.Context, seller sdk.AccAddress, lot sdk.Coin, initialBid sdk.Coin) (types.ID, sdk.Error) {
	// create auction
	auction, initiatorOutput := types.NewForwardAuction(seller, lot, initialBid, types.EndTime(ctx.BlockHeight())+types.DefaultMaxAuctionDuration)
	// start the auction
	auctionID, err := k.startAuction(ctx, &auction, initiatorOutput)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

// StartReverseAuction starts an auction where sellers compete by offering decreasing prices. Known as flop in maker.
func (k Keeper) StartReverseAuction(ctx sdk.Context, buyer sdk.AccAddress, bid sdk.Coin, initialLot sdk.Coin) (types.ID, sdk.Error) {
	// create auction
	auction, initiatorOutput := types.NewReverseAuction(buyer, bid, initialLot, types.EndTime(ctx.BlockHeight())+types.DefaultMaxAuctionDuration)
	// start the auction
	auctionID, err := k.startAuction(ctx, &auction, initiatorOutput)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

// StartForwardReverseAuction starts an auction where bidders bid up to a maxBid, then switch to bidding down on price. Known as flip in maker.
func (k Keeper) StartForwardReverseAuction(ctx sdk.Context, seller sdk.AccAddress, lot sdk.Coin, maxBid sdk.Coin, otherPerson sdk.AccAddress) (types.ID, sdk.Error) {
	// create auction
	initialBid := sdk.NewInt64Coin(maxBid.Denom, 0) // set the bidding coin denomination from the specified max bid
	auction, initiatorOutput := types.NewForwardReverseAuction(seller, lot, initialBid, types.EndTime(ctx.BlockHeight())+types.DefaultMaxAuctionDuration, maxBid, otherPerson)
	// start the auction
	auctionID, err := k.startAuction(ctx, &auction, initiatorOutput)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

func (k Keeper) startAuction(ctx sdk.Context, auction types.Auction, initiatorOutput types.BankOutput) (types.ID, sdk.Error) {
	// get ID
	newAuctionID, err := k.getNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}
	// set ID
	auction.SetID(newAuctionID)

	// subtract coins from initiator
	_, err = k.bankKeeper.SubtractCoins(ctx, initiatorOutput.Address, sdk.NewCoins(initiatorOutput.Coin))
	if err != nil {
		return 0, err
	}

	// store auction
	k.SetAuction(ctx, auction)
	k.incrementNextAuctionID(ctx)
	return newAuctionID, nil
}

// PlaceBid places a bid on any auction.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID types.ID, bidder sdk.AccAddress, bid sdk.Coin, lot sdk.Coin) sdk.Error {

	// get auction from store
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}

	// place bid
	coinOutputs, coinInputs, err := auction.PlaceBid(types.EndTime(ctx.BlockHeight()), bidder, lot, bid) // update auction according to what type of auction it is // TODO should this return updated Auction to be more immutable?
	if err != nil {
		return err
	}
	// TODO this will fail if someone tries to update their bid without the full bid amount sitting in their account
	// sub outputs
	for _, output := range coinOutputs {
		_, err = k.bankKeeper.SubtractCoins(ctx, output.Address, sdk.NewCoins(output.Coin)) // TODO handle errors properly here. All coin transfers should be atomic. InputOutputCoins may work
		if err != nil {
			panic(err)
		}
	}
	// add inputs
	for _, input := range coinInputs {
		_, err = k.bankKeeper.AddCoins(ctx, input.Address, sdk.NewCoins(input.Coin)) // TODO errors
		if err != nil {
			panic(err)
		}
	}

	// store updated auction
	k.SetAuction(ctx, auction)

	return nil
}

// CloseAuction closes an auction and distributes funds to the seller and highest bidder.
// TODO because this is called by the end blocker, it has to be valid for the duration of the EndTime block. Should maybe move this to a begin blocker?
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID types.ID) sdk.Error {

	// get the auction from the store
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}
	// error if auction has not reached the end time
	if ctx.BlockHeight() < int64(auction.GetEndTime()) { // auctions close at the end of the block with blockheight == EndTime
		return sdk.ErrInternal(fmt.Sprintf("auction can't be closed as curent block height (%v) is under auction end time (%v)", ctx.BlockHeight(), auction.GetEndTime()))
	}
	// payout to the last bidder
	coinInput := auction.GetPayout()
	_, err := k.bankKeeper.AddCoins(ctx, coinInput.Address, sdk.NewCoins(coinInput.Coin))
	if err != nil {
		return err
	}

	// delete auction from store (and queue)
	k.deleteAuction(ctx, auctionID)

	return nil
}

// ---------- Store methods ----------
// Use these to add and remove auction from the store.

// getNextAuctionID gets the next available global AuctionID
func (k Keeper) getNextAuctionID(ctx sdk.Context) (types.ID, sdk.Error) { // TODO don't need error return here
	// get next ID from store
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		// if not found, set the id at 0
		bz = k.cdc.MustMarshalBinaryLengthPrefixed(types.ID(0))
		store.Set(k.getNextAuctionIDKey(), bz)
		// TODO Why does the gov module set the id in genesis? :
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
		//return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set") // TODO is this needed? Why not just set it zero here?
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
	k.insertIntoQueue(ctx, auction.GetEndTime(), auction.GetID())
}

// getAuction gets an auction from the store by auctionID
func (k Keeper) GetAuction(ctx sdk.Context, auctionID types.ID) (types.Auction, bool) {
	var auction types.Auction

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getAuctionKey(auctionID))
	if bz == nil {
		return auction, false // TODO what is the correct behavior when an auction is not found? gov module follows this pattern of returning a bool
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auction)
	return auction, true
}

// deleteAuction removes an auction from the store without any validation
func (k Keeper) deleteAuction(ctx sdk.Context, auctionID types.ID) {
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
func (k Keeper) insertIntoQueue(ctx sdk.Context, endTime types.EndTime, auctionID types.ID) {
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
func (k Keeper) removeFromQueue(ctx sdk.Context, endTime types.EndTime, auctionID types.ID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getQueueElementKey(endTime, auctionID))
}

// Returns an iterator for all the auctions in the queue that expire by endTime
func (k Keeper) GetQueueIterator(ctx sdk.Context, endTime types.EndTime) sdk.Iterator { // TODO rename to "getAuctionsByExpiry" ?
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
func getQueueElementKeyPrefix(endTime types.EndTime) []byte {
	return bytes.Join([][]byte{
		queueKeyPrefix,
		sdk.Uint64ToBigEndian(uint64(endTime)), // TODO check this gives correct ordering
	}, keyDelimiter)
}

// Returns the key for an auctionID in the queue
func getQueueElementKey(endTime types.EndTime, auctionID types.ID) []byte {
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
