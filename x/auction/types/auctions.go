package types

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// ID type for auction IDs
type ID uint64

// NewIDFromString generate new auction ID from a string
func NewIDFromString(s string) (ID, error) {
	n, err := strconv.ParseUint(s, 10, 64) // copied from how the gov module rest handler's parse proposal IDs
	if err != nil {
		return 0, err
	}
	return ID(n), nil
}
func NewIDFromBytes(bz []byte) ID {
	return ID(binary.BigEndian.Uint64(bz))

}
func (id ID) Bytes() []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(id))
	return bz
}

// Auction is an interface to several types of auction.
type Auction interface {
	GetID() ID
	SetID(ID)
	GetBidder() sdk.AccAddress
	GetLot() sdk.Coin
	GetEndTime() time.Time
}

// BaseAuction type shared by all Auctions
type BaseAuction struct {
	ID         ID
	Initiator  string         // Module who starts the auction. Giving away Lot (aka seller in a forward auction). Restricted to being a module account name rather than any account.
	Lot        sdk.Coin       // Amount of coins up being given by initiator (FA - amount for sale by seller, RA - cost of good by buyer (bid))
	Bidder     sdk.AccAddress // Person who bids in the auction. Receiver of Lot. (aka buyer in forward auction, seller in RA)
	Bid        sdk.Coin       // Amount of coins being given by the bidder (FA - bid, RA - amount being sold)
	EndTime    time.Time      // Auction closing time. Triggers at the end of the block with time ≥ endTime (bids placed in that block are valid) // TODO ensure everything is consistent with this
	MaxEndTime time.Time      // Maximum closing time. Auctions can close before this but never after.
}

// GetID getter for auction ID
func (a *BaseAuction) GetID() ID { return a.ID }

// SetID setter for auction ID
func (a *BaseAuction) SetID(id ID) { a.ID = id } // TODO if this returns a new auction with ID then no pointers are needed

// GetBid getter for auction bid
func (a *BaseAuction) GetBidder() sdk.AccAddress { return a.Bidder }

// GetLot getter for auction lot
func (a *BaseAuction) GetLot() sdk.Coin { return a.Lot }

// GetEndTime getter for auction end time
func (a *BaseAuction) GetEndTime() time.Time { return a.EndTime }

func (a *BaseAuction) String() string {
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
	*BaseAuction
}

// NewForwardAuction creates a new forward auction
func NewForwardAuction(seller string, lot sdk.Coin, bidDenom string, endTime time.Time) ForwardAuction {
	auction := ForwardAuction{&BaseAuction{
		// no ID
		Initiator:  seller,
		Lot:        lot,
		Bidder:     nil, // TODO on the first place bid, 0 coins will be sent to this address, check if this causes problems or can be avoided
		Bid:        sdk.NewInt64Coin(bidDenom, 0),
		EndTime:    endTime,
		MaxEndTime: endTime,
	}}
	// output := BankOutput{seller, lot}
	return auction
}

// ReverseAuction type for reverse auctions
type ReverseAuction struct {
	*BaseAuction
}

// NewReverseAuction creates a new reverse auction
func NewReverseAuction(buyerModAccName string, bid sdk.Coin, initialLot sdk.Coin, EndTime time.Time) ReverseAuction {
	// TODO setting the bidder here is a bit hacky
	// Needs to be set so that when the first bid is placed, it is paid out to the initiator.
	// Setting to the module account address bypasses calling supply.SendCoinsFromModuleToModule, instead calls SendCoinsFromModuleToModule. Not a problem currently but if checks/logic regarding modules accounts where added to those methods they would be bypassed.
	// Alternative: set address to nil, and catch it in an if statement in place bid
	auction := ReverseAuction{&BaseAuction{
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
	*BaseAuction
	MaxBid      sdk.Coin
	OtherPerson sdk.AccAddress // TODO rename, this is normally the original CDP owner, will have to be updated to account for deposits
}

func (a *ForwardReverseAuction) String() string {
	return fmt.Sprintf(`Auction %d:
  Initiator:              %s
  Lot:               			%s
  Bidder:            		  %s
  Bid:        						%s
  End Time:   						%s
	Max End Time:      			%s
	Max Bid									%s
	Other Person						%s`,
		a.GetID(), a.Initiator, a.Lot,
		a.Bidder, a.Bid, a.GetEndTime().String(),
		a.MaxEndTime.String(), a.MaxBid, a.OtherPerson,
	)
}

// NewForwardReverseAuction creates a new forward reverse auction
func NewForwardReverseAuction(seller string, lot sdk.Coin, EndTime time.Time, maxBid sdk.Coin, otherPerson sdk.AccAddress) ForwardReverseAuction {
	auction := ForwardReverseAuction{
		BaseAuction: &BaseAuction{
			// no ID
			Initiator:  seller,
			Lot:        lot,
			Bidder:     nil, // TODO on the first place bid, 0 coins will be sent to this address, check if this causes problems or can be avoided
			Bid:        sdk.NewInt64Coin(maxBid.Denom, 0),
			EndTime:    EndTime,
			MaxEndTime: EndTime},
		MaxBid:      maxBid,
		OtherPerson: otherPerson,
	}
	//output := BankOutput{seller, lot}
	return auction
}
