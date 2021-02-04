package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "hard"

	// LiquidatorAccount module account for liquidator
	LiquidatorAccount = "hard_liquidator"

	// ModuleAccountName name of module account used to hold deposits
	ModuleAccountName = "hard"

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
	PreviousBlockTimeKey          = []byte{0x01}
	DepositsKeyPrefix             = []byte{0x03}
	BorrowsKeyPrefix              = []byte{0x05}
	BorrowedCoinsPrefix           = []byte{0x06}
	SuppliedCoinsPrefix           = []byte{0x07}
	MoneyMarketsPrefix            = []byte{0x08}
	PreviousAccrualTimePrefix     = []byte{0x09} // denom -> time
	TotalReservesPrefix           = []byte{0x10} // denom -> sdk.Coin
	BorrowInterestFactorPrefix    = []byte{0x11} // denom -> sdk.Dec
	SupplyInterestFactorPrefix    = []byte{0x12} // denom -> sdk.Dec
	DelegatorInterestFactorPrefix = []byte{0x12} // denom -> sdk.Dec
	LtvIndexPrefix                = []byte{0x13}
	sep                           = []byte(":")
)

// DepositTypeIteratorKey returns an interator prefix for interating over deposits by deposit denom
func DepositTypeIteratorKey(denom string) []byte {
	return createKey([]byte(denom))
}

// GetBorrowByLtvKey is used by the LTV index
func GetBorrowByLtvKey(ltv sdk.Dec, borrower sdk.AccAddress) []byte {
	return append(ltv.Bytes(), borrower...)
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
