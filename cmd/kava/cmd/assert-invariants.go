package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/version"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
)

func AssertInvariantsCmd(config params.EncodingConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "assert-invariants [genesis-file]",
		Short:   "Validates that the input genesis file is valid and invariants pass",
		Long:    "Reads the input genesis file into a genesis document, checks that the state is valid and asserts that all invariants pass.",
		Example: fmt.Sprintf(`%s assert-invariants /path/to/genesis.json`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			importGenesis := args[0]
			genDoc, err := validateGenDoc(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to read genesis document from file %s: %w", importGenesis, err)
			}

			tApp := app.NewTestAppFromSealed()
			var newAppState genutiltypes.AppMap
			if err := json.Unmarshal(genDoc.AppState, &newAppState); err != nil {
				return fmt.Errorf("failed to marshal app state from genesis doc: %s: %w", importGenesis, err)
			}
			err = app.ModuleBasics.ValidateGenesis(config.Marshaler, config.TxConfig, newAppState)
			if err != nil {
				return fmt.Errorf("genesis doc did not pass validate genesis: %s: %w", importGenesis, err)
			}
			tApp.InitializeFromGenesisStatesWithTimeAndChainIDAndHeight(
				genDoc.GenesisTime,
				genDoc.ChainID,
				genDoc.InitialHeight,
				false,
				app.GenesisState(newAppState),
			)

			fmt.Printf("successfully asserted all invariants for %s\n", importGenesis)
			return nil
		},
	}

	return cmd
}

// validateGenDoc reads a genesis file and validates that it is a correct
// Tendermint GenesisDoc. This function does not do any cosmos-related
// validation.
func validateGenDoc(importGenesisFile string) (*tmtypes.GenesisDoc, error) {
	genDoc, err := tmtypes.GenesisDocFromFile(importGenesisFile)
	if err != nil {
		return nil, fmt.Errorf("failed to validate CometBFT consensus params: %s", err)
	}
	return genDoc, nil
}
