package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "precisebank"

	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName
)

// key prefixes for store
var (
	FractionalBalancePrefix = []byte{0x01} // address -> fractional balance
)

// Keys for store that are not prefixed
var (
	RemainderBalanceKey = []byte{0x02} // fractional balance remainder
)

// FractionalBalanceKey returns a key from an address
func FractionalBalanceKey(address sdk.AccAddress) []byte {
	return address.Bytes()
}
