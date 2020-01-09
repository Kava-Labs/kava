package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// Auction is an interface for handling common actions on auctions.
type Auction interface {
	GetID() uint64
	WithID(uint64) Auction
	GetEndTime() time.Time
}

// BaseAuction is a common type shared by all Auctions.
type BaseAuction struct {
	ID         uint64
	Initiator  string         // Module name that starts the auction. Pays out Lot.
	Lot        sdk.Coin       // Coins that will paid out by Initiator to the winning bidder.
	Bidder     sdk.AccAddress // Latest bidder. Receiver of Lot.
	Bid        sdk.Coin       // Coins paid into the auction the bidder.
	EndTime    time.Time      // Current auction closing time. Triggers at the end of the block with time ≥ EndTime.
	MaxEndTime time.Time      // Maximum closing time. Auctions can close before this but never after.
}

// GetID is a getter for auction ID.
func (a BaseAuction) GetID() uint64 { return a.ID }

// GetEndTime is a getter for auction end time.
func (a BaseAuction) GetEndTime() time.Time { return a.EndTime }

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

// SurplusAuction is a forward auction that burns what it receives as bids.
type SurplusAuction struct {
	BaseAuction
}

// WithID returns an auction with the ID set.
func (a SurplusAuction) WithID(id uint64) Auction { a.ID = id; return a }

// NewSurplusAuction returns a new surplus auction.
func NewSurplusAuction(seller string, lot sdk.Coin, bidDenom string, endTime time.Time) SurplusAuction {
	auction := SurplusAuction{BaseAuction{
		// no ID
		Initiator:  seller,
		Lot:        lot,
		Bidder:     nil,
		Bid:        sdk.NewInt64Coin(bidDenom, 0),
		EndTime:    endTime,
		MaxEndTime: endTime,
	}}
	return auction
}

// DebtAuction is a reverse auction that mints what it pays out.
type DebtAuction struct {
	BaseAuction
}

// WithID returns an auction with the ID set.
func (a DebtAuction) WithID(id uint64) Auction { a.ID = id; return a }

// NewDebtAuction returns a new debt auction.
func NewDebtAuction(buyerModAccName string, bid sdk.Coin, initialLot sdk.Coin, EndTime time.Time) DebtAuction {
	// Note: Bidder is set to the initiator's module account address instead of module name. (when the first bid is placed, it is paid out to the initiator)
	// Setting to the module account address bypasses calling supply.SendCoinsFromModuleToModule, instead calls SendCoinsFromModuleToAccount.
	// This isn't a problem currently, but if additional logic/validation was added for sending to coins to Module Accounts, it would be bypassed.
	auction := DebtAuction{BaseAuction{
		// no ID
		Initiator:  buyerModAccName,
		Lot:        initialLot,
		Bidder:     supply.NewModuleAddress(buyerModAccName), // send proceeds from the first bid to the buyer.
		Bid:        bid,                                      // amount that the buyer it buying - doesn't change over course of auction
		EndTime:    EndTime,
		MaxEndTime: EndTime,
	}}
	return auction
}

// CollateralAuction is a two phase auction.
// Initially, in forward auction phase, bids can be placed up to a max bid.
// Then it switches to a reverse auction phase, where the initial amount up for auction is bidded down.
// Unsold Lot is sent to LotReturns, being divided among the addresses by weight.
type CollateralAuction struct {
	BaseAuction
	MaxBid     sdk.Coin
	LotReturns WeightedAddresses
}

// WithID returns an auction with the ID set.
func (a CollateralAuction) WithID(id uint64) Auction { a.ID = id; return a }

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
func NewCollateralAuction(seller string, lot sdk.Coin, EndTime time.Time, maxBid sdk.Coin, lotReturns WeightedAddresses) CollateralAuction {
	auction := CollateralAuction{
		BaseAuction: BaseAuction{
			// no ID
			Initiator:  seller,
			Lot:        lot,
			Bidder:     nil,
			Bid:        sdk.NewInt64Coin(maxBid.Denom, 0),
			EndTime:    EndTime,
			MaxEndTime: EndTime},
		MaxBid:     maxBid,
		LotReturns: lotReturns,
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
