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
)

// TODO use cont to keep immutability?
var (
	AuctionKeyPrefix       = []byte{0x00} // prefix for keys that store auctions
	AuctionByTimeKeyPrefix = []byte{0x01} // prefix for keys that are part of the auctionsByTime index

	NextAuctionIDKey = []byte{0x02}
)

func GetAuctionKey(auctionID uint64) []byte {
	return Uint64ToBytes(auctionID)
}

func GetAuctionByTimeKey(endTime time.Time, auctionID uint64) []byte {
	return append(sdk.FormatTimeBytes(endTime), Uint64ToBytes(auctionID)...)
}

func Uint64FromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}

func Uint64ToBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(id))
	return bz
}