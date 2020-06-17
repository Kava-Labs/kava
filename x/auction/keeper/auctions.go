package keeper

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/x/auction/types"
)

// StartSurplusAuction starts a new surplus (forward) auction.
func (k Keeper) StartSurplusAuction(ctx sdk.Context, seller string, lot sdk.Coin, bidDenom string) (uint64, error) {
	auction := types.NewSurplusAuction(
		seller,
		lot,
		bidDenom,
		types.DistantFuture,
	)

	// NOTE: for the duration of the auction the auction module account holds the lot
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(lot))
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
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auctionID)),
			sdk.NewAttribute(types.AttributeKeyAuctionType, auction.GetType()),
			sdk.NewAttribute(types.AttributeKeyBid, auction.Bid.String()),
			sdk.NewAttribute(types.AttributeKeyLot, auction.Lot.String()),
		),
	)
	return auctionID, nil
}

// StartDebtAuction starts a new debt (reverse) auction.
func (k Keeper) StartDebtAuction(ctx sdk.Context, buyer string, bid sdk.Coin, initialLot sdk.Coin, debt sdk.Coin) (uint64, error) {

	auction := types.NewDebtAuction(
		buyer,
		bid,
		initialLot,
		types.DistantFuture,
		debt,
	)

	// This auction type mints coins at close. Need to check module account has minting privileges to avoid potential err in endblocker.
	macc := k.accountKeeper.GetModuleAccount(ctx, buyer)
	if !macc.HasPermission(auth.Minter) {
		panic(fmt.Errorf("module '%s' does not have '%s' permission", buyer, auth.Minter))
	}

	// NOTE: for the duration of the auction the auction module account holds the debt
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, buyer, types.ModuleName, sdk.NewCoins(debt))
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
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auctionID)),
			sdk.NewAttribute(types.AttributeKeyAuctionType, auction.GetType()),
			sdk.NewAttribute(types.AttributeKeyBid, auction.Bid.String()),
			sdk.NewAttribute(types.AttributeKeyLot, auction.Lot.String()),
		),
	)
	return auctionID, nil
}

// StartCollateralAuction starts a new collateral (2-phase) auction.
func (k Keeper) StartCollateralAuction(
	ctx sdk.Context, seller string, lot, maxBid sdk.Coin,
	lotReturnAddrs []sdk.AccAddress, lotReturnWeights []sdk.Int, debt sdk.Coin,
) (uint64, error) {
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
		debt,
	)

	// NOTE: for the duration of the auction the auction module account holds the debt and the lot
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(lot))
	if err != nil {
		return 0, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, seller, types.ModuleName, sdk.NewCoins(debt))
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
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auctionID)),
			sdk.NewAttribute(types.AttributeKeyAuctionType, auction.GetType()),
			sdk.NewAttribute(types.AttributeKeyBid, auction.Bid.String()),
			sdk.NewAttribute(types.AttributeKeyLot, auction.Lot.String()),
			sdk.NewAttribute(types.AttributeKeyMaxBid, auction.MaxBid.String()),
		),
	)
	return auctionID, nil
}

// PlaceBid places a bid on any auction.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID uint64, bidder sdk.AccAddress, newAmount sdk.Coin) error {

	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdkerrors.Wrapf(types.ErrAuctionNotFound, "%d", auctionID)
	}

	// validation common to all auctions
	if ctx.BlockTime().After(auction.GetEndTime()) {
		return sdkerrors.Wrapf(types.ErrAuctionHasExpired, "%d", auctionID)
	}

	// move coins and return updated auction
	var (
		err            error
		updatedAuction types.Auction
	)
	switch auctionType := auction.(type) {
	case types.SurplusAuction:
		updatedAuction, err = k.PlaceBidSurplus(ctx, auctionType, bidder, newAmount)
	case types.DebtAuction:
		updatedAuction, err = k.PlaceBidDebt(ctx, auctionType, bidder, newAmount)
	case types.CollateralAuction:
		if !auctionType.IsReversePhase() {
			updatedAuction, err = k.PlaceForwardBidCollateral(ctx, auctionType, bidder, newAmount)
		} else {
			updatedAuction, err = k.PlaceReverseBidCollateral(ctx, auctionType, bidder, newAmount)
		}
	default:
		err = sdkerrors.Wrap(types.ErrUnrecognizedAuctionType, auction.GetType())
	}

	if err != nil {
		return err
	}

	k.SetAuction(ctx, updatedAuction)

	return nil
}

// PlaceBidSurplus places a forward bid on a surplus auction, moving coins and returning the updated auction.
func (k Keeper) PlaceBidSurplus(ctx sdk.Context, auction types.SurplusAuction, bidder sdk.AccAddress, bid sdk.Coin) (types.SurplusAuction, error) {
	// Validate new bid
	if bid.Denom != auction.Bid.Denom {
		return auction, sdkerrors.Wrapf(types.ErrInvalidBidDenom, "%s ≠ %s)", bid.Denom, auction.Bid.Denom)
	}
	minNewBidAmt := auction.Bid.Amount.Add( // new bids must be some % greater than old bid, and at least 1 larger to avoid replacing an old bid at no cost
		sdk.MaxInt(
			sdk.NewInt(1),
			sdk.NewDecFromInt(auction.Bid.Amount).Mul(k.GetParams(ctx).IncrementSurplus).RoundInt(),
		),
	)
	if bid.Amount.LT(minNewBidAmt) {
		return auction, sdkerrors.Wrapf(types.ErrBidTooSmall, "%s < %s%s", bid, minNewBidAmt, auction.Bid.Denom)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid, or the amount being zero (sending zero coins produces meaningless send events).
	if !bidder.Equals(auction.Bidder) && !auction.Bid.IsZero() {
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, auction.Bidder, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
	}
	// Increase in bid is burned
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, bidder, auction.Initiator, sdk.NewCoins(bid.Sub(auction.Bid)))
	if err != nil {
		return auction, err
	}
	err = k.bankKeeper.BurnCoins(ctx, auction.Initiator, sdk.NewCoins(bid.Sub(auction.Bid)))
	if err != nil {
		return auction, err
	}

	// Update Auction
	auction.Bidder = bidder
	auction.Bid = bid
	if !auction.HasReceivedBids {
		auction.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
		auction.HasReceivedBids = true
	}
	auction.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), auction.MaxEndTime) // increment timeout, up to MaxEndTime

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, auction.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyBid, auction.Bid.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", auction.EndTime.Unix())),
		),
	)

	return auction, nil
}

// PlaceForwardBidCollateral places a forward bid on a collateral auction, moving coins and returning the updated auction.
func (k Keeper) PlaceForwardBidCollateral(ctx sdk.Context, auction types.CollateralAuction, bidder sdk.AccAddress, bid sdk.Coin) (types.CollateralAuction, error) {
	// Validate new bid
	if bid.Denom != auction.Bid.Denom {
		return auction, sdkerrors.Wrapf(types.ErrInvalidBidDenom, "%s ≠ %s", bid.Denom, auction.Bid.Denom)
	}
	if auction.IsReversePhase() {
		panic("cannot place forward bid on auction in reverse phase")
	}
	minNewBidAmt := auction.Bid.Amount.Add( // new bids must be some % greater than old bid, and at least 1 larger to avoid replacing an old bid at no cost
		sdk.MaxInt(
			sdk.NewInt(1),
			sdk.NewDecFromInt(auction.Bid.Amount).Mul(k.GetParams(ctx).IncrementCollateral).RoundInt(),
		),
	)
	minNewBidAmt = sdk.MinInt(minNewBidAmt, auction.MaxBid.Amount) // allow new bids to hit MaxBid even though it may be less than the increment %
	if bid.Amount.LT(minNewBidAmt) {
		return auction, sdkerrors.Wrapf(types.ErrBidTooSmall, "%s < %s%s", bid, minNewBidAmt, auction.Bid.Denom)
	}
	if auction.MaxBid.IsLT(bid) {
		return auction, sdkerrors.Wrapf(types.ErrBidTooLarge, "%s > %s", bid, auction.MaxBid)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid, and the amount being zero (sending zero coins produces meaningless send events).
	if !bidder.Equals(auction.Bidder) && !auction.Bid.IsZero() {
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, auction.Bidder, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
	}
	// Increase in bid sent to auction initiator
	bidIncrement := bid.Sub(auction.Bid)
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, bidder, auction.Initiator, sdk.NewCoins(bidIncrement))
	if err != nil {
		return auction, err
	}
	// Debt coins are sent to liquidator (until there is no CorrespondingDebt left). Amount sent is equal to bidIncrement (or whatever is left if < bidIncrement).
	if auction.CorrespondingDebt.IsPositive() {

		debtAmountToReturn := sdk.MinInt(bidIncrement.Amount, auction.CorrespondingDebt.Amount)
		debtToReturn := sdk.NewCoin(auction.CorrespondingDebt.Denom, debtAmountToReturn)

		err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, auction.Initiator, sdk.NewCoins(debtToReturn))
		if err != nil {
			return auction, err
		}
		auction.CorrespondingDebt = auction.CorrespondingDebt.Sub(debtToReturn) // debtToReturn will always be ≤ auction.CorrespondingDebt from the MinInt above
	}

	// Update Auction
	auction.Bidder = bidder
	auction.Bid = bid
	if !auction.HasReceivedBids {
		auction.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
		auction.HasReceivedBids = true
	}
	auction.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), auction.MaxEndTime) // increment timeout, up to MaxEndTime

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, auction.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyBid, auction.Bid.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", auction.EndTime.Unix())),
		),
	)

	return auction, nil
}

// PlaceReverseBidCollateral places a reverse bid on a collateral auction, moving coins and returning the updated auction.
func (k Keeper) PlaceReverseBidCollateral(ctx sdk.Context, auction types.CollateralAuction, bidder sdk.AccAddress, lot sdk.Coin) (types.CollateralAuction, error) {
	// Validate new bid
	if lot.Denom != auction.Lot.Denom {
		return auction, sdkerrors.Wrapf(types.ErrInvalidLotDenom, lot.Denom, auction.Lot.Denom)
	}
	if !auction.IsReversePhase() {
		panic("cannot place reverse bid on auction in forward phase")
	}
	maxNewLotAmt := auction.Lot.Amount.Sub( // new lot must be some % less than old lot, and at least 1 smaller to avoid replacing an old bid at no cost
		sdk.MaxInt(
			sdk.NewInt(1),
			sdk.NewDecFromInt(auction.Lot.Amount).Mul(k.GetParams(ctx).IncrementCollateral).RoundInt(),
		),
	)
	if lot.Amount.GT(maxNewLotAmt) {
		return auction, sdkerrors.Wrapf(types.ErrLotTooLarge, "%s > %s%s", lot, maxNewLotAmt, auction.Lot.Denom)
	}
	if lot.IsNegative() {
		return auction, sdkerrors.Wrapf(types.ErrLotTooSmall, "%s < 0%s", lot, auction.Lot.Denom)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid
	if !bidder.Equals(auction.Bidder) {
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, auction.Bidder, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
	}

	// Decrease in lot is sent to weighted addresses (normally the CDP depositors)
	// Note: splitting an integer amount across weighted buckets results in small errors.
	lotPayouts, err := splitCoinIntoWeightedBuckets(auction.Lot.Sub(lot), auction.LotReturns.Weights)
	if err != nil {
		return auction, err
	}
	for i, payout := range lotPayouts {
		// if the payout amount is 0, don't send 0 coins
		if !payout.IsPositive() {
			continue
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, auction.LotReturns.Addresses[i], sdk.NewCoins(payout))
		if err != nil {
			return auction, err
		}
	}

	// Update Auction
	auction.Bidder = bidder
	auction.Lot = lot
	if !auction.HasReceivedBids {
		auction.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
		auction.HasReceivedBids = true
	}
	auction.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), auction.MaxEndTime) // increment timeout, up to MaxEndTime

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, auction.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyLot, auction.Lot.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", auction.EndTime.Unix())),
		),
	)

	return auction, nil
}

// PlaceBidDebt places a reverse bid on a debt auction, moving coins and returning the updated auction.
func (k Keeper) PlaceBidDebt(ctx sdk.Context, auction types.DebtAuction, bidder sdk.AccAddress, lot sdk.Coin) (types.DebtAuction, error) {
	// Validate new bid
	if lot.Denom != auction.Lot.Denom {
		return auction, sdkerrors.Wrapf(types.ErrInvalidLotDenom, lot.Denom, auction.Lot.Denom)
	}
	maxNewLotAmt := auction.Lot.Amount.Sub( // new lot must be some % less than old lot, and at least 1 smaller to avoid replacing an old bid at no cost
		sdk.MaxInt(
			sdk.NewInt(1),
			sdk.NewDecFromInt(auction.Lot.Amount).Mul(k.GetParams(ctx).IncrementDebt).RoundInt(),
		),
	)
	if lot.Amount.GT(maxNewLotAmt) {
		return auction, sdkerrors.Wrapf(types.ErrLotTooLarge, "%s > %s%s", lot, maxNewLotAmt, auction.Lot.Denom)
	}
	if lot.IsNegative() {
		return auction, sdkerrors.Wrapf(types.ErrLotTooSmall, "%s ≤ %s%s", lot, sdk.ZeroInt(), auction.Lot.Denom)
	}

	// New bidder pays back old bidder
	// Catch edge cases of a bidder replacing their own bid
	if !bidder.Equals(auction.Bidder) {
		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, bidder, types.ModuleName, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, auction.Bidder, sdk.NewCoins(auction.Bid))
		if err != nil {
			return auction, err
		}
	}
	// Debt coins are sent to liquidator the first time a bid is placed. Amount sent is equal to min of Bid and amount of debt.
	if auction.Bidder.Equals(supply.NewModuleAddress(auction.Initiator)) {

		debtAmountToReturn := sdk.MinInt(auction.Bid.Amount, auction.CorrespondingDebt.Amount)
		debtToReturn := sdk.NewCoin(auction.CorrespondingDebt.Denom, debtAmountToReturn)

		err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, auction.Initiator, sdk.NewCoins(debtToReturn))
		if err != nil {
			return auction, err
		}
		auction.CorrespondingDebt = auction.CorrespondingDebt.Sub(debtToReturn) // debtToReturn will always be ≤ auction.CorrespondingDebt from the MinInt above
	}

	// Update Auction
	auction.Bidder = bidder
	auction.Lot = lot
	if !auction.HasReceivedBids {
		auction.MaxEndTime = ctx.BlockTime().Add(k.GetParams(ctx).MaxAuctionDuration) // set maximum ending time on receipt of first bid
		auction.HasReceivedBids = true
	}
	auction.EndTime = earliestTime(ctx.BlockTime().Add(k.GetParams(ctx).BidDuration), auction.MaxEndTime) // increment timeout, up to MaxEndTime

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionBid,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auction.ID)),
			sdk.NewAttribute(types.AttributeKeyBidder, auction.Bidder.String()),
			sdk.NewAttribute(types.AttributeKeyLot, auction.Lot.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, fmt.Sprintf("%d", auction.EndTime.Unix())),
		),
	)

	return auction, nil
}

// CloseAuction closes an auction and distributes funds to the highest bidder.
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID uint64) error {
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdkerrors.Wrapf(types.ErrAuctionNotFound, "%d", auctionID)
	}

	if ctx.BlockTime().Before(auction.GetEndTime()) {
		return sdkerrors.Wrapf(types.ErrAuctionHasNotExpired, "block time %s, auction end time %s", ctx.BlockTime().UTC(), auction.GetEndTime().UTC())
	}

	// payout to the last bidder
	var err error
	switch auc := auction.(type) {
	case types.SurplusAuction:
		err = k.PayoutSurplusAuction(ctx, auc)
	case types.DebtAuction:
		err = k.PayoutDebtAuction(ctx, auc)
	case types.CollateralAuction:
		err = k.PayoutCollateralAuction(ctx, auc)
	default:
		err = sdkerrors.Wrap(types.ErrUnrecognizedAuctionType, auc.GetType())
	}

	if err != nil {
		return err
	}

	k.DeleteAuction(ctx, auctionID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuctionClose,
			sdk.NewAttribute(types.AttributeKeyAuctionID, fmt.Sprintf("%d", auctionID)),
			sdk.NewAttribute(types.AttributeKeyCloseBlock, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
	return nil
}

// PayoutDebtAuction pays out the proceeds for a debt auction, first minting the coins.
func (k Keeper) PayoutDebtAuction(ctx sdk.Context, auction types.DebtAuction) error {
	// create the coins that are needed to pay off the debt
	err := k.bankKeeper.MintCoins(ctx, auction.Initiator, sdk.NewCoins(auction.Lot))
	if err != nil {
		panic(fmt.Errorf("could not mint coins: %w", err))
	}
	// send the new coins from the initiator module to the bidder
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, auction.Initiator, auction.Bidder, sdk.NewCoins(auction.Lot))
	if err != nil {
		return err
	}
	// if there is remaining debt, return it to the calling module to manage
	if !auction.CorrespondingDebt.IsPositive() {
		return nil
	}

	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, auction.Initiator, sdk.NewCoins(auction.CorrespondingDebt))
}

// PayoutSurplusAuction pays out the proceeds for a surplus auction.
func (k Keeper) PayoutSurplusAuction(ctx sdk.Context, auction types.SurplusAuction) error {
	// Send the tokens from the auction module account where they are being managed to the bidder who won the auction
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, auction.Bidder, sdk.NewCoins(auction.Lot))
}

// PayoutCollateralAuction pays out the proceeds for a collateral auction.
func (k Keeper) PayoutCollateralAuction(ctx sdk.Context, auction types.CollateralAuction) error {
	// Send the tokens from the auction module account where they are being managed to the bidder who won the auction
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, auction.Bidder, sdk.NewCoins(auction.Lot))
	if err != nil {
		return err
	}

	// if there is remaining debt after the auction, send it back to the initiating module for management
	if !auction.CorrespondingDebt.IsPositive() {
		return nil
	}

	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, auction.Initiator, sdk.NewCoins(auction.CorrespondingDebt))
}

// CloseExpiredAuctions iterates over all the auctions stored by until the current
// block timestamp and that are past (or at) their ending times and closes them,
// paying out to the highest bidder.
func (k Keeper) CloseExpiredAuctions(ctx sdk.Context) error {
	var err error
	k.IterateAuctionsByTime(ctx, ctx.BlockTime(), func(id uint64) (stop bool) {
		err = k.CloseAuction(ctx, id)
		if err != nil && !errors.Is(err, types.ErrAuctionNotFound) {
			// stop iteration
			return true
		}
		// reset error in case the last element had an ErrAuctionNotFound
		err = nil
		return false
	})

	return err
}

// earliestTime returns the earliest of two times.
func earliestTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2 // also returned if times are equal
}

// splitCoinIntoWeightedBuckets divides up some amount of coins according to some weights.
func splitCoinIntoWeightedBuckets(coin sdk.Coin, buckets []sdk.Int) ([]sdk.Coin, error) {
	amounts := splitIntIntoWeightedBuckets(coin.Amount, buckets)
	result := make([]sdk.Coin, len(amounts))
	for i, a := range amounts {
		result[i] = sdk.NewCoin(coin.Denom, a)
	}
	return result, nil
}
