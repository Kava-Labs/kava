package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestGetSetRemainderAmount(t *testing.T) {
	tk := NewTestKeeper()
	ctx, k, storeKey := tk.ctx, tk.keeper, tk.storeKey

	// Set amount
	k.SetRemainderAmount(ctx, sdkmath.NewInt(100))

	amt := k.GetRemainderAmount(ctx)
	require.Equal(t, sdkmath.NewInt(100), amt)

	// Set zero balance
	k.SetRemainderAmount(ctx, sdkmath.ZeroInt())

	amt = k.GetRemainderAmount(ctx)
	require.Equal(t, sdkmath.ZeroInt(), amt)

	// Get directly from store to make sure it was actually deleted
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.RemainderBalanceKey)
	require.Nil(t, bz)
}

func TestInvalidRemainderAmount(t *testing.T) {
	tk := NewTestKeeper()
	ctx, k := tk.ctx, tk.keeper

	// Set negative amount
	require.PanicsWithError(t, "remainder amount is invalid: non-positive amount -1", func() {
		k.SetRemainderAmount(ctx, sdkmath.NewInt(-1))
	})

	// Set amount over max
	require.PanicsWithError(t, "remainder amount is invalid: amount 1000000000000 exceeds max of 999999999999", func() {
		k.SetRemainderAmount(ctx, types.ConversionFactor())
	})
}

func TestDeleteRemainderAmount(t *testing.T) {
	tk := NewTestKeeper()
	ctx, k, storeKey := tk.ctx, tk.keeper, tk.storeKey

	require.NotPanics(t, func() {
		k.DeleteRemainderAmount(ctx)
	})

	// Set amount
	k.SetRemainderAmount(ctx, sdkmath.NewInt(100))

	amt := k.GetRemainderAmount(ctx)
	require.Equal(t, sdkmath.NewInt(100), amt)

	// Delete amount
	k.DeleteRemainderAmount(ctx)

	amt = k.GetRemainderAmount(ctx)
	require.Equal(t, sdkmath.ZeroInt(), amt)

	store := ctx.KVStore(storeKey)
	bz := store.Get(types.RemainderBalanceKey)
	require.Nil(t, bz)
}
