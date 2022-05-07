package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "evmutil"

	StoreKey = "utilevm" // cannot be emvutil due to collision with x/evm
)

var AccountStoreKeyPrefix = []byte{0x00} // prefix for keys that store accounts

// AccountStoreKey turns an address to a key used to get the account from the store
func AccountStoreKey(addr sdk.AccAddress) []byte {
	return append(AccountStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}
