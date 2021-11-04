package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validBech32Addr = "kava16g8lzm86f5wwf3x3t67qrpd46sjdpxpfazskwg"
var validJSONAddr = []byte(fmt.Sprintf("\"%s\"", validBech32Addr))

func TestAddress(t *testing.T) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	var a types.Address

	err := a.UnmarshalJSON(validJSONAddr)
	require.NoError(t, err, "expected json unmarshal to not error")

	assert.Equal(t, 20, a.Size(), "expected address size to be 20")

	data, err := a.MarshalJSON()
	assert.Equal(t, validJSONAddr, data, "expected re-marshalled address to equal original")

	bz, err := a.Marshal()
	require.NoError(t, err, "expected marshall to not error")

	assert.Equal(t, sdk.AccAddress(a).Bytes(), bz, "expected binary marshal to match AccAddress bytes")

	var a2 types.Address
	err = a2.Unmarshal(bz)
	require.NoError(t, err, "expected unmarshal to not error")

	assert.Equal(t, a2, a, "expected unmarshalled address to be equal")

	buf := make([]byte, 100)
	n, err := a.MarshalTo(buf)
	require.NoError(t, err, "expected marshalto to not error")

	assert.Equal(t, 20, n, "expected marshalto to write 20 bytes")

	assert.Equal(t, sdk.AccAddress(a).String(), a.String(), "expected string to match sdk")
}
