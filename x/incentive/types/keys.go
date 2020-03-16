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
	RewardPeriodKeyPrefix = []byte{0x00} // prefix for keys that store reward periods
	ClaimPeriodKeyPrefix  = []byte{0x01} // prefix for keys that store claim periods
	ClaimKeyPrefix        = []byte{0x02} // prefix for keys that store claims
	NextClaimPeriodID     = []byte{0x03} // prefix for keys that store the next ID for claims periods
	PreviousBlockTimeKey  = []byte{0x04}
)

// Keys
// 0x00:Denom <> RewardPeriod the current active reward period (max 1 reward period per denom)
// 0x01:Denom:ID <> ClaimPeriod object for that ID, indexed by denom and ID
// 0x02:Denom:ID:Owner <> Claim object, indexed by Denom, ID and owner
// 0x03:Denom <> NextClaimPeriodID the ID of the next claim period, indexed by denom

// GetIDBytes returns the input as a fixed length byte array
func GetIDBytes(id uint64) []byte {
	return Uint64ToBytes(id)
}

// Uint64ToBytes converts a uint64 into fixed length bytes for use in store keys.
func Uint64ToBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(id))
	return bz
}

// GetDenomBytes returns the input as a byte slice
func GetDenomBytes(denom string) []byte {
	return []byte(denom)
}

// GetDenomFromBytes returns the input as a string
func GetDenomFromBytes(db []byte) string {
	return string(db)
}

// GetClaimPeriodPrefix returns the key (denom + id) for a claim prefix
func GetClaimPeriodPrefix(denom string, id uint64) []byte {
	return createKey(GetDenomBytes(denom), GetIDBytes(id))
}

// GetClaimPrefix returns the key (denom + id + address) for a claim
func GetClaimPrefix(addr sdk.AccAddress, denom string, id uint64) []byte {
	return createKey(GetDenomBytes(denom), GetIDBytes(id), addr)
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
