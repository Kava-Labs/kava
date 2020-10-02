package migrate

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/migrate/v0_11"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	flagGenesisTime = "genesis-time"
	flagChainID     = "chain-id"
)

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd(_ *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate [genesis-file]",
		Short:   "Migrate genesis file from kava v0.10 to v0.11",
		Long:    "Migrate the source genesis into the current version, sorts it, and print to STDOUT. If not provided, chain-id is set to kava-4 and genesis time is set to 2020-10-15T:14:00:00Z",
		Example: fmt.Sprintf(`%s migrate /path/to/genesis.json --chain-id=new-chain-id --genesis-time=1998-01-01T00:00:00Z`, version.ServerName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// 1) Unmarshal existing genesis.json

			importGenesis := args[0]
			genDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to read genesis document from file %s: %w", importGenesis, err)
			}

			// 2) Migrate state from kava v0.3 to v0.8

			newGenDoc := v0_11.Migrate(*genDoc)

			// 3) Create and output a new genesis file

			genesisTime := cmd.Flag(flagGenesisTime).Value.String()
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return fmt.Errorf("failed to unmarshal genesis time: %w", err)
				}

				newGenDoc.GenesisTime = t
			}

			chainID := cmd.Flag(flagChainID).Value.String()
			if chainID != "" {
				newGenDoc.ChainID = chainID
			}

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

	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().String(flagChainID, "", "override chain_id with this flag")

	return cmd
}
