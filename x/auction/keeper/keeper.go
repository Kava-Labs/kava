package keeper

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/supply"

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

// StartForwardAuction starts a normal auction. Known as flap in maker.
func (k Keeper) StartForwardAuction(ctx sdk.Context, seller string, lot sdk.Coin, bidDenom string) (types.ID, sdk.Error) {
	// create auction
	auction := types.NewForwardAuction(seller, lot, bidDenom, ctx.BlockTime().Add(types.DefaultMaxAuctionDuration))

	// take coins from module account
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(lot))
	if err != nil {
		return 0, err
	}
	// store the auction
	auctionID, err := k.storeNewAuction(ctx, auction) // TODO does this need to be a pointer to satisfy the interface?
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

// StartReverseAuction starts an auction where sellers compete by offering decreasing prices. Known as flop in maker.
func (k Keeper) StartReverseAuction(ctx sdk.Context, buyer string, bid sdk.Coin, initialLot sdk.Coin) (types.ID, sdk.Error) {
	// create auction
	auction := types.NewReverseAuction(buyer, bid, initialLot, ctx.BlockTime().Add(types.DefaultMaxAuctionDuration))

	// This auction type mints coins at close. Need to check module account has minting privileges to avoid potential err in endblocker.
	macc := k.supplyKeeper.GetModuleAccount(ctx, buyer)
	if !macc.HasPermission(supply.Minter) { // TODO ideally don't want to import supply
		return 0, sdk.ErrInternal("module does not have minting permissions")
	}
	// store the auction
	auctionID, err := k.storeNewAuction(ctx, &auction)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

// StartForwardReverseAuction starts an auction where bidders bid up to a maxBid, then switch to bidding down on price. Known as flip in maker.
func (k Keeper) StartForwardReverseAuction(ctx sdk.Context, seller string, lot sdk.Coin, maxBid sdk.Coin, otherPerson sdk.AccAddress) (types.ID, sdk.Error) {
	// create auction
	auction := types.NewForwardReverseAuction(seller, lot, ctx.BlockTime().Add(types.DefaultMaxAuctionDuration), maxBid, otherPerson)

	// take coins from module account
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.Coins{lot})
	if err != nil {
		return 0, err
	}
	// store the auction
	auctionID, err := k.storeNewAuction(ctx, &auction)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

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

// PlaceBid places a bid on any auction.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID types.ID, bidder sdk.AccAddress, bid sdk.Coin, lot sdk.Coin) sdk.Error {

	// get auction from store
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}

	// check end time
	if ctx.BlockTime().After(auction.GetEndTime()) {
		return sdk.ErrInternal("auction has closed")
	}

	var err sdk.Error
	var a types.Auction
	switch auc := auction.(type) {
	case types.ForwardAuction:
		a, err = k.PlaceBidForward(ctx, auc, bidder, bid)
		if err != nil {
			return err
		}
	case types.ReverseAuction:
		a, err = k.PlaceBidReverse(ctx, auc, bidder, lot)
		if err != nil {
			return err
		}
	case types.ForwardReverseAuction:
		a, err = k.PlaceBidForwardReverse(ctx, auc, bidder, bid, lot)
		if err != nil {
			return err
		}
	default:
		panic("unrecognized auction type")
	}

	// store updated auction
	k.SetAuction(ctx, a) // TODO maybe move into above funcs

	return nil
}

func (k Keeper) PlaceBidForward(ctx sdk.Context, a types.ForwardAuction, bidder sdk.AccAddress, bid sdk.Coin) (types.ForwardAuction, sdk.Error) {
	// Valid New Bid
	if bid.Denom != a.Bid.Denom {
		return a, sdk.ErrInternal("bid denom doesn't match auction")
	}
	if !a.Bid.IsLT(bid) { // TODO add minimum bid size
		return a, sdk.ErrInternal("bid not greater than last bid")
	}

	// Move Coins
	increment := bid.Sub(a.Bid)
	bidAmtToReturn := a.Bid
	if bidder.Equals(a.Bidder) { // catch edge case of someone updating their bid with a low balance
		bidAmtToReturn = sdk.NewInt64Coin(a.Bid.Denom, 0)
	}
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(bidAmtToReturn.Add(increment)))
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, bidder, sdk.NewCoins(bidAmtToReturn))
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, a.Initiator, sdk.NewCoins(increment)) // increase in bid size is burned
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.BurnCoins(ctx, a.Initiator, sdk.NewCoins(increment))
	if err != nil {
		return a, err
	}

	// Update Auction
	a.Bidder = bidder
	a.Bid = bid
	// increment timeout
	a.EndTime = earliestTime(ctx.BlockTime().Add(types.DefaultMaxBidDuration), a.MaxEndTime) // TODO write a min func for time types

	return a, nil
}
func (k Keeper) PlaceBidForwardReverse(ctx sdk.Context, a types.ForwardReverseAuction, bidder sdk.AccAddress, bid sdk.Coin, lot sdk.Coin) (types.ForwardReverseAuction, sdk.Error) {
	// Validate New Bid // TODO min bid increments, make validation code less confusing
	if !a.Bid.IsEqual(a.MaxBid) {
		// Auction is in forward phase, a bid here can put the auction into forward or reverse phases
		if !a.Bid.IsLT(bid) {
			return a, sdk.ErrInternal("auction in forward phase, new bid not higher than last bid")
		}
		if a.MaxBid.IsLT(bid) {
			return a, sdk.ErrInternal("bid higher than max bid")
		}
		if lot.IsNegative() || a.Lot.IsLT(lot) {
			return a, sdk.ErrInternal("lot out of bounds")
		}
		if lot.IsLT(a.Lot) && !bid.IsEqual(a.MaxBid) {
			return a, sdk.ErrInternal("auction cannot enter reverse phase without bidding max bid")
		}
	} else {
		// Auction is in reverse phase, it can never leave reverse phase
		if !bid.IsEqual(a.MaxBid) {
			return a, sdk.ErrInternal("") // not necessary
		}
		if lot.IsNegative() {
			return a, sdk.ErrInternal("can't bid negative amount")
		}
		if !lot.IsLT(a.Lot) {
			return a, sdk.ErrInternal("auction in reverse phase, new bid not less than previous amount")
		}
	}

	// Move Coins
	bidIncrement := bid.Sub(a.Bid)
	bidAmtToReturn := a.Bid
	lotDecrement := a.Lot.Sub(lot)
	if bidder.Equals(a.Bidder) { // catch edge case of someone updating their bid with a low balance
		bidAmtToReturn = sdk.NewInt64Coin(a.Bid.Denom, 0)
	}
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(bidAmtToReturn.Add(bidIncrement)))
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, bidder, sdk.NewCoins(bidAmtToReturn))
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, a.Initiator, sdk.NewCoins(bidIncrement))
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.OtherPerson, sdk.NewCoins(lotDecrement))
	if err != nil {
		return a, err
	}

	// Update Auction
	a.Bidder = bidder
	a.Lot = lot
	a.Bid = bid
	// increment timeout
	a.EndTime = earliestTime(ctx.BlockTime().Add(types.DefaultMaxBidDuration), a.MaxEndTime)

	return types.ForwardReverseAuction{}, nil
}
func (k Keeper) PlaceBidReverse(ctx sdk.Context, a types.ReverseAuction, bidder sdk.AccAddress, lot sdk.Coin) (types.ReverseAuction, sdk.Error) {
	// Validate New Bid
	if lot.Denom != a.Lot.Denom {
		return a, sdk.ErrInternal("lot denom doesn't match auction")
	}
	if lot.IsNegative() {
		return a, sdk.ErrInternal("lot less than 0")
	}
	if !lot.IsLT(a.Lot) { // TODO add min bid decrements
		return a, sdk.ErrInternal("lot not smaller than last lot")
	}

	// Move Coins
	bidAmtToReturn := a.Bid
	if bidder.Equals(a.Bidder) { // catch edge case of someone updating their bid with a low balance
		bidAmtToReturn = sdk.NewInt64Coin(a.Bid.Denom, 0)
	}
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(bidAmtToReturn))
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, bidder, sdk.NewCoins(bidAmtToReturn))
	if err != nil {
		return a, err
	}

	// Update Auction
	a.Bidder = bidder
	a.Lot = lot
	// increment timeout
	a.EndTime = earliestTime(ctx.BlockTime().Add(types.DefaultMaxBidDuration), a.MaxEndTime)

	return a, nil
}

// CloseAuction closes an auction and distributes funds to the highest bidder.
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID types.ID) sdk.Error {

	// get the auction from the store
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}
	// error if auction has not reached the end time
	if ctx.BlockTime().Before(auction.GetEndTime()) {
		return sdk.ErrInternal(fmt.Sprintf("auction can't be closed as curent block time (%v) is under auction end time (%v)", ctx.BlockTime(), auction.GetEndTime()))
	}

	// payout to the last bidder
	var err sdk.Error
	switch auc := auction.(type) {
	case types.ForwardAuction, types.ForwardReverseAuction:
		err = k.PayoutAuctionLot(ctx, auc)
		if err != nil {
			return err
		}
	case types.ReverseAuction:
		err = k.MintAndPayoutAuctionLot(ctx, auc)
		if err != nil {
			return err
		}
	default:
		panic("unrecognized auction type")
	}

	// Delete auction from store (and queue)
	k.DeleteAuction(ctx, auctionID)

	return nil
}
func (k Keeper) MintAndPayoutAuctionLot(ctx sdk.Context, a types.ReverseAuction) sdk.Error {
	err := k.supplyKeeper.MintCoins(ctx, a.Initiator, sdk.NewCoins(a.Lot))
	if err != nil {
		return err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, a.Initiator, a.Bidder, sdk.NewCoins(a.Lot))
	if err != nil {
		return err
	}
	return nil
}
func (k Keeper) PayoutAuctionLot(ctx sdk.Context, a types.Auction) sdk.Error {
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.GetBidder(), sdk.NewCoins(a.GetLot()))
	if err != nil {
		return err
	}
	return nil
}

// ---------- Store methods ----------
// Use these to add and remove auction from the store.

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
