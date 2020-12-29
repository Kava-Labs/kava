package types

import (
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
	ClaimKeyPrefix  = []byte{0x01} // prefix for keys that store claims
	BlockTimeKey    = []byte{0x02} // prefix for key that stores the blocktime
	RewardFactorKey = []byte{0x03} // prefix for key that stores reward factors
)

// Keys
// 0x00:CollateralType <> RewardPeriod the current active reward period (max 1 reward period per collateral type)
// 0x01:CollateralType:ID <> ClaimPeriod object for that ID, indexed by collateral type and ID
// 0x02:CollateralType:ID:Owner <> Claim object, indexed by collateral type, ID and owner
// 0x03:CollateralType <> NextClaimPeriodIDPrefix the ID of the next claim period, indexed by collateral type

// GetClaimPrefix returns the key (addr + collateralType) for a claim
func GetClaimPrefix(addr sdk.AccAddress, collateralType string) []byte {
	return createKey(addr, []byte(collateralType))
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
