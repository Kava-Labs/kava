package types

import (
	"encoding/binary"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "cdp"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

// 1. CDPs are only stored if they have >0 debt
// 2. When a CDPs debt is fully repaid, it is removed from the store
//

// Keys for cdp store
// Items are stored with the following key: values
// - 0x00<cdpOwner_Bytes>: []cdpID
//    - One cdp owner can control one cdp per collateral type
// - 0x01<collateralDenomPrefix><cdpID_Bytes>: CDP
//    - cdps are prefix by denom prefix so we can iterate over cdps of one type
// - 0x02<collateralDenomPrefix><collateralDebtRatio_Bytes><cdpID_Bytes>: cdpID:Denom
// - Ox03: nextCdpID
// - 0x04<cdpID><depositorAddr_bytes>: Deposit
// - 0x20 - 0xff are reserved for collaterals

var (
	CdpIdKeyPrefix             = []byte{0x00}
	CdpKeyPrefix               = []byte{0x01}
	CollateralRatioIndexPrefix = []byte{0x02}
	CdpIdKey                   = []byte{0x03}
	DepositsKeyPrefix          = []byte{0x04}
)

// GetCdpIDBytes returns the byte representation of the cdpID
func GetCdpIDBytes(cdpID uint64) (cdpIDBz []byte) {
	cdpIDBz = make([]byte, 8)
	binary.BigEndian.PutUint64(cdpIDBz, cdpID)
	return
}

// GetCdpIDFromBytes returns cdpID in uint64 format from a byte array
func GetCdpIDFromBytes(bz []byte) (cdpID uint64) {
	return binary.BigEndian.Uint64(bz)
}
