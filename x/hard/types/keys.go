package types

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "hard"

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
	DepositsKeyPrefix             = []byte{0x01}
	BorrowsKeyPrefix              = []byte{0x02}
	BorrowedCoinsPrefix           = []byte{0x03}
	SuppliedCoinsPrefix           = []byte{0x04}
	MoneyMarketsPrefix            = []byte{0x05}
	PreviousAccrualTimePrefix     = []byte{0x06} // denom -> time
	TotalReservesPrefix           = []byte{0x07} // denom -> sdk.Coin
	BorrowInterestFactorPrefix    = []byte{0x08} // denom -> sdk.Dec
	SupplyInterestFactorPrefix    = []byte{0x09} // denom -> sdk.Dec
	DelegatorInterestFactorPrefix = []byte{0x10} // denom -> sdk.Dec
)

// DepositTypeIteratorKey returns an interator prefix for interating over deposits by deposit denom
func DepositTypeIteratorKey(denom string) []byte {
	return createKey([]byte(denom))
}

func createKey(bytes ...[]byte) (r []byte) {
	for _, b := range bytes {
		r = append(r, b...)
	}
	return
}
