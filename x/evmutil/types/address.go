package types

import (
	"encoding/json"
	"errors"
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

// Equal checks if two InternalEVMAddress instances are equal.
func (addr InternalEVMAddress) Equal(other InternalEVMAddress) bool {
	return addr.Address == other.Address
}

// MarshalTo implements the protobuf Marshaler interface.
func (addr InternalEVMAddress) MarshalTo(data []byte) (int, error) {
	addressBytes := addr.Address.Bytes()
	return copy(data, addressBytes[:]), nil
}

// MarshalJSON allows PrintProto to handle InternalEVMAddress
func (addr InternalEVMAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.Hex())
}

// Size implements protobuf Unmarshaler interface.
func (a InternalEVMAddress) Size() int {
	return common.AddressLength
}

// Unmarshal implements the protobuf Unmarshaler interface.
func (addr *InternalEVMAddress) Unmarshal(data []byte) error {
	if len(data) != common.AddressLength {
		return errors.New("invalid data length for InternalEVMAddress")
	}
	addr.Address.SetBytes(data)
	return nil
}
