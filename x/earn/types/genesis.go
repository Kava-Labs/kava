package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

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

func Kava11GenesisState(accountKeeper AccountKeeper) GenesisState {
	return GenesisState{
		Params: NewParams(AllowedVaults{
			NewAllowedVault(
				"usdx",
				StrategyTypes{STRATEGY_TYPE_HARD},
				false,
				nil,
			),
			NewAllowedVault(
				"bkava",
				StrategyTypes{STRATEGY_TYPE_SAVINGS},
				false,
				nil,
			),
			NewAllowedVault(
				"ukava",
				StrategyTypes{STRATEGY_TYPE_SAVINGS},
				true,
				[]sdk.AccAddress{
					accountKeeper.GetModuleAddress(distributiontypes.ModuleName),
				},
			),
		}),
	}
}
