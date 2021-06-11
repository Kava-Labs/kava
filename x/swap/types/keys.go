package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "swap"

	// ModuleAccountName name of module account used to hold liquidity
	ModuleAccountName = "swap"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// QuerierRoute Top level query string
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

var (
	PoolKeyPrefix             = []byte{0x01}
	DepositorPoolSharesPrefix = []byte{0x02}

	sep = []byte(":")
)

func PoolKey(poolName string) []byte {
	return []byte(poolName)
}

func DepositorPoolSharesKey(depositor sdk.AccAddress, poolName string) []byte {
	return createKey(depositor, sep, []byte(poolName))
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
