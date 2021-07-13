package types_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	key := types.PoolKey("ukava/usdx")
	assert.Equal(t, "ukava/usdx", string(key))

	key = types.DepositorPoolSharesKey(sdk.AccAddress("testaddress1"), "ukava/usdx")
	assert.Equal(t, string(sdk.AccAddress("testaddress1"))+":ukava/usdx", string(key))
}
