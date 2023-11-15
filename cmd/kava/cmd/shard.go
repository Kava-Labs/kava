package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"

	ethermintserver "github.com/evmos/ethermint/server"
)

const (
	flagShardStartBlock = "start"
	flagShardEndBlock   = "end"
	flagShardOutput     = "out"
)

func newShardCmd(startOpts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use: "shard",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			startBlock, err := cmd.Flags().GetInt64(flagShardStartBlock)
			if err != nil {
				return err
			}
			endBlock, err := cmd.Flags().GetInt64(flagShardEndBlock)
			if err != nil {
				return err
			}
			if endBlock == 0 || endBlock <= startBlock {
				return fmt.Errorf("end block (%d) must be greater than start block (%d)", endBlock, startBlock)
			}

			fmt.Printf("home: %s\nstart: %d\nend: %d\n", clientCtx.HomeDir, startBlock, endBlock)

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, startOpts.DefaultNodeHome, "The application home directory")
	cmd.Flags().Int64(flagShardStartBlock, 1, "Start block of data shard (inclusive)")
	cmd.Flags().Int64(flagShardEndBlock, 0, "End block of data shard (exclusive)")

	return cmd
}
