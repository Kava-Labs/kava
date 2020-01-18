package types

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	NextHTLTID uint64 `json:"next_htlt_id" yaml:"next_htlt_id"`
	Params     Params `json:"params" yaml:"params"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(nextID uint64, ap Params) GenesisState {
	return GenesisState{
		NextAuctionID: nextID,
		Params:        params,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return NewGenesisState(0, DefaultParams())

}

// IsEmpty returns true if a GenesisState is empty.
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// Validate validates genesis inputs. It returns error if validation of any input fails.
func (gs GenesisState) Validate() error {
	// TODO: validate nextHTLTID
	if err := gs.Params.Validate(); err != nil {
		return err
	}
}
