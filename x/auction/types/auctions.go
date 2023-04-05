package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gogo/protobuf/proto"
)

const (
	CollateralAuctionType = "collateral"
	SurplusAuctionType    = "surplus"
	DebtAuctionType       = "debt"
	ForwardAuctionPhase   = "forward"
	ReverseAuctionPhase   = "reverse"
)

// DistantFuture is a very large time value to use as initial the ending time for auctions.
// It is not set to the max time supported. This can cause problems with time comparisons, see https://stackoverflow.com/a/32620397.
// Also amino panics when encoding times ≥ the start of year 10000.
var DistantFuture = time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)

var (
	_ Auction        = &SurplusAuction{}
	_ GenesisAuction = &SurplusAuction{}
	_ Auction        = &DebtAuction{}
	_ GenesisAuction = &DebtAuction{}
	_ Auction        = &CollateralAuction{}
	_ GenesisAuction = &CollateralAuction{}
)

// --------------- Shared auction functionality ---------------

// Auction is an interface for handling common actions on auctions.
type Auction interface {
	proto.Message

	GetID() uint64
	WithID(uint64) Auction

	GetInitiator() string
	GetLot() sdk.Coin
	GetBidder() sdk.AccAddress
	GetBid() sdk.Coin
	GetEndTime() time.Time
	GetMaxEndTime() time.Time

	GetType() string
	GetPhase() string
}

// --------------- BaseAuction ---------------

func (a BaseAuction) GetID() uint64 { return a.ID }

func (a BaseAuction) GetBid() sdk.Coin { return a.Bid }

func (a BaseAuction) GetLot() sdk.Coin { return a.Lot }

func (a BaseAuction) GetBidder() sdk.AccAddress { return a.Bidder }

func (a BaseAuction) GetInitiator() string { return a.Initiator }

func (a BaseAuction) GetEndTime() time.Time { return a.EndTime }

func (a BaseAuction) GetMaxEndTime() time.Time { return a.MaxEndTime }

// ValidateAuction verifies that the auction end time is before max end time
func ValidateAuction(a Auction) error {
	// ID can be 0 for surplus, debt and collateral auctions
	if strings.TrimSpace(a.GetInitiator()) == "" {
		return errors.New("auction initiator cannot be blank")
	}
	if !a.GetLot().IsValid() {
		return fmt.Errorf("invalid lot: %s", a.GetLot())
	}
	if !a.GetBid().IsValid() {
		return fmt.Errorf("invalid bid: %s", a.GetBid())
	}
	if a.GetEndTime().Unix() <= 0 || a.GetMaxEndTime().Unix() <= 0 {
		return errors.New("end time cannot be zero")
	}
	if a.GetEndTime().After(a.GetMaxEndTime()) {
		return fmt.Errorf("MaxEndTime < EndTime (%s < %s)", a.GetMaxEndTime(), a.GetEndTime())
	}
	return nil
}

// --------------- SurplusAuction ---------------

// NewSurplusAuction returns a new surplus auction.
func NewSurplusAuction(seller string, lot sdk.Coin, bidDenom string, endTime time.Time) SurplusAuction {
	auction := SurplusAuction{
		BaseAuction: BaseAuction{
			// No Id
			Initiator:       seller,
			Lot:             lot,
			Bidder:          nil,
			Bid:             sdk.NewInt64Coin(bidDenom, 0),
			HasReceivedBids: false, // new auctions don't have any bids
			EndTime:         endTime,
			MaxEndTime:      endTime,
		},
	}
	return auction
}

func (a SurplusAuction) WithID(id uint64) Auction {
	a.ID = id
	return Auction(&a)
}

// GetPhase returns the direction of a surplus auction, which never changes.
func (a SurplusAuction) GetPhase() string { return ForwardAuctionPhase }

// GetType returns the auction type. Used to identify auctions in event attributes.
func (a SurplusAuction) GetType() string { return SurplusAuctionType }

// GetModuleAccountCoins returns the total number of coins held in the module account for this auction.
// It is used in genesis initialize the module account correctly.
func (a SurplusAuction) GetModuleAccountCoins() sdk.Coins {
	// a.Bid is paid out on bids, so is never stored in the module account
	return sdk.NewCoins(a.Lot)
}

func (a SurplusAuction) Validate() error {
	return ValidateAuction(&a)
}

// --------------- DebtAuction ---------------

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
			Bidder:          authtypes.NewModuleAddress(buyerModAccName), // send proceeds from the first bid to the buyer.
			Bid:             bid,                                         // amount that the buyer is buying - doesn't change over course of auction
			HasReceivedBids: false,                                       // new auctions don't have any bids
			EndTime:         endTime,
			MaxEndTime:      endTime,
		},
		CorrespondingDebt: debt,
	}
	return auction
}

func (a DebtAuction) WithID(id uint64) Auction {
	a.ID = id
	return Auction(&a)
}

// GetPhase returns the direction of a debt auction, which never changes.
func (a DebtAuction) GetPhase() string { return ReverseAuctionPhase }

// GetType returns the auction type. Used to identify auctions in event attributes.
func (a DebtAuction) GetType() string { return DebtAuctionType }

// GetModuleAccountCoins returns the total number of coins held in the module account for this auction.
// It is used in genesis initialize the module account correctly.
func (a DebtAuction) GetModuleAccountCoins() sdk.Coins {
	// a.Lot is minted at auction close, so is never stored in the module account
	// a.Bid is paid out on bids, so is never stored in the module account
	return sdk.NewCoins(a.CorrespondingDebt)
}

// Validate validates the DebtAuction fields values.
func (a DebtAuction) Validate() error {
	if !a.CorrespondingDebt.IsValid() {
		return fmt.Errorf("invalid corresponding debt: %s", a.CorrespondingDebt)
	}
	return ValidateAuction(&a)
}

// --------------- CollateralAuction ---------------

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
			MaxEndTime:      endTime,
		},
		CorrespondingDebt: debt,
		MaxBid:            maxBid,
		LotReturns:        lotReturns,
	}
	return auction
}

func (a CollateralAuction) WithID(id uint64) Auction {
	a.ID = id
	return Auction(&a)
}

// GetType returns the auction type. Used to identify auctions in event attributes.
func (a CollateralAuction) GetType() string { return CollateralAuctionType }

// IsReversePhase returns whether the auction has switched over to reverse phase or not.
// CollateralAuctions initially start in forward phase.
func (a CollateralAuction) IsReversePhase() bool {
	return a.Bid.IsEqual(a.MaxBid)
}

// GetPhase returns the direction of a collateral auction.
func (a CollateralAuction) GetPhase() string {
	if a.IsReversePhase() {
		return ReverseAuctionPhase
	}
	return ForwardAuctionPhase
}

// GetLotReturns returns the auction's lot returns as weighted addresses
func (a CollateralAuction) GetLotReturns() WeightedAddresses { return a.LotReturns }

// GetModuleAccountCoins returns the total number of coins held in the module account for this auction.
// It is used in genesis initialize the module account correctly.
func (a CollateralAuction) GetModuleAccountCoins() sdk.Coins {
	// a.Bid is paid out on bids, so is never stored in the module account
	return sdk.NewCoins(a.Lot).Add(sdk.NewCoins(a.CorrespondingDebt)...)
}

// Validate validates the CollateralAuction fields values.
func (a CollateralAuction) Validate() error {
	if !a.CorrespondingDebt.IsValid() {
		return fmt.Errorf("invalid corresponding debt: %s", a.CorrespondingDebt)
	}
	if !a.MaxBid.IsValid() {
		return fmt.Errorf("invalid max bid: %s", a.MaxBid)
	}
	if err := a.LotReturns.Validate(); err != nil {
		return fmt.Errorf("invalid lot returns: %w", err)
	}
	return ValidateAuction(&a)
}

// NewWeightedAddresses returns a new list addresses with weights.
func NewWeightedAddresses(addrs []sdk.AccAddress, weights []sdkmath.Int) (WeightedAddresses, error) {
	wa := WeightedAddresses{
		Addresses: addrs,
		Weights:   weights,
	}
	if err := wa.Validate(); err != nil {
		return WeightedAddresses{}, err
	}
	return wa, nil
}

// Validate checks for that the weights are not negative, not all zero, and the lengths match.
func (wa WeightedAddresses) Validate() error {
	if len(wa.Weights) < 1 {
		return fmt.Errorf("must be at least 1 weighted address")
	}

	if len(wa.Addresses) != len(wa.Weights) {
		return fmt.Errorf("number of addresses doesn't match number of weights, %d ≠ %d", len(wa.Addresses), len(wa.Weights))
	}

	totalWeight := sdk.ZeroInt()
	for i := range wa.Addresses {
		if wa.Addresses[i].Empty() {
			return fmt.Errorf("address %d cannot be empty", i)
		}
		if wa.Weights[i].IsNegative() {
			return fmt.Errorf("weight %d contains a negative amount: %s", i, wa.Weights[i])
		}
		totalWeight = totalWeight.Add(wa.Weights[i])
	}

	if !totalWeight.IsPositive() {
		return fmt.Errorf("total weight must be positive")
	}

	return nil
}
