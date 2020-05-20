package v0_8

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultNextAuctionID uint64 = 1
	// ModuleName The name that will be used throughout the module
	ModuleName = "auction"
)

type (
	GenesisState struct {
		NextAuctionID uint64          `json:"next_auction_id" yaml:"next_auction_id"`
		Params        Params          `json:"params" yaml:"params"`
		Auctions      GenesisAuctions `json:"auctions" yaml:"auctions"`
	}

	// Params is the governance parameters for the auction module.
	Params struct {
		MaxAuctionDuration  time.Duration `json:"max_auction_duration" yaml:"max_auction_duration"` // max length of auction
		BidDuration         time.Duration `json:"bid_duration" yaml:"bid_duration"`                 // additional time added to the auction end time after each bid, capped by the expiry.
		IncrementSurplus    sdk.Dec       `json:"increment_surplus" yaml:"increment_surplus"`       // percentage change (of auc.Bid) required for a new bid on a surplus auction
		IncrementDebt       sdk.Dec       `json:"increment_debt" yaml:"increment_debt"`             // percentage change (of auc.Lot) required for a new bid on a debt auction
		IncrementCollateral sdk.Dec       `json:"increment_collateral" yaml:"increment_collateral"` // percentage change (of auc.Bid or auc.Lot) required for a new bid on a collateral auction
	}

	// GenesisAuction is an interface that extends the auction interface to add functionality needed for initializing auctions from genesis.
	GenesisAuction interface {
		Auction
		GetModuleAccountCoins() sdk.Coins
		Validate() error
	}

	// GenesisAuctions is a slice of genesis auctions.
	GenesisAuctions []GenesisAuction

	// Auction is an interface for handling common actions on auctions.
	Auction interface {
		GetID() uint64
		WithID(uint64) Auction

		GetInitiator() string
		GetLot() sdk.Coin
		GetBidder() sdk.AccAddress
		GetBid() sdk.Coin
		GetEndTime() time.Time

		GetType() string
		GetPhase() string
	}
)

func EmptyGenesisState() GenesisState {
	return GenesisState{
		NextAuctionID: DefaultNextAuctionID,
		Params:        Params{}, // TODO how should we set these params, should it be part of migration?
		Auctions:      GenesisAuctions{},
	}
}
