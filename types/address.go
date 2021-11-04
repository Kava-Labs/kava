package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Address represents an AccAddress designed for use with gogoproto
type Address []byte

// ensure we conform to the correct interface
var _ sdk.CustomProtobufType = (*Address)(nil)

// Marshal returns the raw address bytes
func (a Address) Marshal() ([]byte, error) {
	return a, nil
}

// MarshalTo marshals the address to the provided buffer
func (a Address) MarshalTo(buf []byte) (int, error) {
	return copy(buf, a), nil
}

// Unmarshal sets the address to the given bytes
func (a *Address) Unmarshal(data []byte) error {
	*a = data
	return nil
}

// Size returns the byte length of the address
func (a Address) Size() int {
	return len(a)
}

func (a Address) String() string {
	return sdk.AccAddress(a).String()
}

// MarshalJSON marshals to JSON using Bech32
func (a Address) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(sdk.AccAddress(a).String())
	return b, err
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding
func (a *Address) UnmarshalJSON(data []byte) error {
	return (*sdk.AccAddress)(a).UnmarshalJSON(data)
}
