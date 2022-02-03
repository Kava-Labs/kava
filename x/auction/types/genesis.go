package types

import (
	"fmt"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultNextAuctionID is the starting point for auction IDs.
const DefaultNextAuctionID uint64 = 1

// GenesisAuction extends the auction interface to add functionality
// needed for initializing auctions from genesis.
type GenesisAuction interface {
	Auction
	GetModuleAccountCoins() sdk.Coins
	Validate() error
}

// PackGenesisAuctions converts a GenesisAuction slice to Any slice
func PackGenesisAuctions(ga []GenesisAuction) ([]*types.Any, error) {
	gaAny := make([]*types.Any, len(ga))
	for i, genesisAuction := range ga {
		any, err := types.NewAnyWithValue(genesisAuction)
		if err != nil {
			return nil, err
		}
		gaAny[i] = any
	}

	return gaAny, nil
}

func mustPackGenesisAuctions(ga []GenesisAuction) []*types.Any {
	anys, err := PackGenesisAuctions(ga)
	if err != nil {
		panic(err)
	}
	return anys
}

// UnpackGenesisAuctions converts Any slice to GenesisAuctions slice
func UnpackGenesisAuctions(genesisAuctionsAny []*types.Any) ([]GenesisAuction, error) {
	genesisAuctions := make([]GenesisAuction, len(genesisAuctionsAny))
	for i, any := range genesisAuctionsAny {
		genesisAuction, ok := any.GetCachedValue().(GenesisAuction)
		if !ok {
			return nil, fmt.Errorf("expected genesis auction")
		}
		genesisAuctions[i] = genesisAuction
	}

	return genesisAuctions, nil
}

// Ensure this type will unpack contained interface types correctly when it is unmarshalled.
var _ types.UnpackInterfacesMessage = &GenesisState{}

// NewGenesisState returns a new genesis state object for auctions module.
func NewGenesisState(nextID uint64, ap Params, ga []GenesisAuction) (*GenesisState, error) {
	packedGA, err := PackGenesisAuctions(ga)
	if err != nil {
		return &GenesisState{}, err
	}

	return &GenesisState{
		NextAuctionId: nextID,
		Params:        ap,
		Auctions:      packedGA,
	}, nil
}

// DefaultGenesisState returns the default genesis state for auction module.
func DefaultGenesisState() *GenesisState {
	genesis, err := NewGenesisState(
		DefaultNextAuctionID,
		DefaultParams(),
		[]GenesisAuction{},
	)
	if err != nil {
		panic(fmt.Sprintf("could not create default genesis state: %v", err))
	}
	return genesis
}

// Validate validates genesis inputs. It returns error if validation of any input fails.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	auctions, err := UnpackGenesisAuctions(gs.Auctions)
	if err != nil {
		return err
	}

	ids := map[uint64]bool{}
	for _, a := range auctions {

		if err := a.Validate(); err != nil {
			return fmt.Errorf("found invalid auction: %w", err)
		}

		if ids[a.GetID()] {
			return fmt.Errorf("found duplicate auction ID (%d)", a.GetID())
		}
		ids[a.GetID()] = true

		if a.GetID() >= gs.NextAuctionId {
			return fmt.Errorf("found auction ID ≥ the nextAuctionID (%d ≥ %d)", a.GetID(), gs.NextAuctionId)
		}
	}
	return nil
}

// UnpackInterfaces hooks into unmarshalling to unpack any interface types contained within the GenesisState.
func (gs GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, any := range gs.Auctions {
		var auction GenesisAuction
		err := unpacker.UnpackAny(any, &auction)
		if err != nil {
			return err
		}
	}
	return nil
}
