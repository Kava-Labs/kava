package kava3

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

const (
	defaultChainID     = "kava-3"
	defaultGenesisTime = "2020-06-01T14:00:00Z"
	flagGenesisTime    = "genesis-time"
	flagChainID        = "chain-id"
)

// WriteGenesisParamsCmd returns a command to write suggested kava-3 params to a genesis file.
func WriteGenesisParamsCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "write-params [genesis-file]",
		Short:   "Write suggested  params to a genesis file",
		Long:    "Write suggested module parameters to a gensis file, sort it, and print to STDOUT.",
		Example: fmt.Sprintf(`%s write-params /path/to/genesis.json`, version.ServerName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// Unmarshal existing genesis.json

			importGenesis := args[0]
			genDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to read genesis doc from file %s: %w", importGenesis, err)
			}

			// Unmarshal flags

			chainID := cmd.Flag(flagChainID).Value.String()
			genesisTime := cmd.Flag(flagGenesisTime).Value.String()
			var parsedGenesisTime time.Time
			if err := parsedGenesisTime.UnmarshalText([]byte(genesisTime)); err != nil {
				return fmt.Errorf("failed to unmarshal genesis time: %w", err)
			}

			// Write new params to the genesis file

			newGenDoc, err := AddSuggestedParams(cdc, *genDoc, chainID, parsedGenesisTime)
			if err != nil {
				return fmt.Errorf("failed to write params: %w", err)
			}

			// Marshal output a new genesis file

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

	cmd.Flags().String(flagGenesisTime, defaultGenesisTime, "override genesis time")
	cmd.Flags().String(flagChainID, defaultChainID, "override chain-id")

	return cmd
}
