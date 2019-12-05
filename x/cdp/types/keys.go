package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

var lenPositiveDec = len(SortableDecBytes(sdk.OneDec()))
var lenNegativeDec = len(SortableDecBytes(sdk.OneDec().Neg()))

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

// LiquidationRatioKey returns the key for a querying for cdps by liquidation ratio
func LiquidationRatioKey(ratio sdk.Dec) []byte {
	ok := ValidSortableDec(ratio)
	if !ok {
		ratio = sdk.OneDec().Quo(sdk.SmallestDec())
	}
	return SortableDecBytes(ratio)
}

// SplitCollateralRatioKey split the collateral ratio key and return the denom, cdp id, and collateral:debt ratio
func SplitCollateralRatioKey(key []byte) (denom string, ratio sdk.Dec, cdpID uint64) {
	return splitKeyWithDec(key)
}

func splitKeyWithDec(key []byte) (denom string, ratio sdk.Dec, cdpID uint64) {
	denomByte := key[0]
	ratioBytes := key[1 : len(key)-8]
	idBytes := key[len(key)-8:]

	ratio, err := ParseDecBytes(ratioBytes)
	if err != nil {
		panic(err)
	}
	denom = ParseDenomBytes(denomBytes)
	cdpID = GetCdpIDFromBytes(idBytes)
	return

}
