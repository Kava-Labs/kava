package types

import (
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

func GetAuctionKey(auctionID ID) []byte {
	return auctionID.Bytes()
}

func GetAuctionByTimeKey(endTime time.Time, auctionID ID) []byte {
	return append(sdk.FormatTimeBytes(endTime), auctionID.Bytes()...)
}
