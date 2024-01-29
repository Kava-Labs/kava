package util

import (
	"github.com/spf13/cast"

	tdbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAppDBBackend is a server util from future cosmos-sdk versions
// this func is public for use in the rocksdb compact cmd.
func GetAppDBBackend(opts types.AppOptions) tdbm.BackendType {
	rv := cast.ToString(opts.Get("app-db-backend"))
	if len(rv) == 0 {
		rv = sdk.DBBackend
	}
	if len(rv) == 0 {
		rv = cast.ToString(opts.Get("db-backend"))
	}
	if len(rv) != 0 {
		return tdbm.BackendType(rv)
	}
	return tdbm.GoLevelDBBackend
}
