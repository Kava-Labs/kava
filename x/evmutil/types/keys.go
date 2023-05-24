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

// KVStore keys
var (
	// AccountStoreKeyPrefix is the prefix for keys that store accounts
	AccountStoreKeyPrefix = []byte{0x00}
	// DeployedCosmosCoinContractKeyPrefix is the key for storing deployed KavaWrappedCosmosCoinERC20s contract addresses
	DeployedCosmosCoinContractKeyPrefix = []byte{0x01}
)

// AccountStoreKey turns an address to a key used to get the account from the store
func AccountStoreKey(addr sdk.AccAddress) []byte {
	return append(AccountStoreKeyPrefix, address.MustLengthPrefix(addr)...)
}

// DeployedCosmosCoinContractKey gives the store key that holds the address of the deployed ERC20
// that wraps the given cosmosDenom sdk.Coin
func DeployedCosmosCoinContractKey(cosmosDenom string) []byte {
	return append(DeployedCosmosCoinContractKeyPrefix, []byte(cosmosDenom)...)
}

// ModuleAddress is the native module address for EVM
var ModuleEVMAddress common.Address

func init() {
	ModuleEVMAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}
