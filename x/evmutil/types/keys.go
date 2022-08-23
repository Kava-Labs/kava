package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "evmutil"

	StoreKey = "utilevm" // note: cannot be emvutil due to collision with x/evm

	// RouterKey Top level router key
	RouterKey = ModuleName
)

var AccountStoreKeyPrefix = []byte{0x00} // prefix for keys that store accounts

// AccountStoreKey turns an address to a key used to get the account from the store
func AccountStoreKey(addr sdk.AccAddress) []byte {
	return append(AccountStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}

// ModuleAddress is the native module address for EVM
var ModuleEVMAddress common.Address

func init() {
	ModuleEVMAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}
