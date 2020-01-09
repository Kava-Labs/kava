package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// Auction is an interface to several types of auction.
type Auction interface {
	GetID() uint64
	WithID(uint64) Auction
	GetEndTime() time.Time
}

// BaseAuction type shared by all Auctions
type BaseAuction struct {
	ID         uint64
	Initiator  string         // Module that starts the auction. Giving away Lot (aka seller in a forward auction). Restricted to being a module account name rather than any account.
	Lot        sdk.Coin       // Amount of coins up being given by initiator (FA - amount for sale by seller, RA - cost of good by buyer (bid))
	Bidder     sdk.AccAddress // Person who bids in the auction. Receiver of Lot. (aka buyer in forward auction, seller in RA)
	Bid        sdk.Coin       // Amount of coins being given by the bidder (FA - bid, RA - amount being sold)
	EndTime    time.Time      // Auction closing time. Triggers at the end of the block with time ≥ endTime (bids placed in that block are valid) // TODO ensure everything is consistent with this
	MaxEndTime time.Time      // Maximum closing time. Auctions can close before this but never after.
}

// GetID getter for auction ID
func (a BaseAuction) GetID() uint64 { return a.ID }

// GetEndTime getter for auction end time
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

// ForwardAuction type for forward auctions
type ForwardAuction struct {
	BaseAuction
}

// WithID returns an auction with the ID set
func (a ForwardAuction) WithID(id uint64) Auction { a.ID = id; return a }

// NewForwardAuction creates a new forward auction
func NewForwardAuction(seller string, lot sdk.Coin, bidDenom string, endTime time.Time) ForwardAuction {
	auction := ForwardAuction{BaseAuction{
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

// ReverseAuction type for reverse auctions
type ReverseAuction struct {
	BaseAuction
}

// WithID returns an auction with the ID set
func (a ReverseAuction) WithID(id uint64) Auction { a.ID = id; return a }

// NewReverseAuction creates a new reverse auction
func NewReverseAuction(buyerModAccName string, bid sdk.Coin, initialLot sdk.Coin, EndTime time.Time) ReverseAuction {
	// Note: Bidder is set to the initiator's module account address instead of module name. (when the first bid is placed, it is paid out to the initiator)
	// Setting to the module account address bypasses calling supply.SendCoinsFromModuleToModule, instead calls SendCoinsFromModuleToAccount.
	// This isn't a problem currently, but if additional logic/validation was added for sending to coins to Module Accounts, it would be bypassed.
	auction := ReverseAuction{BaseAuction{
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

// ForwardReverseAuction type for forward reverse auction
type ForwardReverseAuction struct {
	BaseAuction
	MaxBid     sdk.Coin
	LotReturns WeightedAddresses // return addresses to pay out reductions in the lot amount to. Lot is bid down during reverse phase.
}

// WithID returns an auction with the ID set
func (a ForwardReverseAuction) WithID(id uint64) Auction { a.ID = id; return a }

func (a ForwardReverseAuction) IsReversePhase() bool {
	return a.Bid.IsEqual(a.MaxBid)
}

func (a ForwardReverseAuction) String() string {
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

// NewForwardReverseAuction creates a new forward reverse auction
func NewForwardReverseAuction(seller string, lot sdk.Coin, EndTime time.Time, maxBid sdk.Coin, lotReturns WeightedAddresses) ForwardReverseAuction {
	auction := ForwardReverseAuction{
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

// WeightedAddresses type for storing an address and its associated weight
type WeightedAddresses struct {
	Addresses []sdk.AccAddress
	Weights   []sdk.Int
}

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
