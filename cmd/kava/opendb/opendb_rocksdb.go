//go:build rocksdb
// +build rocksdb

package opendb

import (
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/server/types"
	dbm "github.com/tendermint/tm-db"
)

func OpenDB(appOpts types.AppOptions, home string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(home, "data")
	if backendType == dbm.RocksDBBackend {
		return openRocksdb(filepath.Join(dataDir, "application.db"), appOpts)
	}

	return dbm.NewDB("application", backendType, dataDir)
}
