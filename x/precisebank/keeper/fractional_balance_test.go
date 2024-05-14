package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
)

func NewTestKeeper() (sdk.Context, keeper.Keeper) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	tApp := app.NewTestApp()
	cdc := tApp.AppCodec()
	k := keeper.NewKeeper(cdc, storeKey)

	return ctx, k
}

func TestSetGetFractionalBalance(t *testing.T) {
	ctx, k := NewTestKeeper()

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
			types.MaxFractionalAmount(),
			"",
		},
		{
			"invalid - zero amount",
			addr,
			sdkmath.ZeroInt(),
			"amount is invalid: non-positive amount 0",
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
			types.MaxFractionalAmount().AddRaw(1),
			"amount is invalid: amount 1000000000000 exceeds max of 999999999999",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.setPanicMsg != "" {
				require.PanicsWithError(t, tt.setPanicMsg, func() {
					k.SetFractionalBalance(ctx, tt.address, tt.amount)
				})

				return
			}

			require.NotPanics(t, func() {
				k.SetFractionalBalance(ctx, tt.address, tt.amount)
			})

			gotAmount, exists := k.GetFractionalBalance(ctx, tt.address)
			require.True(t, exists)
			require.Equal(t, tt.amount, gotAmount)
		})
	}
}

func TestIterateFractionalBalances(t *testing.T) {
	ctx, k := NewTestKeeper()

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
	ctx, k := NewTestKeeper()

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

	gotSum := k.GetAggregateSumFractionalBalances(ctx)
	require.Equal(t, sum, gotSum)
}
