package types

// NewGenesisState returns a new genesis state object
func NewGenesisState(params Params, stakingRewardsState StakingRewardsState) GenesisState {
	return GenesisState{
		Params:              params,
		StakingRewardsState: stakingRewardsState,
	}
}

// DefaultGenesisState returns default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		DefaultStakingRewardsState(),
	)
}

// Validate checks the params are valid
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return gs.StakingRewardsState.Validate()
}
