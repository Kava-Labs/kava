package types

import (
	"bytes"
)

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	NextHTLTID uint64 `json:"next_htlt_id" yaml:"next_htlt_id"`
	Params     Params `json:"params" yaml:"params"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(nextID uint64, params Params) GenesisState {
	return GenesisState{
		NextHTLTID: nextID,
		Params:     params,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return NewGenesisState(0, DefaultParams())

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

	// TODO: Allow initial HTLTs upon genesis?
	// ids := map[uint64]bool{}
	// for _, a := range gs.HTLTs {

	// 	if err := a.Validate(); err != nil {
	// 		return fmt.Errorf("found invalid HTLT: %w", err)
	// 	}

	// 	if ids[a.GetID()] {
	// 		return fmt.Errorf("found duplicate HTLT ID (%d)", a.GetID())
	// 	}
	// 	ids[a.GetID()] = true

	// 	if a.GetID() >= gs.NextHTLTID {
	// 		return fmt.Errorf("found HTLT ID >= the nextHTLDID (%d >= %d)", a.GetID(), gs.NextHTLTID)
	// 	}
	// }
	return nil
}
