package types

import (
	"encoding/binary"
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
)

// Key prefixes
var (
	KHTLTKeyPrefix       = []byte{0x00} // prefix for keys that store KHTLTs
	KHTLTByTimeKeyPrefix = []byte{0x01} // prefix for keys that are part of the khtltByTime index

	NextKHTLTIDKey = []byte{0x02} // key for the next KHTLT id
)

// GetKHTLTKey returns the bytes of an KHTLT key
func GetKHTLTKey(htltID uint64) []byte {
	return Uint64ToBytes(htltID)
}

// GetKHTLTByTimeKey returns the key for iterating KHTLTs by time
func GetKHTLTByTimeKey(endTime time.Time, htltID uint64) []byte {
	return append(sdk.FormatTimeBytes(endTime), Uint64ToBytes(htltID)...)
}

// Uint64ToBytes converts a uint64 into fixed length bytes for use in store keys.
func Uint64ToBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(id))
	return bz
}

// TODO: needed?
// Uint64FromBytes converts some fixed length bytes back into a uint64.
func Uint64FromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
