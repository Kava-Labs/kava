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

// TODO: Consider store optimizations
var (
	PreviousBlockTimeKey              = []byte{0x01}
	PreviousDelegationDistributionKey = []byte{0x02}
	DepositsKeyPrefix                 = []byte{0x03}
	ClaimsKeyPrefix                   = []byte{0x04}
	BorrowsKeyPrefix                  = []byte{0x05}
	BorrowedCoinsPrefix               = []byte{0x06}
	InterestRateModelsPrefix          = []byte{0x07}
	PreviousAccrualTimePrefix         = []byte{0x08} // denom -> time
	TotalBorrowsPrefix                = []byte{0x09} // denom -> sdk.Coin
	TotalReservesPrefix               = []byte{0x10} // denom -> sdk.Coin
	BorrowIndexPrefix                 = []byte{0x11} // denom -> sdk.Dec
	ReserveFactorPrefix               = []byte{0x12}
	sep                               = []byte(":")
)

// DepositKey key of a specific deposit in the store
func DepositKey(denom string, depositor sdk.AccAddress) []byte {
	return createKey([]byte(denom), sep, depositor)
}

// DepositTypeIteratorKey returns an interator prefix for interating over deposits by deposit denom
func DepositTypeIteratorKey(denom string) []byte {
	return createKey([]byte(denom))
}

// ClaimKey key of a specific deposit in the store
func ClaimKey(depositType ClaimType, denom string, owner sdk.AccAddress) []byte {
	return createKey([]byte(depositType), sep, []byte(denom), sep, owner)
}

// ClaimTypeIteratorKey returns an interator prefix for interating over claims by deposit type and denom
func ClaimTypeIteratorKey(depositType ClaimType, denom string) []byte {
	return createKey([]byte(depositType), sep, []byte(denom))
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
