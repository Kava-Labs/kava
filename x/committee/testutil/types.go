package testutil

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

// Avoid cluttering test cases with long function names
func I(in int64) sdkmath.Int                { return sdkmath.NewInt(in) }
func D(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func C(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func Cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func AssertProtoMessageJSON(t *testing.T, cdc codec.Codec, expected proto.Message, actual proto.Message) {
	expectedJson, err := cdc.MarshalJSON(expected)
	assert.NoError(t, err)
	actualJson, err := cdc.MarshalJSON(actual)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedJson), string(actualJson))
}
