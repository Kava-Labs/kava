package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "bep3"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute is the querier route for bep3
	QuerierRoute = ModuleName

	// DefaultParamspace default namestore
	DefaultParamspace = ModuleName

	// DefaultLongtermStorageDuration is 1 week (assuming a block time of 7 seconds)
	DefaultLongtermStorageDuration uint64 = 86400
)

// Key prefixes
var (
	// SupplyLimitUpgradeTime is the block time after which the asset supply limits are updated from params
	SupplyLimitUpgradeTime time.Time = time.Date(2020, 7, 2, 14, 0, 0, 0, time.UTC)

	AtomicSwapKeyPrefix             = []byte{0x00} // prefix for keys that store AtomicSwaps
	AtomicSwapByBlockPrefix         = []byte{0x01} // prefix for keys of the AtomicSwapsByBlock index
	AssetSupplyKeyPrefix            = []byte{0x02} // prefix for keys that store global asset supply counts
	AtomicSwapLongtermStoragePrefix = []byte{0x03} // prefix for keys of the AtomicSwapLongtermStorage index
)

// GetAtomicSwapByHeightKey is used by the AtomicSwapByBlock index and AtomicSwapLongtermStorage index
func GetAtomicSwapByHeightKey(height uint64, swapID []byte) []byte {
	return append(sdk.Uint64ToBigEndian(height), swapID...)
}
