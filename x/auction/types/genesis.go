package types

import (
	"bytes"
)

// Auctions type for an array of auctions
type Auctions []Auction

// GenesisState - auction state that must be provided at genesis
type GenesisState struct {
	NextAuctionID uint64          `json:"next_auction_id" yaml:"next_auction_id"`
	Params        Params          `json:"auction_params" yaml:"auction_params"`
	Auctions      Auctions `json:"genesis_auctions" yaml:"genesis_auctions"`
}

// NewGenesisState returns a new genesis state object for auctions module
func NewGenesisState(nextID uint64, ap Params, ga Auctions) GenesisState {
	return GenesisState{
		NextAuctionID: nextID,
		Params:        ap,
		Auctions:      ga,
	}
}

// DefaultGenesisState defines default genesis state for auction module
func DefaultGenesisState() GenesisState {
	return NewGenesisState(0, DefaultParams(), Auctions{})
}

// Equal checks whether two GenesisState structs are equivalent
func (data GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(data)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (data GenesisState) IsEmpty() bool {
	return data.Equal(GenesisState{})
}

// ValidateGenesis validates genesis inputs. Returns error if validation of any input fails.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return nil
}
