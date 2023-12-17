package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/kava-labs/kava/x/auction/types"
)

type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.Codec
	paramSubspace paramtypes.Subspace
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
}

// NewKeeper returns a new auction keeper.
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, paramstore paramtypes.Subspace,
	bankKeeper types.BankKeeper, accountKeeper types.AccountKeeper,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubspace: paramstore,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// MustUnmarshalAuction attempts to decode and return an Auction object from
// raw encoded bytes. It panics on error.
func (k Keeper) MustUnmarshalAuction(bz []byte) types.Auction {
	auction, err := k.UnmarshalAuction(bz)
	if err != nil {
		panic(fmt.Errorf("failed to decode auction: %w", err))
	}

	return auction
}

// MustMarshalAuction attempts to encode an Auction object and returns the
// raw encoded bytes. It panics on error.
func (k Keeper) MustMarshalAuction(auction types.Auction) []byte {
	bz, err := k.MarshalAuction(auction)
	if err != nil {
		panic(fmt.Errorf("failed to encode auction: %w", err))
	}

	return bz
}

// MarshalAuction protobuf serializes an Auction interface
func (k Keeper) MarshalAuction(auctionI types.Auction) ([]byte, error) {
	return k.cdc.MarshalInterface(auctionI)
}

// UnmarshalAuction returns an Auction interface from raw encoded auction
// bytes of a Proto-based Auction type
func (k Keeper) UnmarshalAuction(bz []byte) (types.Auction, error) {
	var evi types.Auction
	return evi, k.cdc.UnmarshalInterface(bz, &evi)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetNextAuctionID stores an ID to be used for the next created auction
func (k Keeper) SetNextAuctionID(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.NextAuctionIDKey)
	store.Set(types.NextAuctionIDKey, types.Uint64ToBytes(id))
}

// GetNextAuctionID reads the next available global ID from store
func (k Keeper) GetNextAuctionID(ctx sdk.Context) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.NextAuctionIDKey)
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

	store.Set(types.GetAuctionKey(auction.GetID()), k.MustMarshalAuction(auction))
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

	return k.MustUnmarshalAuction(bz), true
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
func (k Keeper) InsertIntoByTimeIndex(ctx sdk.Context, endTime time.Time, auctionID uint64) {
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
		auction := k.MustUnmarshalAuction(iterator.Value())

		if cb(auction) {
			break
		}
	}
}

// GetAllAuctions returns all auctions from the store
func (k Keeper) GetAllAuctions(ctx sdk.Context) (auctions []types.Auction) {
	k.IterateAuctions(ctx, func(auction types.Auction) bool {
		auctions = append(auctions, auction)
		return false
	})
	return
}
