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
)

// Keys
// 0x00:Denom <> RewardPeriod the current active reward period (max 1 reward period per denom)
// 0x01:Denom:ID <> ClaimPeriod object for that ID, indexed by denom and ID
// 0x02:Denom:ID:Owner <> Claim object, indexed by Denom, ID and owner
// 0x03:Denom <> NextClaimPeriodIDPrefix the ID of the next claim period, indexed by denom

// BytesToUint64 returns uint64 format from a byte array
func BytesToUint64(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}

// GetClaimPeriodPrefix returns the key (denom + id) for a claim prefix
func GetClaimPeriodPrefix(denom string, id uint64) []byte {
	return createKey([]byte(denom), sdk.Uint64ToBigEndian(id))
}

// GetClaimPrefix returns the key (denom + id + address) for a claim
func GetClaimPrefix(addr sdk.AccAddress, denom string, id uint64) []byte {
	return createKey([]byte(denom), sdk.Uint64ToBigEndian(id), addr)
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
