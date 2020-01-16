package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "auction"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// QuerierRoute route used for abci queries
	QuerierRoute = ModuleName
)

// Key prefixes
var (
	AuctionKeyPrefix       = []byte{0x00} // prefix for keys that store auctions
	AuctionByTimeKeyPrefix = []byte{0x01} // prefix for keys that are part of the auctionsByTime index

	NextAuctionIDKey = []byte{0x02} // key for the next auction id
)

// GetAuctionKey returns the bytes of an auction key
func GetAuctionKey(auctionID uint64) []byte {
	return Uint64ToBytes(auctionID)
}

// GetAuctionByTimeKey returns the key for iterating auctions by time
func GetAuctionByTimeKey(endTime time.Time, auctionID uint64) []byte {
	return append(sdk.FormatTimeBytes(endTime), Uint64ToBytes(auctionID)...)
}

// Uint64ToBytes converts a uint64 into fixed length bytes for use in store keys.
func Uint64ToBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(id))
	return bz
}

// Uint64FromBytes converts some fixed length bytes back into a uint64.
func Uint64FromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
