package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// DistantFuture is a very large time value to use as initial the ending time for auctions.
// It is not set to the max time supported. This can cause problems with time comparisons, see https://stackoverflow.com/a/32620397.
// Also amino panics when encoding times ≥ the start of year 10000.
var DistantFuture = time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)

// Auction is an interface for handling common actions on auctions.
type Auction interface {
	GetID() uint64
	WithID(uint64) Auction
	GetInitiator() string
	GetLot() sdk.Coin
	GetBidder() sdk.AccAddress
	GetBid() sdk.Coin
	GetEndTime() time.Time
}

// Auctions is a slice of auctions.
type Auctions []Auction

// BaseAuction is a common type shared by all Auctions.
type BaseAuction struct {
	ID              uint64
	Initiator       string         // Module name that starts the auction. Pays out Lot.
	Lot             sdk.Coin       // Coins that will paid out by Initiator to the winning bidder.
	Bidder          sdk.AccAddress // Latest bidder. Receiver of Lot.
	Bid             sdk.Coin       // Coins paid into the auction the bidder.
	HasReceivedBids bool           // Whether the auction has received any bids or not.
	EndTime         time.Time      // Current auction closing time. Triggers at the end of the block with time ≥ EndTime.
	MaxEndTime      time.Time      // Maximum closing time. Auctions can close before this but never after.
}

// GetID is a getter for auction ID.
func (a BaseAuction) GetID() uint64 { return a.ID }

// GetInitiator is a getter for auction Initiator.
func (a BaseAuction) GetInitiator() string { return a.Initiator }

// GetLot is a getter for auction Lot.
func (a BaseAuction) GetLot() sdk.Coin { return a.Lot }

// GetBidder is a getter for auction Bidder.
func (a BaseAuction) GetBidder() sdk.AccAddress { return a.Bidder }

// GetBid is a getter for auction Bid.
func (a BaseAuction) GetBid() sdk.Coin { return a.Bid }

// GetEndTime is a getter for auction end time.
func (a BaseAuction) GetEndTime() time.Time { return a.EndTime }

// Validate verifies that the auction end time is before max end time
func (a BaseAuction) Validate() error {
	if a.EndTime.After(a.MaxEndTime) {
		return fmt.Errorf("MaxEndTime < EndTime (%s < %s)", a.MaxEndTime, a.EndTime)
	}
	return nil
}

func (a BaseAuction) String() string {
	return fmt.Sprintf(`Auction %d:
  Initiator:              %s
  Lot:               			%s
  Bidder:            		  %s
  Bid:        						%s
  End Time:   						%s
  Max End Time:      			%s`,
		a.GetID(), a.Initiator, a.Lot,
		a.Bidder, a.Bid, a.GetEndTime().String(),
		a.MaxEndTime.String(),
	)
}

// SurplusAuction is a forward auction that burns what it receives from bids.
// It is normally used to sell off excess pegged asset acquired by the CDP system.
type SurplusAuction struct {
	BaseAuction
}

// WithID returns an auction with the ID set.
func (a SurplusAuction) WithID(id uint64) Auction { a.ID = id; return a }

// Name returns a name for this auction type. Used to identify auctions in event attributes.
func (a SurplusAuction) Name() string { return "surplus" }

// GetModuleAccountCoins returns the total number of coins held in the module account for this auction.
// It is used in genesis initialize the module account correctly.
func (a SurplusAuction) GetModuleAccountCoins() sdk.Coins {
	// a.Bid is paid out on bids, so is never stored in the module account
	return sdk.NewCoins(a.Lot)
}

// NewSurplusAuction returns a new surplus auction.
func NewSurplusAuction(seller string, lot sdk.Coin, bidDenom string, endTime time.Time) SurplusAuction {
	auction := SurplusAuction{BaseAuction{
		// no ID
		Initiator:       seller,
		Lot:             lot,
		Bidder:          nil,
		Bid:             sdk.NewInt64Coin(bidDenom, 0),
		HasReceivedBids: false, // new auctions don't have any bids
		EndTime:         endTime,
		MaxEndTime:      endTime,
	}}
	return auction
}

// DebtAuction is a reverse auction that mints what it pays out.
// It is normally used to acquire pegged asset to cover the CDP system's debts that were not covered by selling collateral.
type DebtAuction struct {
	BaseAuction
	CorrespondingDebt sdk.Coin
}

// WithID returns an auction with the ID set.
func (a DebtAuction) WithID(id uint64) Auction { a.ID = id; return a }

// Name returns a name for this auction type. Used to identify auctions in event attributes.
func (a DebtAuction) Name() string { return "debt" }

// GetModuleAccountCoins returns the total number of coins held in the module account for this auction.
// It is used in genesis initialize the module account correctly.
func (a DebtAuction) GetModuleAccountCoins() sdk.Coins {
	// a.Lot is minted at auction close, so is never stored in the module account
	// a.Bid is paid out on bids, so is never stored in the module account
	return sdk.NewCoins(a.CorrespondingDebt)
}

// NewDebtAuction returns a new debt auction.
func NewDebtAuction(buyerModAccName string, bid sdk.Coin, initialLot sdk.Coin, endTime time.Time, debt sdk.Coin) DebtAuction {
	// Note: Bidder is set to the initiator's module account address instead of module name. (when the first bid is placed, it is paid out to the initiator)
	// Setting to the module account address bypasses calling supply.SendCoinsFromModuleToModule, instead calls SendCoinsFromModuleToAccount.
	// This isn't a problem currently, but if additional logic/validation was added for sending to coins to Module Accounts, it would be bypassed.
	auction := DebtAuction{
		BaseAuction: BaseAuction{
			// no ID
			Initiator:       buyerModAccName,
			Lot:             initialLot,
			Bidder:          supply.NewModuleAddress(buyerModAccName), // send proceeds from the first bid to the buyer.
			Bid:             bid,                                      // amount that the buyer is buying - doesn't change over course of auction
			HasReceivedBids: false,                                    // new auctions don't have any bids
			EndTime:         endTime,
			MaxEndTime:      endTime},
		CorrespondingDebt: debt,
	}
	return auction
}

// CollateralAuction is a two phase auction.
// Initially, in forward auction phase, bids can be placed up to a max bid.
// Then it switches to a reverse auction phase, where the initial amount up for auction is bid down.
// Unsold Lot is sent to LotReturns, being divided among the addresses by weight.
// Collateral auctions are normally used to sell off collateral seized from CDPs.
type CollateralAuction struct {
	BaseAuction
	CorrespondingDebt sdk.Coin
	MaxBid            sdk.Coin
	LotReturns        WeightedAddresses
}

// WithID returns an auction with the ID set.
func (a CollateralAuction) WithID(id uint64) Auction { a.ID = id; return a }

// Name returns a name for this auction type. Used to identify auctions in event attributes.
func (a CollateralAuction) Name() string { return "collateral" }

// GetModuleAccountCoins returns the total number of coins held in the module account for this auction.
// It is used in genesis initialize the module account correctly.
func (a CollateralAuction) GetModuleAccountCoins() sdk.Coins {
	// a.Bid is paid out on bids, so is never stored in the module account
	return sdk.NewCoins(a.Lot).Add(sdk.NewCoins(a.CorrespondingDebt))
}

// IsReversePhase returns whether the auction has switched over to reverse phase or not.
// Auction initially start in forward phase.
func (a CollateralAuction) IsReversePhase() bool {
	return a.Bid.IsEqual(a.MaxBid)
}

func (a CollateralAuction) String() string {
	return fmt.Sprintf(`Auction %d:
  Initiator:              %s
  Lot:               			%s
  Bidder:            		  %s
  Bid:        						%s
  End Time:   						%s
	Max End Time:      			%s
	Max Bid									%s
	LotReturns						%s`,
		a.GetID(), a.Initiator, a.Lot,
		a.Bidder, a.Bid, a.GetEndTime().String(),
		a.MaxEndTime.String(), a.MaxBid, a.LotReturns,
	)
}

// NewCollateralAuction returns a new collateral auction.
func NewCollateralAuction(seller string, lot sdk.Coin, endTime time.Time, maxBid sdk.Coin, lotReturns WeightedAddresses, debt sdk.Coin) CollateralAuction {
	auction := CollateralAuction{
		BaseAuction: BaseAuction{
			// no ID
			Initiator:       seller,
			Lot:             lot,
			Bidder:          nil,
			Bid:             sdk.NewInt64Coin(maxBid.Denom, 0),
			HasReceivedBids: false, // new auctions don't have any bids
			EndTime:         endTime,
			MaxEndTime:      endTime},
		CorrespondingDebt: debt,
		MaxBid:            maxBid,
		LotReturns:        lotReturns,
	}
	return auction
}

// WeightedAddresses is a type for storing some addresses and associated weights.
type WeightedAddresses struct {
	Addresses []sdk.AccAddress
	Weights   []sdk.Int
}

// NewWeightedAddresses returns a new list addresses with weights.
func NewWeightedAddresses(addrs []sdk.AccAddress, weights []sdk.Int) (WeightedAddresses, sdk.Error) {
	if len(addrs) != len(weights) {
		return WeightedAddresses{}, sdk.ErrInternal("number of addresses doesn't match number of weights")
	}
	for _, w := range weights {
		if w.IsNegative() {
			return WeightedAddresses{}, sdk.ErrInternal("weights contain a negative amount")
		}
	}
	return WeightedAddresses{
		Addresses: addrs,
		Weights:   weights,
	}, nil
}
