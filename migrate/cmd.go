package migrate

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/migrate/v0_16"
)

type (
	// MigrationCallback converts a genesis doc from the previous version to the
	// targeted one.
	MigrationCallback func(*tmtypes.GenesisDoc, client.Context) (*tmtypes.GenesisDoc, error)

	// MigrationMap defines a mapping from a version to a MigrationCallback.
	MigrationMap map[string]MigrationCallback
)

var migrationMap = MigrationMap{
	// TODO: Upgrade v15 migrations and add it here to support it,
	"v0.16": v0_16.Migrate,
}

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate [target-version] [genesis-file]",
		Short:   "Migrate genesis to a specified target version (v0.15 or v0.16)",
		Long:    "Migrate the source genesis into the target version and print to STDOUT.",
		Example: fmt.Sprintf(`%s migrate v0.16 /path/to/genesis.json`, version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			target := args[0]
			importGenesis := args[1]

			oldGenDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to read genesis document from file %s: %w", importGenesis, err)
			}

			migrationFunc := migrationMap[target]
			if migrationFunc == nil {
				return fmt.Errorf("unknown migration function for version: %s", target)
			}

			newGenDoc, err := migrationFunc(oldGenDoc, clientCtx)
			if err != nil {
				return fmt.Errorf("failed to run migration: %w", err)
			}

			bz, err := tmjson.Marshal(newGenDoc)
			if err != nil {
				return fmt.Errorf("failed to marshal genesis doc: %w", err)
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return fmt.Errorf("failed to sort JSON genesis doc: %w", err)
			}

			fmt.Println(string(sortedBz))
			return nil
		},
	}

	return cmd
}

func AssertInvariantsCmd() *cobra.Command {
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
				return fmt.Errorf("failed to marchal app state from genesis doc: %s: %w", importGenesis, err)
			}
			encodingConfig := app.MakeEncodingConfig()
			err = app.ModuleBasics.ValidateGenesis(encodingConfig.Marshaler, encodingConfig.TxConfig, newAppState)
			if err != nil {
				return fmt.Errorf("genesis doc did not pass validate genesis: %s: %w", importGenesis, err)
			}
			tApp.InitializeFromGenesisStatesWithTimeAndChainID(genDoc.GenesisTime, genDoc.ChainID, app.GenesisState(newAppState))

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
		return nil, fmt.Errorf("%s. Make sure that"+
			" you have correctly migrated all Tendermint consensus params, please see the"+
			" chain migration guide at https://docs.cosmos.network/master/migrations/chain-upgrade-guide-040.html for more info",
			err.Error(),
		)
	}

	return genDoc, nil
}
