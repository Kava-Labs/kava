//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"github.com/spf13/cobra"
)

// RocksDBCmd defines the root command containing subcommands that assist in
// rocksdb related tasks such as manual compaction.
var RocksDBCmd = &cobra.Command{
	Use:   "rocksdb",
	Short: "RocksDB util commands",
}

func init() {
	RocksDBCmd.AddCommand(CompactRocksDBCmd())
}
