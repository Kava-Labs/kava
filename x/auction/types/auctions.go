package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// Auction is an interface to several types of auction.
type Auction interface {
	GetID() ID
	SetID(ID)
	// PlaceBid(currentBlockHeight EndTime, bidder sdk.AccAddress, lot sdk.Coin, bid sdk.Coin) ([]BankOutput, []BankInput, sdk.Error)
	GetEndTime() EndTime // auctions close at the end of the block with blockheight EndTime (ie bids placed in that block are valid)
	// GetPayout() BankInput
}

// BaseAuction type shared by all Auctions
type BaseAuction struct {
	ID         ID
	Initiator  string         // Module who starts the auction. Giving away Lot (aka seller in a forward auction). Restricted to being a module account name rather than any account.
	Lot        sdk.Coin       // Amount of coins up being given by initiator (FA - amount for sale by seller, RA - cost of good by buyer (bid))
	Bidder     sdk.AccAddress // Person who bids in the auction. Receiver of Lot. (aka buyer in forward auction, seller in RA)
	Bid        sdk.Coin       // Amount of coins being given by the bidder (FA - bid, RA - amount being sold)
	EndTime    EndTime        // Block height at which the auction closes. It closes at the end of this block // TODO change to time type
	MaxEndTime EndTime        // Maximum closing time. Auctions can close before this but never after.
}

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

// EndTime type for end time of auctions
type EndTime int64 // TODO rename to Blockheight or don't define custom type

// BankInput the input and output types from the bank module where used here. But they use sdk.Coins instad of sdk.Coin. So it caused a lot of type conversion as auction mainly uses sdk.Coin.
type BankInput struct {
	Address sdk.AccAddress
	Coin    sdk.Coin
}

// BankOutput output type for auction bids
type BankOutput struct {
	Address sdk.AccAddress
	Coin    sdk.Coin
}

// GetID getter for auction ID
func (a *BaseAuction) GetID() ID { return a.ID }

// SetID setter for auction ID
func (a *BaseAuction) SetID(id ID) { a.ID = id }

// GetBid getter for auction bid
func (a *BaseAuction) GetBid() sdk.Coin { return a.Bid }

// GetLot getter for auction lot
func (a *BaseAuction) GetLot() sdk.Coin { return a.Lot }

// GetEndTime getter for auction end time
func (a *BaseAuction) GetEndTime() EndTime { return a.EndTime }

// GetPayout implements Auction
// func (a BaseAuction) GetPayout() BankInput {
// 	return BankInput{a.Bidder, a.Lot}
// }

func (e EndTime) String() string {
	return string(e)
}

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
func NewForwardAuction(seller string, lot sdk.Coin, bidDenom string, EndTime EndTime) ForwardAuction {
	auction := ForwardAuction{&BaseAuction{
		// no ID
		Initiator:  seller,
		Lot:        lot,
		Bidder:     nil, // TODO on the first place bid, 0 coins will be sent to this address, check if this causes problems or can be avoided
		Bid:        sdk.NewInt64Coin(bidDenom, 0),
		EndTime:    EndTime,
		MaxEndTime: EndTime,
	}}
	// output := BankOutput{seller, lot}
	return auction
}

// PlaceBid implements Auction
// func (a *ForwardAuction) PlaceBid(currentBlockHeight EndTime, bidder sdk.AccAddress, lot sdk.Coin, bid sdk.Coin) ([]BankOutput, []BankInput, sdk.Error) {
// 	// TODO check lot size matches lot?
// 	// check auction has not closed
// 	if currentBlockHeight > a.EndTime {
// 		return []BankOutput{}, []BankInput{}, sdk.ErrInternal("auction has closed")
// 	}
// 	// check bid is greater than last bid
// 	if !a.Bid.IsLT(bid) { // TODO add minimum bid size
// 		return []BankOutput{}, []BankInput{}, sdk.ErrInternal("bid not greater than last bid")
// 	}
// 	// calculate coin movements
// 	outputs := []BankOutput{{bidder, bid}}                                  // new bidder pays bid now
// 	inputs := []BankInput{{a.Bidder, a.Bid}, {a.Initiator, bid.Sub(a.Bid)}} // old bidder is paid back, extra goes to seller

// 	// update auction
// 	a.Bidder = bidder
// 	a.Bid = bid
// 	// increment timeout // TODO into keeper?
// 	a.EndTime = EndTime(min(int64(currentBlockHeight+DefaultMaxBidDuration), int64(a.MaxEndTime))) // TODO is there a better way to structure these types?

// 	return outputs, inputs, nil
// }

// ReverseAuction type for reverse auctions
// TODO  when exporting state and initializing a new genesis, we'll need a way to differentiate forward from reverse auctions
type ReverseAuction struct {
	*BaseAuction
}

// NewReverseAuction creates a new reverse auction
func NewReverseAuction(buyerModAccName string, bid sdk.Coin, initialLot sdk.Coin, EndTime EndTime) ReverseAuction {
	// Bidder set here receives the proceeds from the first bid placed. This is set to the address of the module account.
	// When this happens it uses supply.SendCoinsFromModuleToAccount, rather than SendCoinsFromModuleToModule.
	// Currently not a problem but if extra checks are added to module accounts this will skip them.
	// TODO description
	auction := ReverseAuction{&BaseAuction{
		// no ID
		Initiator:  buyerModAccName,
		Lot:        initialLot,
		Bidder:     supply.NewModuleAddress(buyerModAccName), // send proceeds from the first bid to the buyer.
		Bid:        bid,                                      // amount that the buyer it buying - doesn't change over course of auction
		EndTime:    EndTime,
		MaxEndTime: EndTime,
	}}
	//output := BankOutput{buyer, initialLot}
	return auction
}

// PlaceBid implements Auction
// func (a *ReverseAuction) PlaceBid(currentBlockHeight EndTime, bidder sdk.AccAddress, lot sdk.Coin, bid sdk.Coin) ([]BankOutput, []BankInput, sdk.Error) {

// 	// check bid size matches bid?
// 	// check auction has not closed
// 	if currentBlockHeight > a.EndTime {
// 		return []BankOutput{}, []BankInput{}, sdk.ErrInternal("auction has closed")
// 	}
// 	// check bid is less than last bid
// 	if !lot.IsLT(a.Lot) { // TODO add min bid decrements
// 		return []BankOutput{}, []BankInput{}, sdk.ErrInternal("lot not smaller than last lot")
// 	}
// 	// calculate coin movements
// 	outputs := []BankOutput{{bidder, a.Bid}}                                // new bidder pays bid now
// 	inputs := []BankInput{{a.Bidder, a.Bid}, {a.Initiator, a.Lot.Sub(lot)}} // old bidder is paid back, decrease in price for goes to buyer

// 	// update auction
// 	a.Bidder = bidder
// 	a.Lot = lot
// 	// increment timeout // TODO into keeper?
// 	a.EndTime = EndTime(min(int64(currentBlockHeight+DefaultMaxBidDuration), int64(a.MaxEndTime))) // TODO is there a better way to structure these types?

// 	return outputs, inputs, nil
// }

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
func NewForwardReverseAuction(seller string, lot sdk.Coin, EndTime EndTime, maxBid sdk.Coin, otherPerson sdk.AccAddress) ForwardReverseAuction {
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

// PlaceBid implements auction
// func (a *ForwardReverseAuction) PlaceBid(currentBlockHeight EndTime, bidder sdk.AccAddress, lot sdk.Coin, bid sdk.Coin) (outputs []BankOutput, inputs []BankInput, err sdk.Error) {
// 	// check auction has not closed
// 	if currentBlockHeight > a.EndTime {
// 		return []BankOutput{}, []BankInput{}, sdk.ErrInternal("auction has closed")
// 	}

// 	// determine phase of auction
// 	switch {
// 	case a.Bid.IsLT(a.MaxBid) && bid.IsLT(a.MaxBid):
// 		// Forward auction phase
// 		if !a.Bid.IsLT(bid) { // TODO add min bid increments
// 			return []BankOutput{}, []BankInput{}, sdk.ErrInternal("bid not greater than last bid")
// 		}
// 		outputs = []BankOutput{{bidder, bid}}                                  // new bidder pays bid now
// 		inputs = []BankInput{{a.Bidder, a.Bid}, {a.Initiator, bid.Sub(a.Bid)}} // old bidder is paid back, extra goes to seller
// 	case a.Bid.IsLT(a.MaxBid):
// 		// Switch over phase
// 		if !bid.IsEqual(a.MaxBid) { // require bid == a.MaxBid
// 			return []BankOutput{}, []BankInput{}, sdk.ErrInternal("bid greater than the max bid")
// 		}
// 		outputs = []BankOutput{{bidder, bid}} // new bidder pays bid now
// 		inputs = []BankInput{
// 			{a.Bidder, a.Bid},               // old bidder is paid back
// 			{a.Initiator, bid.Sub(a.Bid)},   // extra goes to seller
// 			{a.OtherPerson, a.Lot.Sub(lot)}, //decrease in price for goes to original CDP owner
// 		}

// 	case a.Bid.IsEqual(a.MaxBid):
// 		// Reverse auction phase
// 		if !lot.IsLT(a.Lot) { // TODO add min bid decrements
// 			return []BankOutput{}, []BankInput{}, sdk.ErrInternal("lot not smaller than last lot")
// 		}
// 		outputs = []BankOutput{{bidder, a.Bid}}                                  // new bidder pays bid now
// 		inputs = []BankInput{{a.Bidder, a.Bid}, {a.OtherPerson, a.Lot.Sub(lot)}} // old bidder is paid back, decrease in price for goes to original CDP owner
// 	default:
// 		panic("should never be reached") // TODO
// 	}

// 	// update auction
// 	a.Bidder = bidder
// 	a.Lot = lot
// 	a.Bid = bid
// 	// increment timeout
// 	// TODO use bid duration param
// 	a.EndTime = EndTime(min(int64(currentBlockHeight+DefaultMaxBidDuration), int64(a.MaxEndTime))) // TODO is there a better way to structure these types?

// 	return outputs, inputs, nil
// }
