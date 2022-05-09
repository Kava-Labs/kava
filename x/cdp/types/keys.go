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

	// QuerierRoute Top level query string
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// LiquidatorMacc module account for liquidator
	LiquidatorMacc = "liquidator"
)

var sep = []byte(":")

// Keys for cdp store
// Items are stored with the following key: values
// - 0x00<cdpOwner_Bytes>: []cdpID
//    - One cdp owner can control one cdp per collateral type
// - 0x01<collateralDenomPrefix>:<cdpID_Bytes>: CDP
//    - cdps are prefix by denom prefix so we can iterate over cdps of one type
//    - uses : as separator
// - 0x02<collateralDenomPrefix>:<collateralDebtRatio_Bytes>:<cdpID_Bytes>: cdpID
// - Ox03: nextCdpID
// - 0x04: debtDenom
// - 0x05<depositState>:<cdpID>:<depositorAddr_bytes>: Deposit
// - 0x06<denom>:totalPrincipal
// - 0x07<denom>:feeRate
// - 0x08:previousDistributionTime
// - 0x09<marketID>:downTime
// - 0x10:totalDistributed

// KVStore key prefixes
var (
	CdpIDKeyPrefix             = []byte{0x01}
	CdpKeyPrefix               = []byte{0x02}
	CollateralRatioIndexPrefix = []byte{0x03}
	CdpIDKey                   = []byte{0x04}
	DebtDenomKey               = []byte{0x05}
	GovDenomKey                = []byte{0x06}
	DepositKeyPrefix           = []byte{0x07}
	PrincipalKeyPrefix         = []byte{0x08}
	PricefeedStatusKeyPrefix   = []byte{0x10}
	PreviousAccrualTimePrefix  = []byte{0x12}
	InterestFactorPrefix       = []byte{0x13}
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

// CdpKey key of a specific cdp in the store
func CdpKey(collateralType string, cdpID uint64) []byte {
	return createKey([]byte(collateralType), sep, GetCdpIDBytes(cdpID))
}

// SplitCdpKey returns the component parts of a cdp key
func SplitCdpKey(key []byte) (string, uint64) {
	split := bytes.Split(key, sep)
	return string(split[0]), GetCdpIDFromBytes(split[1])
}

// DenomIterKey returns the key for iterating over cdps of a certain denom in the store
func DenomIterKey(collateralType string) []byte {
	return append([]byte(collateralType), sep...)
}

// SplitDenomIterKey returns the component part of a key for iterating over cdps by denom
func SplitDenomIterKey(key []byte) string {
	split := bytes.Split(key, sep)
	return string(split[0])
}

// DepositKey key of a specific deposit in the store
func DepositKey(cdpID uint64, depositor sdk.AccAddress) []byte {
	return createKey(GetCdpIDBytes(cdpID), sep, depositor)
}

// SplitDepositKey returns the component parts of a deposit key
func SplitDepositKey(key []byte) (uint64, sdk.AccAddress) {
	cdpID := GetCdpIDFromBytes(key[0:8])
	addr := key[9:]
	return cdpID, addr
}

// DepositIterKey returns the prefix key for iterating over deposits to a cdp
func DepositIterKey(cdpID uint64) []byte {
	return GetCdpIDBytes(cdpID)
}

// SplitDepositIterKey returns the component parts of a key for iterating over deposits on a cdp
func SplitDepositIterKey(key []byte) (cdpID uint64) {
	return GetCdpIDFromBytes(key)
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
func CollateralRatioKey(collateralType string, cdpID uint64, ratio sdk.Dec) []byte {
	ratioBytes := CollateralRatioBytes(ratio)
	idBytes := GetCdpIDBytes(cdpID)

	return createKey([]byte(collateralType), sep, ratioBytes, sep, idBytes)
}

// SplitCollateralRatioKey split the collateral ratio key and return the denom, cdp id, and collateral:debt ratio
func SplitCollateralRatioKey(key []byte) (string, uint64, sdk.Dec) {
	cdpID := GetCdpIDFromBytes(key[len(key)-8:])
	split := bytes.Split(key[:len(key)-8], sep)
	collateralType := string(split[0])

	ratio, err := ParseDecBytes(split[1])
	if err != nil {
		panic(err)
	}
	return collateralType, cdpID, ratio
}

// CollateralRatioIterKey returns the key for iterating over cdps by denom and liquidation ratio
func CollateralRatioIterKey(collateralType string, ratio sdk.Dec) []byte {
	ratioBytes := CollateralRatioBytes(ratio)
	return createKey([]byte(collateralType), sep, ratioBytes)
}

// SplitCollateralRatioIterKey split the collateral ratio key and return the denom, cdp id, and collateral:debt ratio
func SplitCollateralRatioIterKey(key []byte) (string, sdk.Dec) {
	split := bytes.Split(key, sep)
	collateralType := string(split[0])

	ratio, err := ParseDecBytes(split[1])
	if err != nil {
		panic(err)
	}
	return collateralType, ratio
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
