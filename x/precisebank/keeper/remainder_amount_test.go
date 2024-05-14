package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestGetSetRemainderAmount(t *testing.T) {
	ctx, k := NewTestKeeper()

	// Set amount
	k.SetRemainderAmount(ctx, sdkmath.NewInt(100))

	amt := k.GetRemainderAmount(ctx)
	require.Equal(t, sdkmath.NewInt(100), amt)

	// Set zero balance
	k.SetRemainderAmount(ctx, sdkmath.ZeroInt())

	amt = k.GetRemainderAmount(ctx)
	require.Equal(t, sdkmath.ZeroInt(), amt)
}
