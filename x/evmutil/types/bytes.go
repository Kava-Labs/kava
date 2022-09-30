package types

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// HexBytes represents a byte slice that marshals into a 0x representation
type HexBytes []byte

// MarshalJSON marshals HexBytes into a 0x json string
func (b HexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hexutil.Encode(b))
}

// UnmarshalJSON unmarshals a 0x json string into bytes
func (b *HexBytes) UnmarshalJSON(input []byte) error {
	return (*hexutil.Bytes)(b).UnmarshalJSON(input)
}

// String implements Stringer and returns the 0x representation
func (b HexBytes) String() string {
	return hexutil.Encode(b)
}
