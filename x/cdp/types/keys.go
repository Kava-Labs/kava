package types

import (
	"bytes"
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

var sep = []byte(":")

// 1. CDPs are only stored if they have >0 debt
// 2. When a CDPs debt is fully repaid, it is removed from the store
//

// Keys for cdp store
// Items are stored with the following key: values
// - 0x00<cdpOwner_Bytes>: []cdpID
//    - One cdp owner can control one cdp per collateral type
// - 0x01<collateralDenomPrefix>:<cdpID_Bytes>: CDP
//    - cdps are prefix by denom prefix so we can iterate over cdps of one type
//    - uses : as separator, otherwise the cdp with ID 11 would be selected when iterating over denom with prefix 1
// - 0x02<collateralDenomPrefix>:<collateralDebtRatio_Bytes>:<cdpID_Bytes>: cdpID
// - Ox03: nextCdpID
// - 0x04: debtDenom
// - 0x05<cdpID>:<depositorAddr_bytes>: Deposit
// - 0x20 - 0xff are reserved for collaterals

var (
	CdpIdKeyPrefix             = []byte{0x00}
	CdpKeyPrefix               = []byte{0x01}
	CollateralRatioIndexPrefix = []byte{0x02}
	CdpIdKey                   = []byte{0x03}
	DebtDenomKey               = []byte{0x04}
	DepositKeyPrefix           = []byte{0x05}
	PrincipalKeyPrefix         = []byte{0x06}
	AccumulatorKeyPrefix       = []byte{0x07}
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

// CdpKey key of a specific cdp in the store
func CdpKey(denomByte byte, cdpID uint64) []byte {
	prefix := append([]byte{denomByte}, sep...)
	return append(prefix, GetCdpIDBytes(cdpID)...)
}

// SplitCdpKey returns the component parts of a cdp key
func SplitCdpKey(key []byte) (byte, uint64) {
	split := bytes.Split(key, sep)
	return split[0][0], GetCdpIDFromBytes(split[1])
}

// DenomIterKey returns the key for iterating over cdps of a certain denom in the store
func DenomIterKey(denomByte byte) []byte {
	return append([]byte{denomByte}, sep...)
}

// SplitDenomIterKey returns the component part of a key for iterating over cdps by denom
func SplitDenomIterKey(key []byte) byte {
	split := bytes.Split(key, sep)
	return split[0][0]
}

// DepositKey key of a specific deposit in the store
func DepositKey(cdpID uint64, depositor sdk.AccAddress) []byte {
	prefix := append(GetCdpIDBytes(cdpID), sep...)
	return append(prefix, depositor...)
}

// SplitDepositKey returns the component parts of a cdp key
func SplitDepositKey(key []byte) (uint64, sdk.AccAddress) {
	split := bytes.Split(key, sep)
	return GetCdpIDFromBytes(split[0]), split[1]
}

// CollateralRatioBytes returns the liquidation ratio as sortable bytes
func CollateralRatioBytes(ratio sdk.Dec) []byte {
	ok := ValidSortableDec(ratio)
	if !ok {
		// set to max sortable if input is too large.
		ratio = sdk.OneDec().Quo(sdk.SmallestDec())
	}
	return SortableDecBytes(ratio)
}

// CollateralRatioKey returns the key for querying a cdp by its liquidation ratio
func CollateralRatioKey(denomByte byte, cdpID uint64, ratio sdk.Dec) []byte {
	ratioBytes := CollateralRatioBytes(ratio)
	prefix := append([]byte{denomByte}, sep...)
	prefix = append(prefix, ratioBytes...)
	prefix = append(prefix, sep...)
	return append(prefix, GetCdpIDBytes(cdpID)...)
}

// SplitCollateralRatioKey split the collateral ratio key and return the denom, cdp id, and collateral:debt ratio
func SplitCollateralRatioKey(key []byte) (denom byte, cdpID uint64, ratio sdk.Dec) {
	split := bytes.Split(key, sep)
	denom = split[0][0]

	ratio, err := ParseDecBytes(split[1])
	if err != nil {
		panic(err)
	}
	cdpID = GetCdpIDFromBytes(split[2])
	return
}

// CollateralRatioIterKey returns the key for iterating over cdps by denom and liquidation ratio
func CollateralRatioIterKey(denomByte byte, ratio sdk.Dec) []byte {
	ratioBytes := CollateralRatioBytes(ratio)
	prefix := append([]byte{denomByte}, sep...)
	return append(prefix, ratioBytes...)
}

// SplitCollateralRatioIterKey split the collateral ratio key and return the denom, cdp id, and collateral:debt ratio
func SplitCollateralRatioIterKey(key []byte) (denom byte, ratio sdk.Dec) {
	split := bytes.Split(key, sep)
	denom = split[0][0]

	ratio, err := ParseDecBytes(split[1])
	if err != nil {
		panic(err)
	}
	return
}
