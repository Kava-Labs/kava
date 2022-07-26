package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kava-labs/kava-bridge/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestHexBytes(t *testing.T) {
	testCases := []struct {
		input  types.HexBytes
		output string
	}{
		{[]byte{}, "0x"}, // empty slice is 0x
		{[]byte{0}, "0x00"},
		{[]byte{1}, "0x01"},
		{[]byte{1, 2, 3}, "0x010203"},
		{[]byte{255}, "0xff"},
		{[]byte{16, 16}, "0x1010"},
		{[]byte{16, 16}, "0x1010"},
	}

	for _, tc := range testCases {
		bz, err := json.Marshal(tc.input)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("\"%s\"", tc.output), string(bz))
	}
}
