package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/auction/types"
)

type Keeper struct {
	supplyKeeper  types.SupplyKeeper
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
}

// NewKeeper returns a new auction keeper.
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, supplyKeeper types.SupplyKeeper, paramstore subspace.Subspace) Keeper {
	if addr := supplyKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		supplyKeeper:  supplyKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubspace: paramstore,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetNextAuctionID stores an ID to be used for the next created auction
func (k Keeper) SetNextAuctionID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextAuctionIDKey, types.Uint64ToBytes(id))
}

// GetNextAuctionID reads the next available global ID from store
func (k Keeper) GetNextAuctionID(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextAuctionIDKey)
	if bz == nil {
		return 0, types.ErrInvalidInitialAuctionID
	}
	return types.Uint64FromBytes(bz), nil
}

// IncrementNextAuctionID increments the next auction ID in the store by 1.
func (k Keeper) IncrementNextAuctionID(ctx sdk.Context) error {
	id, err := k.GetNextAuctionID(ctx)
	if err != nil {
		return err
	}
	k.SetNextAuctionID(ctx, id+1)
	return nil
}

// StoreNewAuction stores an auction, adding a new ID
func (k Keeper) StoreNewAuction(ctx sdk.Context, auction types.Auction) (uint64, error) {
	newAuctionID, err := k.GetNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}
	auction = auction.WithID(newAuctionID)

	k.SetAuction(ctx, auction)

	err = k.IncrementNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}
	return newAuctionID, nil
}

// SetAuction puts the auction into the store, and updates any indexes.
func (k Keeper) SetAuction(ctx sdk.Context, auction types.Auction) {
	// remove the auction from the byTime index if it is already in there
	existingAuction, found := k.GetAuction(ctx, auction.GetID())
	if found {
		k.removeFromByTimeIndex(ctx, existingAuction.GetEndTime(), existingAuction.GetID())
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auction)
	store.Set(types.GetAuctionKey(auction.GetID()), bz)

	k.InsertIntoByTimeIndex(ctx, auction.GetEndTime(), auction.GetID())
}

// GetAuction gets an auction from the store.
func (k Keeper) GetAuction(ctx sdk.Context, auctionID uint64) (types.Auction, bool) {
	var auction types.Auction

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)
	bz := store.Get(types.GetAuctionKey(auctionID))
	if bz == nil {
		return auction, false
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auction)
	return auction, true
}

// DeleteAuction removes an auction from the store, and any indexes.
func (k Keeper) DeleteAuction(ctx sdk.Context, auctionID uint64) {
	auction, found := k.GetAuction(ctx, auctionID)
	if found {
		k.removeFromByTimeIndex(ctx, auction.GetEndTime(), auctionID)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)
	store.Delete(types.GetAuctionKey(auctionID))
}

// InsertIntoByTimeIndex adds an auction ID and end time into the byTime index.
func (k Keeper) InsertIntoByTimeIndex(ctx sdk.Context, endTime time.Time, auctionID uint64) { // TODO make private, and find way to make tests work
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
	store.Set(types.GetAuctionByTimeKey(endTime, auctionID), types.Uint64ToBytes(auctionID))
}

// removeFromByTimeIndex removes an auction ID and end time from the byTime index.
func (k Keeper) removeFromByTimeIndex(ctx sdk.Context, endTime time.Time, auctionID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
	store.Delete(types.GetAuctionByTimeKey(endTime, auctionID))
}

// IterateAuctionByTime provides an iterator over auctions ordered by auction.EndTime.
// For each auction cb will be callled. If cb returns true the iterator will close and stop.
func (k Keeper) IterateAuctionsByTime(ctx sdk.Context, inclusiveCutoffTime time.Time, cb func(auctionID uint64) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionByTimeKeyPrefix)
	iterator := store.Iterator(
		nil, // start at the very start of the prefix store
		sdk.PrefixEndBytes(sdk.FormatTimeBytes(inclusiveCutoffTime)), // include any keys with times equal to inclusiveCutoffTime
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		auctionID := types.Uint64FromBytes(iterator.Value())

		if cb(auctionID) {
			break
		}
	}
}

// IterateAuctions provides an iterator over all stored auctions.
// For each auction, cb will be called. If cb returns true, the iterator will close and stop.
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
