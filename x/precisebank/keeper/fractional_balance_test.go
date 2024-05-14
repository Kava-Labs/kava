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

func TestSetGetFractionalBalance(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	tApp := app.NewTestApp()
	cdc := tApp.AppCodec()
	k := keeper.NewKeeper(cdc, storeKey)

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
