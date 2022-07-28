package types

// NewGenesisState creates a new genesis state.
func NewGenesisState(
	params Params,
	vaultRecords VaultRecords,
	vaultShareRecords VaultShareRecords,
) GenesisState {
	return GenesisState{
		Params:            params,
		VaultRecords:      vaultRecords,
		VaultShareRecords: vaultShareRecords,
	}
}

// Validate validates the module's genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if err := gs.VaultRecords.Validate(); err != nil {
		return err
	}

	if err := gs.VaultShareRecords.Validate(); err != nil {
		return err
	}

	return nil
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		VaultRecords{},
		VaultShareRecords{},
	)
}
