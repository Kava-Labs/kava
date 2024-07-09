package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// KVStore keys
var (
	// AccountStoreKeyPrefix is the prefix for keys that store accounts
	AccountStoreKeyPrefix = []byte{0x00}
)

// AccountStoreKey turns an address to a key used to get the account from the store
func AccountStoreKey(addr sdk.AccAddress) []byte {
	return append(AccountStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}
