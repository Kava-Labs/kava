//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/linxGnu/grocksdb"
	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/log"
)

var CompactRocksDBCmd = &cobra.Command{
	Use:   "compact <state|blockstore>",
	Short: "force compacts the CometBFT storage engine (only RocksDB supported)",
	Long: `
This is a temporary utility command that performs a force compaction on the state 
and blockstores to reduce disk space for a pruning node. This should only be run 
once the node has stopped. This command will likely be omitted in the future after
the planned refactor to the storage engine.

Currently, only RocksDB is supported.
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

		clientCtx := client.GetClientContextFromCmd(cmd)
		ctx := server.GetServerContextFromCmd(cmd)

		if server.GetAppDBBackend(ctx.Viper) != "rocksdb" {
			return errors.New("compaction is currently only supported with rocksdb")
		}

		if args[0] != "state" && args[0] != "blockstore" {
			return errors.New("invalid db name, must be either 'state' or 'blockstore'")
		}

		compactRocksDBs(clientCtx.HomeDir, logger, args[0])
		return nil
	},
}

// compactRocksDBs performs a manual compaction on the given db.
func compactRocksDBs(
	rootDir string,
	logger log.Logger,
	dbName string,
) {
	options := grocksdb.NewDefaultOptions()

	dbPath := filepath.Join(rootDir, "data", dbName+".db")
	logger.Info("opening db", "path", dbPath)

	store, err := grocksdb.OpenDb(options, dbPath)
	if err != nil {
		logger.Error("failed to initialize cometbft db", "path", dbPath, "err", err)
		return
	}
	defer store.Close()

	logColumnFamilyMetadata(store, logger)

	logger.Info("starting compaction...", "db", dbPath)

	done := make(chan bool, 1)
	registerSignalHandler(store, logger, done)
	startCompactionStatsOutput(store, logger, done)

	// Actually run the compaction
	store.CompactRange(grocksdb.Range{Start: nil, Limit: nil})
	logger.Info("done compaction", "db", dbPath)

	done <- true
}

// bytesToMB converts bytes to megabytes.
func bytesToMB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024
}

func logColumnFamilyMetadata(
	store *grocksdb.DB,
	logger log.Logger,
) {
	metadata := store.GetColumnFamilyMetadata()

	logger.Info(
		"column family metadata",
		"name", metadata.Name(),
		"sizeMB", bytesToMB(metadata.Size()),
		"fileCount", metadata.FileCount(),
		"levels", len(metadata.LevelMetas()),
	)

	for _, level := range metadata.LevelMetas() {
		logger.Info(
			fmt.Sprintf("level %d metadata", level.Level()),
			"sstMetas", strconv.Itoa(len(level.SstMetas())),
			"sizeMB", strconv.FormatFloat(bytesToMB(level.Size()), 'f', 2, 64),
		)
	}
}

// startCompactionStatsOutput starts a goroutine that outputs compaction stats
// every minute.
func startCompactionStatsOutput(
	store *grocksdb.DB,
	logger log.Logger,
	done chan bool,
) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)

		for {
			select {
			case <-done:
				// Make sure we don't try reading from the closed db
				return
			case <-ticker.C:
				{
					compactionStats := store.GetProperty("rocksdb.stats")
					fmt.Printf("%s\n", compactionStats)
				}
			}
		}
	}()
}

// registerSignalHandler registers a signal handler that will cancel any running
// compaction when the user presses Ctrl+C.
func registerSignalHandler(
	store *grocksdb.DB,
	logger log.Logger,
	done chan bool,
) {
	// https://github.com/facebook/rocksdb/wiki/RocksDB-FAQ
	// Q: Can I close the DB when a manual compaction is in progress?
	//
	// A: No, it's not safe to do that. However, you call
	// CancelAllBackgroundWork(db, true) in another thread to abort the
	// running compactions, so that you can close the DB sooner. Since
	// 6.5, you can also speed it up using
	// DB::DisableManualCompaction().
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range c {
			logger.Info(fmt.Sprintf(
				"received %s signal, aborting running compaction... Do NOT kill me before compaction is cancelled. I will exit when compaction is cancelled.",
				sig,
			))
			store.DisableManualCompaction()
			logger.Info("manual compaction disabled")

			// Stop the logging
			done <- true

			logger.Info("exiting...")
			os.Exit(0)
		}
	}()
}
