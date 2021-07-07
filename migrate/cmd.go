package migrate

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

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
