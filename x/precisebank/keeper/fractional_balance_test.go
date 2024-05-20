package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/precisebank/types"
)

func TestSetGetFractionalBalance(t *testing.T) {
	addr := sdk.AccAddress([]byte("test-address"))

	tests := []struct {
		name        string
		address     sdk.AccAddress
		amount      sdkmath.Int
		setPanicMsg string
	}{
		{
			"valid - min amount",
			addr,
			sdkmath.NewInt(1),
			"",
		},
		{
			"valid - positive amount",
			addr,
			sdkmath.NewInt(100),
			"",
		},
		{
			"valid - max amount",
			addr,
			types.ConversionFactor().SubRaw(1),
			"",
		},
		{
			"valid - zero amount (deletes)",
			addr,
			sdkmath.ZeroInt(),
			"",
		},
		{
			"invalid - negative amount",
			addr,
			sdkmath.NewInt(-1),
			"amount is invalid: non-positive amount -1",
		},
		{
			"invalid - over max amount",
			addr,
			types.ConversionFactor(),
			"amount is invalid: amount 1000000000000 exceeds max of 999999999999",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			td := NewMockedTestData(t)
			ctx, k := td.ctx, td.keeper

			if tt.setPanicMsg != "" {
				require.PanicsWithError(t, tt.setPanicMsg, func() {
					k.SetFractionalBalance(ctx, tt.address, tt.amount)
				})

				return
			}

			require.NotPanics(t, func() {
				k.SetFractionalBalance(ctx, tt.address, tt.amount)
			})

			// If its zero balance, check it was deleted in store
			if tt.amount.IsZero() {
				store := prefix.NewStore(ctx.KVStore(td.storeKey), types.FractionalBalancePrefix)
				bz := store.Get(types.FractionalBalanceKey(tt.address))
				require.Nil(t, bz)

				return
			}

			gotAmount := k.GetFractionalBalance(ctx, tt.address)
			require.Equal(t, tt.amount, gotAmount)

			// Delete balance
			k.DeleteFractionalBalance(ctx, tt.address)

			store := prefix.NewStore(ctx.KVStore(td.storeKey), types.FractionalBalancePrefix)
			bz := store.Get(types.FractionalBalanceKey(tt.address))
			require.Nil(t, bz)
		})
	}
}

func TestSetFractionalBalance_InvalidAddr(t *testing.T) {
	tk := NewMockedTestData(t)
	ctx, k := tk.ctx, tk.keeper

	require.PanicsWithError(
		t,
		"address cannot be empty",
		func() {
			k.SetFractionalBalance(ctx, sdk.AccAddress{}, sdkmath.NewInt(100))
		},
		"setting balance with empty address should panic",
	)
}

func TestSetFractionalBalance_ZeroDeletes(t *testing.T) {
	td := NewMockedTestData(t)
	ctx, k := td.ctx, td.keeper

	addr := sdk.AccAddress([]byte("test-address"))

	// Set balance
	k.SetFractionalBalance(ctx, addr, sdkmath.NewInt(100))

	bal := k.GetFractionalBalance(ctx, addr)
	require.Equal(t, sdkmath.NewInt(100), bal)

	// Set zero balance
	k.SetFractionalBalance(ctx, addr, sdkmath.ZeroInt())

	// Check balance was deleted
	store := prefix.NewStore(ctx.KVStore(td.storeKey), types.FractionalBalancePrefix)
	bz := store.Get(types.FractionalBalanceKey(addr))
	require.Nil(t, bz)

	// Set zero balance again on non-existent balance
	require.NotPanics(
		t,
		func() {
			k.SetFractionalBalance(ctx, addr, sdkmath.ZeroInt())
		},
		"deleting non-existent balance should not panic",
	)
}

func TestIterateFractionalBalances(t *testing.T) {
	tk := NewMockedTestData(t)
	ctx, k := tk.ctx, tk.keeper

	addrs := []sdk.AccAddress{}

	for i := 1; i < 10; i++ {
		addr := sdk.AccAddress([]byte{byte(i)})
		addrs = append(addrs, addr)

		// Set balance same as their address byte
		k.SetFractionalBalance(ctx, addr, sdkmath.NewInt(int64(i)))
	}

	seenAddrs := []sdk.AccAddress{}

	k.IterateFractionalBalances(ctx, func(addr sdk.AccAddress, bal sdkmath.Int) bool {
		seenAddrs = append(seenAddrs, addr)

		// Balance is same as first address byte
		require.Equal(t, int64(addr.Bytes()[0]), bal.Int64())

		return false
	})

	require.ElementsMatch(t, addrs, seenAddrs, "all addresses should be seen")
}

func TestGetAggregateSumFractionalBalances(t *testing.T) {
	tk := NewMockedTestData(t)
	ctx, k := tk.ctx, tk.keeper

	// Set balances from 1 to 10
	sum := sdkmath.ZeroInt()
	for i := 1; i < 10; i++ {
		addr := sdk.AccAddress([]byte{byte(i)})
		amt := sdkmath.NewInt(int64(i))

		sum = sum.Add(amt)

		require.NotPanics(t, func() {
			k.SetFractionalBalance(ctx, addr, amt)
		})
	}

	gotSum := k.GetTotalSumFractionalBalances(ctx)
	require.Equal(t, sum, gotSum)
}
