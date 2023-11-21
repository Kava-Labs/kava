package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"

	tmcmd "github.com/tendermint/tendermint/cmd/cometbft/commands"
	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	tmstate "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"

	ethermintserver "github.com/evmos/ethermint/server"
)

const (
	flagShardStartBlock = "start"
	flagShardEndBlock   = "end"
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

			// enumerate all heights to prune
			pruneHeights := make([]int64, 0, latest-shardSize)
			for i := int64(1); i < startBlock; i++ {
				pruneHeights = append(pruneHeights, i)
			}

			if len(pruneHeights) > 0 {
				// prune application state
				fmt.Printf("pruning application state to height %d\n", startBlock)
				if err := multistore.PruneStores(true, pruneHeights); err != nil {
					return fmt.Errorf("failed to prune application state: %s", err)
				}
			}

			// open block store & cometbft state to manually prune blocks
			blockStore, stateStore, err := openCometBftDbs(ctx.Config)
			if err != nil {
				return fmt.Errorf("failed to open cometbft dbs: %s", err)
			}

			// get starting block of block store
			baseBlock := blockStore.Base()
			fmt.Printf("block store base: %d\n", blockStore.Base())

			// only prune if data exists, otherwise blockStore.PruneBlocks will panic
			if baseBlock < startBlock {
				// prune block store
				fmt.Printf("pruning block store from %d - %d\n", baseBlock, startBlock)
				if _, err := blockStore.PruneBlocks(startBlock); err != nil {
					return fmt.Errorf("failed to prune block store (retainHeight=%d): %s", startBlock, err)
				}

				// prune cometbft state
				fmt.Printf("pruning cometbft state from %d - %d\n", baseBlock, startBlock)
				if err := stateStore.PruneStates(baseBlock, startBlock); err != nil {
					return fmt.Errorf("failed to prune cometbft state store (%d - %d): %s", baseBlock, startBlock, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, opts.DefaultNodeHome, "The application home directory")
	cmd.Flags().Int64(flagShardStartBlock, 1, "Start block of data shard (inclusive)")
	cmd.Flags().Int64(flagShardEndBlock, 0, "End block of data shard (exclusive)")

	return cmd
}

// inspired by https://github.com/Kava-Labs/cometbft/blob/277b0853db3f67865a55aa1c54f59790b5f591be/node/node.go#L234
func openCometBftDbs(config *tmconfig.Config) (blockStore *store.BlockStore, stateStore tmstate.Store, err error) {
	dbProvider := node.DefaultDBProvider

	var blockStoreDB dbm.DB
	blockStoreDB, err = dbProvider(&node.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return
	}
	blockStore = store.NewBlockStore(blockStoreDB)

	stateDB, err := dbProvider(&node.DBContext{ID: "state", Config: config})
	if err != nil {
		return
	}

	stateStore = tmstate.NewStore(stateDB, tmstate.StoreOptions{
		DiscardABCIResponses: config.Storage.DiscardABCIResponses,
	})

	return
}
