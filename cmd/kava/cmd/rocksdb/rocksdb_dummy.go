//go:build !rocksdb
// +build !rocksdb

package rocksdb

import (
	"github.com/spf13/cobra"
)

// RocksDBCmd defines the root command when the rocksdb build tag is not set.
var RocksDBCmd = &cobra.Command{
	Use:   "rocksdb",
	Short: "RocksDB util commands, disabled because rocksdb build tag not set",
}
