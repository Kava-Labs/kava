//go:build !rocksdb
// +build !rocksdb

package opendb

import (
	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/server/types"
)

// OpenDB is a copy of default DBOpener function used by ethermint, see for details:
// https://github.com/evmos/ethermint/blob/07cf2bd2b1ce9bdb2e44ec42a39e7239292a14af/server/start.go#L647
func OpenDB(_ types.AppOptions, home string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(home, "data")
	return dbm.NewDB("application", backendType, dataDir)
}
