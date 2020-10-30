package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "harvest"

	// LPAccount LP distribution module account
	LPAccount = "harvest_lp_distribution"

	// DelegatorAccount delegator distribution module account
	DelegatorAccount = "harvest_delegator_distribution"

	// ModuleAccountName name of module account used to hold deposits
	ModuleAccountName = "harvest"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// QuerierRoute Top level query string
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

var (
	PreviousBlockTimeKey              = []byte{0x01}
	PreviousDelegationDistributionKey = []byte{0x02}
	DepositsKeyPrefix                 = []byte{0x03}
	ClaimsKeyPrefix                   = []byte{0x04}
	BorrowsKeyPrefix                  = []byte{0x05}
	sep                               = []byte(":")
)

// DepositKey key of a specific deposit in the store
func DepositKey(depositType DepositType, denom string, depositor sdk.AccAddress) []byte {
	return createKey([]byte(depositType), sep, []byte(denom), sep, depositor)
}

// DepositTypeIteratorKey returns an interator prefix for interating over deposits by deposit type and denom
func DepositTypeIteratorKey(depositType DepositType, denom string) []byte {
	return createKey([]byte(depositType), sep, []byte(denom))
}

// ClaimKey key of a specific deposit in the store
func ClaimKey(depositType DepositType, denom string, owner sdk.AccAddress) []byte {
	return createKey([]byte(depositType), sep, []byte(denom), sep, owner)
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
