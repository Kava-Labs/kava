package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	tmcmd "github.com/tendermint/tendermint/cmd/cometbft/commands"

	ethermintserver "github.com/evmos/ethermint/server"
)

const (
	flagShardStartBlock = "start"
	flagShardEndBlock   = "end"
	flagShardOutput     = "out"
)

func newShardCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use: "shard",
		RunE: func(cmd *cobra.Command, args []string) error {
			// parse flags
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
			shardSize := endBlock - startBlock

			clientCtx := client.GetClientContextFromCmd(cmd)

			ctx := server.GetServerContextFromCmd(cmd)
			ctx.Config.SetRoot(clientCtx.HomeDir)

			home := ctx.Viper.GetString(flags.FlagHome)

			// connect to database
			db, err := opts.DBOpener(ctx.Viper, home, server.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}

			// close db connection when done
			defer func() {
				if err := db.Close(); err != nil {
					ctx.Logger.Error("error closing db", "error", err.Error())
				}
			}()

			// get the multistore
			app := opts.AppCreator(ctx.Logger, db, nil, ctx.Viper)
			cms := app.CommitMultiStore()
			multistore, ok := cms.(*rootmulti.Store)
			if !ok {
				return fmt.Errorf("only sharding of rootmulti.Store type is supported")
			}

			latest := multistore.LatestVersion()
			fmt.Printf("latest height: %d\n", latest)

			fmt.Printf("pruning data in %s down to heights %d - %d (%d blocks)\n", home, startBlock, endBlock, shardSize)

			// set pruning options to prevent no-ops from `PruneStores`
			multistore.SetPruning(pruningtypes.PruningOptions{KeepRecent: uint64(shardSize), Interval: 0})

			// rollback application state
			fmt.Printf("rolling back application state to height %d\n", endBlock-1)
			if err = multistore.RollbackToVersion(endBlock - 1); err != nil {
				return fmt.Errorf("failed to rollback application state: %s", err)
			}

			// rollback tendermint db
			height := latest
			for height >= endBlock {
				fmt.Printf("rolling back state for height %d\n", height)
				height, _, err = tmcmd.RollbackState(ctx.Config, true)
				if err != nil {
					return fmt.Errorf("failed to rollback tendermint state: %w", err)
				}
				fmt.Printf("successfully rolled back to height %d\n", height)
			}

			// prune all blocks from before start of shard
			pruneHeights := make([]int64, 0, latest-shardSize)
			// prune heights before start block
			for i := int64(1); i < startBlock; i++ {
				pruneHeights = append(pruneHeights, i)
			}

			fmt.Printf("pruning application heights: %+v\n", pruneHeights)
			if len(pruneHeights) > 0 {
				if err := multistore.PruneStores(false, pruneHeights); err != nil {
					return fmt.Errorf("failed to prune application state: %s", err)
				}
			}

			fmt.Printf("post-prune height: %d\n", multistore.LatestVersion())

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, opts.DefaultNodeHome, "The application home directory")
	cmd.Flags().Int64(flagShardStartBlock, 1, "Start block of data shard (inclusive)")
	cmd.Flags().Int64(flagShardEndBlock, 0, "End block of data shard (exclusive)")

	return cmd
}
