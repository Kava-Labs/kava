package types

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// ExternalEVMAddress is a type alias of common.Address to represent an address
// on an external EVM, e.g. Ethereum. This is used to make external / internal
// addresses type safe and un-assignable to each other. This also makes it more
// clear which address belongs where.
type ExternalEVMAddress struct {
	common.Address
}

// NewExternalEVMAddress returns a new ExternalEVMAddress from a common.Address.
func NewExternalEVMAddress(addr common.Address) ExternalEVMAddress {
	return ExternalEVMAddress{
		Address: addr,
	}
}

// NewExternalEVMAddressFromString returns a new ExternalEVMAddress from a hex
// string. Returns an error if hex string is invalid.
func NewExternalEVMAddressFromString(addrStr string) (ExternalEVMAddress, error) {
	if !common.IsHexAddress(addrStr) {
		return ExternalEVMAddress{}, fmt.Errorf("string is not a hex address %v", addrStr)
	}

	// common.HexToAddress ignores hex decoding errors
	addr := common.HexToAddress(addrStr)

	return NewExternalEVMAddress(addr), nil
}

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
