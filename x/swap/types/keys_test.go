package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/swap/types"
)

func TestKeys(t *testing.T) {
	key := types.PoolKey(types.PoolID("ukava", "usdx"))
	assert.Equal(t, types.PoolID("ukava", "usdx"), string(key))

	key = types.DepositorPoolSharesKey(sdk.AccAddress("testaddress1"), types.PoolID("ukava", "usdx"))
	assert.Equal(t, string(sdk.AccAddress("testaddress1"))+"|"+types.PoolID("ukava", "usdx"), string(key))
}
