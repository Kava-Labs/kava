package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/kava-labs/kava/x/auction/types"
)

// StartForwardAuction starts a normal auction that mints the sold coins.
func (k Keeper) StartForwardAuction(ctx sdk.Context, seller string, lot sdk.Coin, bidDenom string) (uint64, sdk.Error) {
	// create auction
	auction := types.NewForwardAuction(seller, lot, bidDenom, ctx.BlockTime().Add(types.DefaultMaxAuctionDuration))

	// take coins from module account
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(lot))
	if err != nil {
		return 0, err
	}
	// store the auction
	auctionID, err := k.StoreNewAuction(ctx, auction)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

// StartReverseAuction starts an auction where sellers compete by offering decreasing prices.
func (k Keeper) StartReverseAuction(ctx sdk.Context, buyer string, bid sdk.Coin, initialLot sdk.Coin) (uint64, sdk.Error) {
	// create auction
	auction := types.NewReverseAuction(buyer, bid, initialLot, ctx.BlockTime().Add(types.DefaultMaxAuctionDuration))

	// This auction type mints coins at close. Need to check module account has minting privileges to avoid potential err in endblocker.
	macc := k.supplyKeeper.GetModuleAccount(ctx, buyer)
	if !macc.HasPermission(supply.Minter) {
		return 0, sdk.ErrInternal("module does not have minting permissions")
	}
	// store the auction
	auctionID, err := k.StoreNewAuction(ctx, auction)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

// StartForwardReverseAuction starts an auction where bidders bid up to a maxBid, then switch to bidding down on price.
func (k Keeper) StartForwardReverseAuction(ctx sdk.Context, seller string, lot sdk.Coin, maxBid sdk.Coin, lotReturnAddrs []sdk.AccAddress, lotReturnWeights []sdk.Int) (uint64, sdk.Error) {
	// create auction
	weightedAddresses, err := types.NewWeightedAddresses(lotReturnAddrs, lotReturnWeights)
	if err != nil {
		return 0, err
	}
	auction := types.NewForwardReverseAuction(seller, lot, ctx.BlockTime().Add(types.DefaultMaxAuctionDuration), maxBid, weightedAddresses)

	// take coins from module account
	err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(lot))
	if err != nil {
		return 0, err
	}
	// store the auction
	auctionID, err := k.StoreNewAuction(ctx, auction)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

// PlaceBid places a bid on any auction.
// TODO passing bid and lot is weird when only one needed
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID uint64, bidder sdk.AccAddress, bid sdk.Coin, lot sdk.Coin) sdk.Error {

	// get auction from store
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}

	// validate
	if ctx.BlockTime().After(auction.GetEndTime()) {
		return sdk.ErrInternal("auction has closed")
	}
	if auction.GetBid().Denom != bid.Denom {
		return sdk.ErrInternal("bid has incorrect denom")
	}
	if auction.GetLot().Denom != lot.Denom {
		return sdk.ErrInternal("lot has incorrect denom")
	}

	// place bid
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
		if !auc.IsReversePhase() {
			a, err = k.PlaceBidForwardReverseForward(ctx, auc, bidder, bid)
		} else {
			a, err = k.PlaceBidForwardReverseReverse(ctx, auc, bidder, lot)
		}
		if err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("unrecognized auction type: %T", auction))
	}

	// store updated auction
	k.SetAuction(ctx, a)
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
	if !bidder.Equals(a.Bidder) && !a.Bid.IsZero() { // catch edge case of someone updating their bid with a low balance, also don't send if amt is zero
		// pay back previous bidder
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}
	// burn increase in bid
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, a.Initiator, sdk.NewCoins(bid.Sub(a.Bid)))
	if err != nil {
		return a, err
	}
	err = k.supplyKeeper.BurnCoins(ctx, a.Initiator, sdk.NewCoins(bid.Sub(a.Bid)))
	if err != nil {
		return a, err
	}

	// Update Auction
	a.Bidder = bidder
	a.Bid = bid
	// increment timeout
	a.EndTime = earliestTime(ctx.BlockTime().Add(types.DefaultBidDuration), a.MaxEndTime)

	return a, nil
}

// TODO naming
func (k Keeper) PlaceBidForwardReverseForward(ctx sdk.Context, a types.ForwardReverseAuction, bidder sdk.AccAddress, bid sdk.Coin) (types.ForwardReverseAuction, sdk.Error) {
	// Validate bid
	if a.IsReversePhase() {
		return a, sdk.ErrInternal("auction is not in forward phase")
	}
	if !a.Bid.IsLT(bid) {
		return a, sdk.ErrInternal("auction in forward phase, new bid not higher than last bid")
	}
	if a.MaxBid.IsLT(bid) {
		return a, sdk.ErrInternal("bid higher than max bid")
	}
	// Move Coins
	// pay back previous bidder
	if !bidder.Equals(a.Bidder) && !a.Bid.IsZero() { // catch edge case of someone updating their bid with a low balance, also don't send if amt is zero
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}
	// pay increase in bid to auction initiator
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, a.Initiator, sdk.NewCoins(bid.Sub(a.Bid)))
	if err != nil {
		return a, err
	}

	// Update Auction
	a.Bidder = bidder
	a.Bid = bid
	// increment timeout
	a.EndTime = earliestTime(ctx.BlockTime().Add(types.DefaultBidDuration), a.MaxEndTime)

	return a, nil
}

func (k Keeper) PlaceBidForwardReverseReverse(ctx sdk.Context, a types.ForwardReverseAuction, bidder sdk.AccAddress, lot sdk.Coin) (types.ForwardReverseAuction, sdk.Error) {
	// Validate bid
	if !a.IsReversePhase() {
		return a, sdk.ErrInternal("auction not in reverse phase")
	}
	if lot.IsNegative() {
		return a, sdk.ErrInternal("can't bid negative amount")
	}
	if !lot.IsLT(a.Lot) {
		return a, sdk.ErrInternal("auction in reverse phase, new bid not less than previous amount")
	}

	// Move Coins
	if !bidder.Equals(a.Bidder) { // catch edge case of someone updating their bid with a low balance
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}
	// FIXME paying out rateably to cdp depositors is vulnerable to errors compounding over multiple bids
	lotPayouts, err := splitCoinIntoWeightedBuckets(a.Lot.Sub(lot), a.LotReturns.Weights)
	if err != nil {
		return a, err
	}
	for i, payout := range lotPayouts {
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.LotReturns.Addresses[i], sdk.NewCoins(payout))
		if err != nil {
			return a, err
		}
	}

	// Update Auction
	a.Bidder = bidder
	a.Lot = lot
	// increment timeout
	a.EndTime = earliestTime(ctx.BlockTime().Add(types.DefaultBidDuration), a.MaxEndTime)

	return a, nil
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
	if !bidder.Equals(a.Bidder) { // catch edge case of someone updating their bid with a low balance
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}

	// Update Auction
	a.Bidder = bidder
	a.Lot = lot
	// increment timeout
	a.EndTime = earliestTime(ctx.BlockTime().Add(types.DefaultBidDuration), a.MaxEndTime)

	return a, nil
}

// CloseAuction closes an auction and distributes funds to the highest bidder.
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID uint64) sdk.Error {

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

// earliestTime returns the earliest of two times.
func earliestTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	} else {
		return t2 // also returned if times are equal
	}
}

func splitCoinIntoWeightedBuckets(coin sdk.Coin, buckets []sdk.Int) ([]sdk.Coin, sdk.Error) {
	for _, bucket := range buckets {
		if bucket.IsNegative() {
			return nil, sdk.ErrInternal("cannot split coin into bucket with negative weight")
		}
	}
	amounts := splitIntIntoWeightedBuckets(coin.Amount, buckets)
	result := make([]sdk.Coin, len(amounts))
	for i, a := range amounts {
		result[i] = sdk.NewCoin(coin.Denom, a)
	}
	return result, nil
}
