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
	"strings"
	"syscall"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/linxGnu/grocksdb"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"

	"github.com/Kava-Labs/opendb"
)

const (
	flagPrintStatsInterval = "print-stats-interval"
)

var allowedDBs = []string{"application", "blockstore", "state"}

func CompactRocksDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: fmt.Sprintf(
			"compact <%s>",
			strings.Join(allowedDBs, "|"),
		),
		Short: "force compacts RocksDB",
		Long: `This is a utility command that performs a force compaction on the state or
	blockstore. This should only be run once the node has stopped.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

			statsIntervalStr, err := cmd.Flags().GetString(flagPrintStatsInterval)
			if err != nil {
				return err
			}

			statsInterval, err := time.ParseDuration(statsIntervalStr)
			if err != nil {
				return fmt.Errorf("failed to parse duration for --%s: %w", flagPrintStatsInterval, err)
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			ctx := server.GetServerContextFromCmd(cmd)

			if server.GetAppDBBackend(ctx.Viper) != "rocksdb" {
				return errors.New("compaction is currently only supported with rocksdb")
			}

			if !slices.Contains(allowedDBs, args[0]) {
				return fmt.Errorf(
					"invalid db name, must be one of the following: %s",
					strings.Join(allowedDBs, ", "),
				)
			}

			return compactRocksDBs(clientCtx.HomeDir, logger, args[0], statsInterval)
		},
	}

	cmd.Flags().String(flagPrintStatsInterval, "1m", "duration string for how often to print compaction stats")

	return cmd
}

// compactRocksDBs performs a manual compaction on the given db.
func compactRocksDBs(
	rootDir string,
	logger log.Logger,
	dbName string,
	statsInterval time.Duration,
) error {
	dbPath := filepath.Join(rootDir, "data", dbName+".db")

	dbOpts, cfOpts, err := opendb.LoadLatestOptions(dbPath)
	if err != nil {
		return err
	}

	logger.Info("opening db", "path", dbPath)
	db, _, err := grocksdb.OpenDbColumnFamilies(
		dbOpts,
		dbPath,
		[]string{opendb.DefaultColumnFamilyName},
		[]*grocksdb.Options{cfOpts},
	)
	if err != nil {
		return err
	}

	if err != nil {
		logger.Error("failed to initialize cometbft db", "path", dbPath, "err", err)
		return fmt.Errorf("failed to open db %s %w", dbPath, err)
	}
	defer db.Close()

	logColumnFamilyMetadata(db, logger)

	logger.Info("starting compaction...", "db", dbPath)

	done := make(chan bool)
	registerSignalHandler(db, logger, done)
	startCompactionStatsOutput(db, logger, done, statsInterval)

	// Actually run the compaction
	db.CompactRange(grocksdb.Range{Start: nil, Limit: nil})
	logger.Info("done compaction", "db", dbPath)

	done <- true
	return nil
}

// bytesToMB converts bytes to megabytes.
func bytesToMB(bytes uint64) float64 {
	return float64(bytes) / 1024 / 1024
}

// logColumnFamilyMetadata outputs the column family and level metadata.
func logColumnFamilyMetadata(
	db *grocksdb.DB,
	logger log.Logger,
) {
	metadata := db.GetColumnFamilyMetadata()

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
	db *grocksdb.DB,
	logger log.Logger,
	done chan bool,
	statsInterval time.Duration,
) {
	go func() {
		ticker := time.NewTicker(statsInterval)
		isClosed := false

		for {
			select {
			// Make sure we don't try reading from the closed db.
			// We continue the loop so that we can make sure the done channel
			// does not stall indefinitely from repeated writes and no reader.
			case <-done:
				logger.Debug("stopping compaction stats output")
				isClosed = true
			case <-ticker.C:
				if !isClosed {
					compactionStats := db.GetProperty("rocksdb.stats")
					fmt.Printf("%s\n", compactionStats)
				}
			}
		}
	}()
}

// registerSignalHandler registers a signal handler that will cancel any running
// compaction when the user presses Ctrl+C.
func registerSignalHandler(
	db *grocksdb.DB,
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
			db.DisableManualCompaction()
			logger.Info("manual compaction disabled")

			// Stop the logging
			done <- true
		}
	}()
}
