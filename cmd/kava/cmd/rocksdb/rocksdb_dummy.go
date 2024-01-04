//go:build !rocksdb
// +build !rocksdb

package rocksdb

import (
	"github.com/spf13/cobra"
)

// RocksDBCmd defines the root command containing subcommands that assist in
// debugging rocksdb.
var RocksDBCmd = &cobra.Command{
	Use:   "rocksdb",
	Short: "RocksDB util commands, disabled because rocksdb build tag not set",
}
