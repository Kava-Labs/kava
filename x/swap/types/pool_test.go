package types_test

import (
	"testing"

	types "github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

// TODO: example tests for init share math, cover edge cases
func TestPool_InitShares(t *testing.T) {
	a := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	b := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	pool, err := types.NewPool(a, b)
	assert.Nil(t, err)

	assert.Equal(t, a, pool.ReservesA)
	assert.Equal(t, b, pool.ReservesB)

	assert.Equal(t, sdk.NewInt(2236067), pool.TotalShares)
}

func TestPool_Name(t *testing.T) {
	a := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	b := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	pool, err := types.NewPool(a, b)
	assert.Nil(t, err)

	assert.Equal(t, "ukava/usdx", pool.Name())
}

func TestPool_ShareValue(t *testing.T) {
	a := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	b := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	pool, err := types.NewPool(a, b)
	assert.Nil(t, err)

	shareValue, err := pool.ShareValue(pool.TotalShares)
	assert.Nil(t, err)
	assert.Equal(t, sdk.NewCoins(a, b), shareValue)
}
