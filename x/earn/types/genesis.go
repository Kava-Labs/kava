package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
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
		Params{
			AllowedVaults: AllowedVaults{
				// ukava - Community Pool
				NewAllowedVault(
					"ukava",
					StrategyTypes{STRATEGY_TYPE_SAVINGS},
					true,
					[]sdk.AccAddress{authtypes.NewModuleAddress(kavadisttypes.FundModuleAccount)},
				),
				NewAllowedVault(
					"bkava",
					StrategyTypes{STRATEGY_TYPE_SAVINGS},
					false,
					[]sdk.AccAddress{},
				),
				NewAllowedVault(
					"erc20/multichain/usdc",
					StrategyTypes{STRATEGY_TYPE_SAVINGS},
					false,
					[]sdk.AccAddress{},
				),
			},
		},
		VaultRecords{},
		VaultShareRecords{},
	)
}
