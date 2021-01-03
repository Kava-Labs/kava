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

	// SavingsRateMacc module account for savings rate
	SavingsRateMacc = "savings"
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
	CdpIDKeyPrefix              = []byte{0x01}
	CdpKeyPrefix                = []byte{0x02}
	CollateralRatioIndexPrefix  = []byte{0x03}
	CdpIDKey                    = []byte{0x04}
	DebtDenomKey                = []byte{0x05}
	GovDenomKey                 = []byte{0x06}
	DepositKeyPrefix            = []byte{0x07}
	PrincipalKeyPrefix          = []byte{0x08}
	PreviousDistributionTimeKey = []byte{0x09}
	PricefeedStatusKeyPrefix    = []byte{0x10}
	SavingsRateDistributedKey   = []byte{0x11}
	PreviousAccrualTimePrefix   = []byte{0x12}
	InterestFactorPrefix        = []byte{0x13}
	SavingsFactorPrefix         = []byte{0x14}
	SavingsClaimsPrefix         = []byte{0x15}
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
func CdpKey(denomByte byte, cdpID uint64) []byte {
	return createKey([]byte{denomByte}, sep, GetCdpIDBytes(cdpID))
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
func CollateralRatioKey(denomByte byte, cdpID uint64, ratio sdk.Dec) []byte {
	ratioBytes := CollateralRatioBytes(ratio)
	idBytes := GetCdpIDBytes(cdpID)

	return createKey([]byte{denomByte}, sep, ratioBytes, sep, idBytes)
}

// SplitCollateralRatioKey split the collateral ratio key and return the denom, cdp id, and collateral:debt ratio
func SplitCollateralRatioKey(key []byte) (denom byte, cdpID uint64, ratio sdk.Dec) {

	cdpID = GetCdpIDFromBytes(key[len(key)-8:])
	split := bytes.Split(key[:len(key)-8], sep)
	denom = split[0][0]

	ratio, err := ParseDecBytes(split[1])
	if err != nil {
		panic(err)
	}
	return
}

// CollateralRatioIterKey returns the key for iterating over cdps by denom and liquidation ratio
func CollateralRatioIterKey(denomByte byte, ratio sdk.Dec) []byte {
	ratioBytes := CollateralRatioBytes(ratio)
	return createKey([]byte{denomByte}, sep, ratioBytes)
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

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
