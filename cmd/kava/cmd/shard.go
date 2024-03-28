package cmd

import (
	"fmt"
	"strings"

	"github.com/kava-labs/kava/app"
	"github.com/spf13/cobra"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"

	tmconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	tmstate "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"

	ethermintserver "github.com/evmos/ethermint/server"
)

const (
	flagShardStartBlock        = "start"
	flagShardEndBlock          = "end"
	flagShardOnlyAppState      = "only-app-state"
	flagShardForceAppVersion   = "force-app-version"
	flagShardOnlyCometbftState = "only-cometbft-state"
	// TODO: --preserve flag for creating & operating on a copy?

	// allow using -1 to mean "latest" (perform no rollbacks)
	shardEndBlockLatest = -1
)

func newShardCmd(opts ethermintserver.StartOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shard --home <path-to-home-dir> --start <start-block> --end <end-block> [--only-app-state] [--only-cometbft-state] [--force-app-version <app-version>]",
		Short: "Strip all blocks from the database outside of a given range",
		Long: `shard opens a local kava home directory's databases and removes all blocks outside a range defined by --start and --end. The range is inclusive of the end block.

It works by first rolling back the latest state to the block before the end block, and then by pruning all state before the start block.

Setting the end block to -1 signals to keep the latest block (no rollbacks).

The application.db can be loaded at a particular height via the --force-app-version option. This is useful if the sharding process is prematurely terminated while the application.db is being sharded.

The --only-app-state flag can be used to skip the pruning of the blockstore and cometbft state. This matches the functionality of the cosmos-sdk's "prune" command. Note that rolled back blocks will still affect all stores.

Similarly, the --only-cometbft-state flag skips pruning app state. This can be useful if the shard command is prematurely terminated during the shard process.

The shard command only flags the iavl tree nodes for deletion. Actual removal from the databases will be performed when each database is compacted.

WARNING: this is a destructive action.`,
		Example: `Create a 1M block data shard (keeps blocks kava 1,000,000 to 2,000,000)
$ kava shard --home path/to/.kava --start 1000000 --end 2000000

Prune all blocks up to 5,000,000:
$ kava shard --home path/to/.kava --start 5000000 --end -1

Prune first 1M blocks _without_ affecting blockstore or cometBFT state:
$ kava shard --home path/to/.kava --start 1000000 --end -1 --only-app-state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			//////////////////////////
			// parse & validate flags
			//////////////////////////
			startBlock, err := cmd.Flags().GetInt64(flagShardStartBlock)
			if err != nil {
				return err
			}
			endBlock, err := cmd.Flags().GetInt64(flagShardEndBlock)
			if err != nil {
				return err
			}
			if (endBlock == 0 || endBlock < startBlock) && endBlock != shardEndBlockLatest {
				return fmt.Errorf("end block (%d) must be greater than start block (%d)", endBlock, startBlock)
			}
			onlyAppState, err := cmd.Flags().GetBool(flagShardOnlyAppState)
			if err != nil {
				return err
			}
			forceAppVersion, err := cmd.Flags().GetInt64(flagShardForceAppVersion)
			if err != nil {
				return err
			}
			onlyCometbftState, err := cmd.Flags().GetBool(flagShardOnlyCometbftState)
			if err != nil {
				return err
			}

			clientCtx := client.GetClientContextFromCmd(cmd)

			ctx := server.GetServerContextFromCmd(cmd)
			ctx.Config.SetRoot(clientCtx.HomeDir)

			////////////////////////
			// manage db connection
			////////////////////////
			// connect to database
			db, err := opts.DBOpener(ctx.Viper, clientCtx.HomeDir, server.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}

			// close db connection when done
			defer func() {
				if err := db.Close(); err != nil {
					ctx.Logger.Error("error closing db", "error", err.Error())
				}
			}()

			///////////////////
			// load multistore
			///////////////////
			// create app in order to load the multistore
			// skip loading the latest version so the desired height can be manually loaded
			ctx.Viper.Set("skip-load-latest", true)

			app := opts.AppCreator(ctx.Logger, db, nil, ctx.Viper).(*app.App)
			if forceAppVersion == shardEndBlockLatest {
				if err := app.LoadLatestVersion(); err != nil {
					return err
				}
			} else {
				if err := app.LoadVersion(forceAppVersion); err != nil {
					return err
				}
			}
			// get the multistore
			cms := app.CommitMultiStore()
			multistore, ok := cms.(*rootmulti.Store)
			if !ok {
				return fmt.Errorf("only sharding of rootmulti.Store type is supported")
			}

			////////////////////////
			// shard application.db
			////////////////////////
			if !onlyCometbftState {
				if err := shardApplicationDb(multistore, startBlock, endBlock); err != nil {
					return err
				}
			} else {
				fmt.Printf("[%s] skipping sharding of application.db\n", flagShardOnlyCometbftState)
			}

			//////////////////////////////////
			// shard blockstore.db & state.db
			//////////////////////////////////
			// open block store & cometbft state
			blockStore, stateStore, err := openCometBftDbs(ctx.Config)
			if err != nil {
				return fmt.Errorf("failed to open cometbft dbs: %s", err)
			}

			if !onlyAppState {
				if err := shardCometBftDbs(blockStore, stateStore, startBlock, endBlock); err != nil {
					return err
				}
			} else {
				fmt.Printf("[%s] skipping sharding of blockstore.db and state.db\n", flagShardOnlyAppState)
				fmt.Printf("blockstore contains blocks %d - %d\n", blockStore.Base(), blockStore.Height())
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, opts.DefaultNodeHome, "The application home directory")
	cmd.Flags().Int64(flagShardStartBlock, 1, "Start block of data shard (inclusive)")
	cmd.Flags().Int64(flagShardEndBlock, 0, "End block of data shard (inclusive)")
	cmd.Flags().Bool(flagShardOnlyAppState, false, "Skip pruning of blockstore & cometbft state")
	cmd.Flags().Bool(flagShardOnlyCometbftState, false, "Skip pruning of application state")
	cmd.Flags().Int64(flagShardForceAppVersion, shardEndBlockLatest, "Instead of loading latest, force set the version of the multistore that is loaded")

	return cmd
}

// shardApplicationDb prunes the multistore up to startBlock and rolls it back to endBlock
func shardApplicationDb(multistore *rootmulti.Store, startBlock, endBlock int64) error {
	//////////////////////////////
	// Rollback state to endBlock
	//////////////////////////////
	// handle desired endblock being latest
	latest := multistore.LastCommitID().Version
	if latest == 0 {
		return fmt.Errorf("failed to find latest height >0")
	}
	fmt.Printf("latest height: %d\n", latest)
	if endBlock == shardEndBlockLatest {
		endBlock = latest
	}
	shardSize := endBlock - startBlock + 1

	// error if requesting block range the database does not have
	if endBlock > latest {
		return fmt.Errorf("data does not contain end block (%d): latest version is %d", endBlock, latest)
	}

	fmt.Printf("pruning data down to heights %d - %d (%d blocks)\n", startBlock, endBlock, shardSize)

	// set pruning options to prevent no-ops from `PruneStores`
	multistore.SetPruning(pruningtypes.PruningOptions{KeepRecent: uint64(shardSize), Interval: 0})

	// rollback application state
	if err := multistore.RollbackToVersion(endBlock); err != nil {
		return fmt.Errorf("failed to rollback application state: %s", err)
	}

	//////////////////////////////
	// Prune blocks to startBlock
	//////////////////////////////
	// enumerate all heights to prune
	pruneHeights := make([]int64, 0, latest-shardSize)
	for i := int64(1); i < startBlock; i++ {
		pruneHeights = append(pruneHeights, i)
	}

	if len(pruneHeights) > 0 {
		fmt.Printf("pruning application state to height %d\n", startBlock)
		if err := multistore.PruneStores(true, pruneHeights); err != nil {
			return fmt.Errorf("failed to prune application state: %s", err)
		}
	}

	return nil
}

// shardCometBftDbs shrinks blockstore.db & state.db down to the desired block range
func shardCometBftDbs(blockStore *store.BlockStore, stateStore tmstate.Store, startBlock, endBlock int64) error {
	var err error
	latest := blockStore.Height()
	if endBlock == shardEndBlockLatest {
		endBlock = latest
	}

	//////////////////////////////
	// Rollback state to endBlock
	//////////////////////////////
	// prep for outputting progress repeatedly to same line
	needsRollback := endBlock < latest
	progress := "rolling back blockstore & cometbft state to height %d"
	numChars := len(fmt.Sprintf(progress, latest))
	clearLine := fmt.Sprintf("\r%s\r", strings.Repeat(" ", numChars))
	printRollbackProgress := func(h int64) {
		fmt.Print(clearLine)
		fmt.Printf(progress, h)
	}

	// rollback tendermint db
	height := latest
	for height > endBlock {
		beforeRollbackHeight := height
		printRollbackProgress(height - 1)
		height, _, err = tmstate.Rollback(blockStore, stateStore, true)
		if err != nil {
			return fmt.Errorf("failed to rollback cometbft state: %w", err)
		}
		if beforeRollbackHeight == height {
			return fmt.Errorf("attempting to rollback cometbft state height %d failed (no rollback performed)", height)
		}
	}

	if needsRollback {
		fmt.Println()
	} else {
		fmt.Printf("latest store height is already %d\n", latest)
	}

	//////////////////////////////
	// Prune blocks to startBlock
	//////////////////////////////
	// get starting block of block store
	baseBlock := blockStore.Base()

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
	} else {
		fmt.Printf("blockstore and cometbft state begins at block %d\n", baseBlock)
	}

	return nil
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
