package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisAuction is an interface that extends the auction interface to add functionality needed for initializing auctions from genesis.
type GenesisAuction interface {
	Auction
	GetModuleAccountCoins() sdk.Coins
	Validate() error
}

// GenesisAuctions is a slice of genesis auctions.
type GenesisAuctions []GenesisAuction

// GenesisState is auction state that must be provided at chain genesis.
type GenesisState struct {
	NextAuctionID uint64          `json:"next_auction_id" yaml:"next_auction_id"`
	Params        Params          `json:"params" yaml:"params"`
	Auctions      GenesisAuctions `json:"genesis_auctions" yaml:"genesis_auctions"`
}

// NewGenesisState returns a new genesis state object for auctions module.
func NewGenesisState(nextID uint64, ap Params, ga GenesisAuctions) GenesisState {
	return GenesisState{
		NextAuctionID: nextID,
		Params:        ap,
		Auctions:      ga,
	}
}

// DefaultGenesisState returns the default genesis state for auction module.
func DefaultGenesisState() GenesisState {
	return NewGenesisState(0, DefaultParams(), GenesisAuctions{})
}

// Equal checks whether two GenesisState structs are equivalent.
func (gs GenesisState) Equal(gs2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty.
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// Validate validates genesis inputs. It returns error if validation of any input fails.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	ids := map[uint64]bool{}
	for _, a := range gs.Auctions {

		if err := a.Validate(); err != nil {
			return fmt.Errorf("found invalid auction: %w", err)
		}

		if ids[a.GetID()] {
			return fmt.Errorf("found duplicate auction ID (%d)", a.GetID())
		}
		ids[a.GetID()] = true

		if a.GetID() >= gs.NextAuctionID {
			return fmt.Errorf("found auction ID >= the nextAuctionID (%d >= %d)", a.GetID(), gs.NextAuctionID)
		}
	}
	return nil
}
