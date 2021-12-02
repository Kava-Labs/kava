package v0_15

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is auction state that must be provided at chain genesis.
type GenesisState struct {
	NextAuctionID uint64          `json:"next_auction_id" yaml:"next_auction_id"`
	Params        Params          `json:"params" yaml:"params"`
	Auctions      GenesisAuctions `json:"auctions" yaml:"auctions"`
}

// GenesisAuctions is a slice of genesis auctions.
type GenesisAuctions []GenesisAuction

// GenesisAuction is an interface that extends the auction interface to add functionality needed for initializing auctions from genesis.
type GenesisAuction interface {
	Auction
	GetModuleAccountCoins() sdk.Coins
	Validate() error
}

// Params is the governance parameters for the auction module.
type Params struct {
	MaxAuctionDuration  time.Duration `json:"max_auction_duration" yaml:"max_auction_duration"` // max length of auction
	BidDuration         time.Duration `json:"bid_duration" yaml:"bid_duration"`                 // additional time added to the auction end time after each bid, capped by the expiry.
	IncrementSurplus    sdk.Dec       `json:"increment_surplus" yaml:"increment_surplus"`       // percentage change (of auc.Bid) required for a new bid on a surplus auction
	IncrementDebt       sdk.Dec       `json:"increment_debt" yaml:"increment_debt"`             // percentage change (of auc.Lot) required for a new bid on a debt auction
	IncrementCollateral sdk.Dec       `json:"increment_collateral" yaml:"increment_collateral"` // percentage change (of auc.Bid or auc.Lot) required for a new bid on a collateral auction
}

// RegisterCodec registers concrete types on the codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*GenesisAuction)(nil), nil)
	cdc.RegisterInterface((*Auction)(nil), nil)
	cdc.RegisterConcrete(SurplusAuction{}, "auction/SurplusAuction", nil)
	cdc.RegisterConcrete(DebtAuction{}, "auction/DebtAuction", nil)
	cdc.RegisterConcrete(CollateralAuction{}, "auction/CollateralAuction", nil)
}
