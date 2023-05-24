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

// IsNil returns true when the address is the 0 address
func (a InternalEVMAddress) IsNil() bool {
	return a.Address == common.Address{}
}

// NewInternalEVMAddress returns a new InternalEVMAddress from a common.Address.
func NewInternalEVMAddress(addr common.Address) InternalEVMAddress {
	return InternalEVMAddress{
		Address: addr,
	}
}

// BytesToInternalEVMAddress creates an InternalEVMAddress from a slice of bytes
func BytesToInternalEVMAddress(bz []byte) InternalEVMAddress {
	return NewInternalEVMAddress(common.BytesToAddress(bz))
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
