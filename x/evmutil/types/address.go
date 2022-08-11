package types

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// InternalEVMAddress is a type alias of common.Address to represent an address
// on the Kava EVM.
type InternalEVMAddress struct {
	common.Address
}

// NewInternalEVMAddress returns a new InternalEVMAddress from a common.Address.
func NewInternalEVMAddress(addr common.Address) InternalEVMAddress {
	return InternalEVMAddress{
		Address: addr,
	}
}

// NewInternalEVMAddressFromString returns a new InternalEVMAddress from a hex
// string. Returns an error if hex string is invalid.
func NewInternalEVMAddressFromString(addrStr string) (InternalEVMAddress, error) {
	if !common.IsHexAddress(addrStr) {
		return InternalEVMAddress{}, fmt.Errorf("string is not a hex address %v", addrStr)
	}

	// common.HexToAddress ignores hex decoding errors
	addr := common.HexToAddress(addrStr)

	return NewInternalEVMAddress(addr), nil
}
