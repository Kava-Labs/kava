package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/kava-labs/kava/x/auction/types"
)

// StartSurplusAuction starts a new surplus (forward) auction.
func (k Keeper) StartSurplusAuction(ctx sdk.Context, seller string, lot sdk.Coin, bidDenom string) (uint64, sdk.Error) {

	auction := types.NewSurplusAuction(
		seller,
		lot,
		bidDenom,
		types.DistantFuture)

	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(lot))
	if err != nil {
		return 0, err
	}

	auctionID, err := k.StoreNewAuction(ctx, auction)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionStart,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.GetID())),
			sdk.NewAttribute(types.AttributeKeyAuctionType, auction.GetType()),
			sdk.NewAttribute(types.AttributeKeyBidDenom, auction.Bid.Denom),
			sdk.NewAttribute(types.AttributeKeyLotDenom, auction.Lot.Denom),
		),
	)
	return auctionID, nil
}

// StartDebtAuction starts a new debt (reverse) auction.
func (k Keeper) StartDebtAuction(ctx sdk.Context, buyer string, bid sdk.Coin, initialLot sdk.Coin, debt sdk.Coin) (uint64, sdk.Error) {

	auction := types.NewDebtAuction(
		buyer,
		bid,
		initialLot,
		types.DistantFuture,
		debt)

	// This auction type mints coins at close. Need to check module account has minting privileges to avoid potential err in endblocker.
	macc := k.supplyKeeper.GetModuleAccount(ctx, buyer)
	if !macc.HasPermission(supply.Minter) {
		return 0, types.ErrInvalidModulePermissions(k.codespace, supply.Minter)
	}

	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, buyer, types.ModuleName, sdk.NewCoins(debt))
	if err != nil {
		return 0, err
	}

	auctionID, err := k.StoreNewAuction(ctx, auction)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionStart,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.GetID())),
			sdk.NewAttribute(types.AttributeKeyAuctionType, auction.GetType()),
			sdk.NewAttribute(types.AttributeKeyBidDenom, auction.Bid.Denom),
			sdk.NewAttribute(types.AttributeKeyLotDenom, auction.Lot.Denom),
		),
	)
	return auctionID, nil
}

// StartCollateralAuction starts a new collateral (2-phase) auction.
func (k Keeper) StartCollateralAuction(ctx sdk.Context, seller string, lot sdk.Coin, maxBid sdk.Coin, lotReturnAddrs []sdk.AccAddress, lotReturnWeights []sdk.Int, debt sdk.Coin) (uint64, sdk.Error) {

	weightedAddresses, err := types.NewWeightedAddresses(lotReturnAddrs, lotReturnWeights)
	if err != nil {
		return 0, err
	}
	auction := types.NewCollateralAuction(
		seller,
		lot,
		types.DistantFuture,
		maxBid,
		weightedAddresses,
		debt)

	err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(lot))
	if err != nil {
		return 0, err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(debt))
	if err != nil {
		return 0, err
	}

	auctionID, err := k.StoreNewAuction(ctx, auction)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionStart,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.GetID())),
			sdk.NewAttribute(types.AttributeKeyAuctionType, auction.GetType()),
			sdk.NewAttribute(types.AttributeKeyBidDenom, auction.Bid.Denom),
			sdk.NewAttribute(types.AttributeKeyLotDenom, auction.Lot.Denom),
		),
	)
	return auctionID, nil
}

// PlaceBid places a bid on any auction.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID uint64, bidder sdk.AccAddress, newAmount sdk.Coin) sdk.Error {

	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return types.ErrAuctionNotFound(k.codespace, auctionID)
	}

	// validation common to all auctions
	if ctx.BlockTime().After(auction.GetEndTime()) {
		return types.ErrAuctionHasExpired(k.codespace, auctionID)
	}

	// move coins and return updated auction
	var err sdk.Error
	var updatedAuction types.Auction
	switch a := auction.(type) {
	case types.SurplusAuction:
		if updatedAuction, err = k.PlaceBidSurplus(ctx, a, bidder, newAmount); err != nil {
			return err
		}
	case types.DebtAuction:
		if updatedAuction, err = k.PlaceBidDebt(ctx, a, bidder, newAmount); err != nil {
			return err
		}
	case types.CollateralAuction:
		if !a.IsReversePhase() {
			updatedAuction, err = k.PlaceForwardBidCollateral(ctx, a, bidder, newAmount)
		} else {
			updatedAuction, err = k.PlaceReverseBidCollateral(ctx, a, bidder, newAmount)
		}
		if err != nil {
			return err
		}
	default:
		return types.ErrUnrecognizedAuctionType(k.codespace)
	}

	k.SetAuction(ctx, updatedAuction)

	return nil
}

// PlaceBidSurplus places a forward bid on a surplus auction, moving coins and returning the updated auction.
func (k Keeper) PlaceBidSurplus(ctx sdk.Context, a types.SurplusAuction, bidder sdk.AccAddress, bid sdk.Coin) (types.SurplusAuction, sdk.Error) {
	// Validate new bid
	if bid.Denom != a.Bid.Denom {
		return a, types.ErrInvalidBidDenom(k.codespace, bid.Denom, a.Bid.Denom)
	}
	if !a.Bid.IsLT(bid) {
		return a, types.ErrBidTooSmall(k.codespace, bid, a.Bid)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid, or the amount being zero (sending zero coins produces meaningless send events).
	if !bidder.Equals(a.Bidder) && !a.Bid.IsZero() {
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}
	// Increase in bid is burned
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
	if !a.HasReceivedBids {
		a.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
	}
	a.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), a.MaxEndTime) // increment timeout, up to MaxEndTime
	a.HasReceivedBids = true

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", a.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, a.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyBidAmount, a.Bid.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", a.EndTime.Unix())),
		),
	)

	return a, nil
}

// PlaceForwardBidCollateral places a forward bid on a collateral auction, moving coins and returning the updated auction.
func (k Keeper) PlaceForwardBidCollateral(ctx sdk.Context, a types.CollateralAuction, bidder sdk.AccAddress, bid sdk.Coin) (types.CollateralAuction, sdk.Error) {
	// Validate new bid
	if bid.Denom != a.Bid.Denom {
		return a, types.ErrInvalidBidDenom(k.codespace, bid.Denom, a.Bid.Denom)
	}
	if a.IsReversePhase() {
		return a, types.ErrCollateralAuctionIsInReversePhase(k.codespace, a.ID)
	}
	if !a.Bid.IsLT(bid) {
		return a, types.ErrBidTooSmall(k.codespace, bid, a.Bid)
	}
	if a.MaxBid.IsLT(bid) {
		return a, types.ErrBidTooLarge(k.codespace, bid, a.MaxBid)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid, and the amount being zero (sending zero coins produces meaningless send events).
	if !bidder.Equals(a.Bidder) && !a.Bid.IsZero() {
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}
	// Increase in bid sent to auction initiator
	bidIncrement := bid.Sub(a.Bid)
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, a.Initiator, sdk.NewCoins(bidIncrement))
	if err != nil {
		return a, err
	}
	// Debt coins are sent to liquidator (until there is no CorrespondingDebt left). Amount sent is equal to bidIncrement (or whatever is left if < bidIncrement).
	if a.CorrespondingDebt.IsPositive() {

		debtAmountToReturn := sdk.MinInt(bidIncrement.Amount, a.CorrespondingDebt.Amount)
		debtToReturn := sdk.NewCoin(a.CorrespondingDebt.Denom, debtAmountToReturn)

		err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, a.Initiator, sdk.NewCoins(debtToReturn))
		if err != nil {
			return a, err
		}
		a.CorrespondingDebt = a.CorrespondingDebt.Sub(debtToReturn) // debtToReturn will always be ≤ a.CorrespondingDebt from the MinInt above
	}

	// Update Auction
	a.Bidder = bidder
	a.Bid = bid
	if !a.HasReceivedBids {
		a.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
	}
	a.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), a.MaxEndTime) // increment timeout, up to MaxEndTime
	a.HasReceivedBids = true

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", a.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, a.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyBidAmount, a.Bid.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", a.EndTime.Unix())),
		),
	)

	return a, nil
}

// PlaceReverseBidCollateral places a reverse bid on a collateral auction, moving coins and returning the updated auction.
func (k Keeper) PlaceReverseBidCollateral(ctx sdk.Context, a types.CollateralAuction, bidder sdk.AccAddress, lot sdk.Coin) (types.CollateralAuction, sdk.Error) {
	// Validate new bid
	if lot.Denom != a.Lot.Denom {
		return a, types.ErrInvalidLotDenom(k.codespace, lot.Denom, a.Lot.Denom)
	}
	if !a.IsReversePhase() {
		return a, types.ErrCollateralAuctionIsInForwardPhase(k.codespace, a.ID)
	}
	if !lot.IsLT(a.Lot) {
		return a, types.ErrLotTooLarge(k.codespace, lot, a.Lot)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid
	if !bidder.Equals(a.Bidder) {
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}
	// Decrease in lot is sent to weighted addresses (normally the CDP depositors)
	// TODO paying out rateably to cdp depositors is vulnerable to errors compounding over multiple bids - check this can't be gamed.
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
	if !a.HasReceivedBids {
		a.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
	}
	a.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), a.MaxEndTime) // increment timeout, up to MaxEndTime
	a.HasReceivedBids = true
	a.Phase = "reverse"

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", a.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, a.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyLotAmount, a.Lot.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", a.EndTime.Unix())),
		),
	)

	return a, nil
}

// PlaceBidDebt places a reverse bid on a debt auction, moving coins and returning the updated auction.
func (k Keeper) PlaceBidDebt(ctx sdk.Context, a types.DebtAuction, bidder sdk.AccAddress, lot sdk.Coin) (types.DebtAuction, sdk.Error) {
	// Validate new bid
	if lot.Denom != a.Lot.Denom {
		return a, types.ErrInvalidLotDenom(k.codespace, lot.Denom, a.Lot.Denom)
	}
	if !lot.IsLT(a.Lot) {
		return a, types.ErrLotTooLarge(k.codespace, lot, a.Lot)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid
	if !bidder.Equals(a.Bidder) {
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Bid))
		if err != nil {
			return a, err
		}
	}
	// Debt coins are sent to liquidator the first time a bid is placed. Amount sent is equal to min of Bid and amount of debt.
	if a.Bidder.Equals(supply.NewModuleAddress(a.Initiator)) {

		debtAmountToReturn := sdk.MinInt(a.Bid.Amount, a.CorrespondingDebt.Amount)
		debtToReturn := sdk.NewCoin(a.CorrespondingDebt.Denom, debtAmountToReturn)

		err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, a.Initiator, sdk.NewCoins(debtToReturn))
		if err != nil {
			return a, err
		}
		a.CorrespondingDebt = a.CorrespondingDebt.Sub(debtToReturn) // debtToReturn will always be ≤ a.CorrespondingDebt from the MinInt above
	}

	// Update Auction
	a.Bidder = bidder
	a.Lot = lot
	if !a.HasReceivedBids {
		a.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
	}
	a.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), a.MaxEndTime) // increment timeout, up to MaxEndTime
	a.HasReceivedBids = true

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", a.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, a.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyLotAmount, a.Lot.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", a.EndTime.Unix())),
		),
	)

	return a, nil
}

// CloseAuction closes an auction and distributes funds to the highest bidder.
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID uint64) sdk.Error {

	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return types.ErrAuctionNotFound(k.codespace, auctionID)
	}

	if ctx.BlockTime().Before(auction.GetEndTime()) {
		return types.ErrAuctionHasNotExpired(k.codespace, ctx.BlockTime(), auction.GetEndTime())
	}

	// payout to the last bidder
	switch auc := auction.(type) {
	case types.SurplusAuction:
		if err := k.PayoutSurplusAuction(ctx, auc); err != nil {
			return err
		}
	case types.DebtAuction:
		if err := k.PayoutDebtAuction(ctx, auc); err != nil {
			return err
		}
	case types.CollateralAuction:
		if err := k.PayoutCollateralAuction(ctx, auc); err != nil {
			return err
		}
	default:
		return types.ErrUnrecognizedAuctionType(k.codespace)
	}

	k.DeleteAuction(ctx, auctionID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionClose,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.GetID())),
		),
	)
	return nil
}

// PayoutDebtAuction pays out the proceeds for a debt auction, first minting the coins.
func (k Keeper) PayoutDebtAuction(ctx sdk.Context, a types.DebtAuction) sdk.Error {
	err := k.supplyKeeper.MintCoins(ctx, a.Initiator, sdk.NewCoins(a.Lot))
	if err != nil {
		return err
	}
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, a.Initiator, a.Bidder, sdk.NewCoins(a.Lot))
	if err != nil {
		return err
	}
	if a.CorrespondingDebt.IsPositive() {
		err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, a.Initiator, sdk.NewCoins(a.CorrespondingDebt))
		if err != nil {
			return err
		}
	}
	return nil
}

// PayoutSurplusAuction pays out the proceeds for a surplus auction.
func (k Keeper) PayoutSurplusAuction(ctx sdk.Context, a types.SurplusAuction) sdk.Error {
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Lot))
	if err != nil {
		return err
	}
	return nil
}

// PayoutCollateralAuction pays out the proceeds for a collateral auction.
func (k Keeper) PayoutCollateralAuction(ctx sdk.Context, a types.CollateralAuction) sdk.Error {
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, a.Bidder, sdk.NewCoins(a.Lot))
	if err != nil {
		return err
	}
	if a.CorrespondingDebt.IsPositive() {
		err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, a.Initiator, sdk.NewCoins(a.CorrespondingDebt))
		if err != nil {
			return err
		}
	}
	return nil
}

// CloseExpiredAuctions finds all auctions that are past (or at) their ending times and closes them, paying out to the highest bidder.
func (k Keeper) CloseExpiredAuctions(ctx sdk.Context) sdk.Error {
	var expiredAuctions []uint64
	k.IterateAuctionsByTime(ctx, ctx.BlockTime(), func(id uint64) bool {
		expiredAuctions = append(expiredAuctions, id)
		return false
	})
	// Note: iteration and auction closing are in separate loops as db should not be modified during iteration // TODO is this correct? gov modifies during iteration
	for _, id := range expiredAuctions {
		if err := k.CloseAuction(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// earliestTime returns the earliest of two times.
func earliestTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2 // also returned if times are equal
}

// splitCoinIntoWeightedBuckets divides up some amount of coins according to some weights.
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
