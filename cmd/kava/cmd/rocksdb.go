package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/linxGnu/grocksdb"
	"github.com/spf13/cobra"
)

var (
	dbnames = []string{"application.db", "blockstore.db", "evidence.db", "state.db", "tx_index.db"}
)

func RocksDBInfo(homeDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rocksdbinfo",
		Short: "Get rocksdb stats",
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			dataDir := filepath.Join(serverCtx.Config.RootDir, serverCtx.Config.DBPath)

			bbto := grocksdb.NewDefaultBlockBasedTableOptions()
			bbto.SetBlockCache(grocksdb.NewLRUCache(1 << 30))
			bbto.SetFilterPolicy(grocksdb.NewBloomFilter(10))
			opts := grocksdb.NewDefaultOptions()
			opts.SetBlockBasedTableFactory(bbto)
			opts.SetMaxOpenFiles(-1)
			opts.SetCreateIfMissing(true)
			opts.IncreaseParallelism(runtime.NumCPU())
			opts.OptimizeLevelStyleCompaction(512 * 1024 * 1024)

			opts.SetOptimizeFiltersForHits(true)

			for _, dbname := range dbnames {
				db, err := grocksdb.OpenDb(opts, filepath.Join(dataDir, dbname))

				if err != nil {
					return err
				}

				estReaderMem := db.GetProperty("rocksdb.estimate-table-readers-mem")
				numKeys := db.GetProperty("rocksdb.estimate-num-keys")

				fmt.Printf("%s %s %s\n", dbname, estReaderMem, numKeys)

				stats := db.GetProperty("rocksdb.stats")
				fmt.Println(stats)

				db.Close()
			}

			return nil
		},
	}

	return cmd
}
