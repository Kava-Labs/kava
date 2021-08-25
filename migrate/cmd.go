package migrate

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/migrate/v0_15"
)

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd(_ *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate [genesis-file]",
		Short:   "Migrate genesis file from kava v0.14 to v0.15",
		Long:    "Migrate the source genesis into the current version, sorts it, and print to STDOUT.",
		Example: fmt.Sprintf(`%s migrate /path/to/genesis.json`, version.ServerName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			importGenesis := args[0]
			genDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to read genesis document from file %s: %w", importGenesis, err)
			}

			newGenDoc := v0_15.Migrate(*genDoc)

			bz, err := cdc.MarshalJSONIndent(newGenDoc, "", "  ")
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

func AssertInvariantsCmd(_ *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "assert-invariants [genesis-file]",
		Short:   "Validates that the input genesis file is valid and invariants pass",
		Long:    "Reads the input genesis file into a genesis document, checks that the state is valid and asserts that all invariants pass.",
		Example: fmt.Sprintf(`%s assert-invariants /path/to/genesis.json`, version.ServerName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			importGenesis := args[0]
			genDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to read genesis document from file %s: %w", importGenesis, err)
			}
			tApp := app.NewTestAppFromSealed()
			var newAppState genutil.AppMap
			cdc := app.MakeCodec()
			err = cdc.UnmarshalJSON(genDoc.AppState, &newAppState)
			if err != nil {
				return fmt.Errorf("failed to marchal app state from genesis doc: %s: %w", importGenesis, err)
			}
			err = app.ModuleBasics.ValidateGenesis(newAppState)
			if err != nil {
				return fmt.Errorf("genesis doc did not pass validate genesis: %s: %w", importGenesis, err)
			}
			tApp.InitializeFromGenesisStatesWithTimeAndChainID(genDoc.GenesisTime, genDoc.ChainID, app.GenesisState(newAppState))

			fmt.Printf("asserted all invariants for %s\n", importGenesis)
			return nil
		},
	}

	return cmd
}
