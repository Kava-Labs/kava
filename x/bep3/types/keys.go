package types

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
)

// Key prefixes
var (
	HTLTKeyPrefix       = []byte{0x00} // prefix for keys that store KTLTs
	HTLTByTimeKeyPrefix = []byte{0x01} // prefix for keys that are part of the htltByTime index
)

// GetHTLTByTimeKey returns the key for iterating HTLTs by time
// func GetHTLTByTimeKey(endTime time.Time, htltID uint64) []byte {
// 	return append(sdk.FormatTimeBytes(endTime), Uint64ToBytes(htltID)...)
// }
