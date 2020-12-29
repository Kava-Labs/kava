package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "incentive"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// QuerierRoute route used for abci queries
	QuerierRoute = ModuleName
)

// Key Prefixes
var (
	RewardPeriodKeyPrefix   = []byte{0x01} // prefix for keys that store reward periods
	ClaimPeriodKeyPrefix    = []byte{0x02} // prefix for keys that store claim periods
	ClaimKeyPrefix          = []byte{0x03} // prefix for keys that store claims
	NextClaimPeriodIDPrefix = []byte{0x04} // prefix for keys that store the next ID for claims periods
	PreviousBlockTimeKey    = []byte{0x05} // prefix for key that stores the previous blocktime
	RewardFactorKey         = []byte{0x06}
)

// Keys
// 0x00:CollateralType <> RewardPeriod the current active reward period (max 1 reward period per collateral type)
// 0x01:CollateralType:ID <> ClaimPeriod object for that ID, indexed by collateral type and ID
// 0x02:CollateralType:ID:Owner <> Claim object, indexed by collateral type, ID and owner
// 0x03:CollateralType <> NextClaimPeriodIDPrefix the ID of the next claim period, indexed by collateral type

// BytesToUint64 returns uint64 format from a byte array
func BytesToUint64(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}

// GetClaimPeriodPrefix returns the key (collateral type + id) for a claim prefix
func GetClaimPeriodPrefix(collateralType string, id uint64) []byte {
	return createKey([]byte(collateralType), sdk.Uint64ToBigEndian(id))
}

// GetClaimPrefix returns the key (collateral type + id + address) for a claim
func GetClaimPrefix(addr sdk.AccAddress, collateralType string) []byte {
	return createKey([]byte(collateralType), addr)
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
