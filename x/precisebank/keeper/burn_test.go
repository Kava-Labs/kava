package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestBurnCoins_PanicValidations(t *testing.T) {
	// panic tests for invalid inputs

	tests := []struct {
		name            string
		recipientModule string
		setupFn         func(td testData)
		mintAmount      sdk.Coins
		wantPanic       string
	}{
		{
			"invalid module",
			"notamodule",
			func(td testData) {
				// Make module not found
				td.ak.EXPECT().
					GetModuleAccount(td.ctx, "notamodule").
					Return(nil).
					Once()
			},
			cs(c("ukava", 1000)),
			"module account notamodule does not exist: unknown address",
		},
		{
			"no permission",
			minttypes.ModuleName,
			func(td testData) {
				td.ak.EXPECT().
					GetModuleAccount(td.ctx, minttypes.ModuleName).
					Return(authtypes.NewModuleAccount(
						authtypes.NewBaseAccountWithAddress(sdk.AccAddress{1}),
						minttypes.ModuleName,
						// no burn permission
					)).
					Once()
			},
			cs(c("ukava", 1000)),
			"module account mint does not have permissions to mint tokens: unauthorized",
		},
		{
			"has mint permission",
			minttypes.ModuleName,
			func(td testData) {
				td.ak.EXPECT().
					GetModuleAccount(td.ctx, minttypes.ModuleName).
					Return(authtypes.NewModuleAccount(
						authtypes.NewBaseAccountWithAddress(sdk.AccAddress{1}),
						minttypes.ModuleName,
						// includes burner permission
						authtypes.Burner,
					)).
					Once()

				// Will call x/bank BurnCoins coins
				td.bk.EXPECT().
					BurnCoins(td.ctx, minttypes.ModuleName, cs(c("ukava", 1000))).
					Return(nil).
					Once()
			},
			cs(c("ukava", 1000)),
			"",
		},
		{
			"disallow burning from x/precisebank",
			types.ModuleName,
			func(td testData) {
				// No mock setup needed since this is checked before module
				// account checks
			},
			cs(c("ukava", 1000)),
			"module account precisebank cannot be burned from: unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewMockedTestData(t)
			tt.setupFn(td)

			if tt.wantPanic != "" {
				require.PanicsWithError(t, tt.wantPanic, func() {
					_ = td.keeper.BurnCoins(td.ctx, tt.recipientModule, tt.mintAmount)
				})
				return
			}

			require.NotPanics(t, func() {
				// Not testing errors, only panics for this test
				_ = td.keeper.BurnCoins(td.ctx, tt.recipientModule, tt.mintAmount)
			})
		})
	}
}

func TestBurnCoins_Errors(t *testing.T) {
	// returned errors, not panics

	tests := []struct {
		name            string
		recipientModule string
		setupFn         func(td testData)
		mintAmount      sdk.Coins
		wantError       string
	}{
		{
			"invalid coins",
			minttypes.ModuleName,
			func(td testData) {
				// Valid module account minter
				td.ak.EXPECT().
					GetModuleAccount(td.ctx, minttypes.ModuleName).
					Return(authtypes.NewModuleAccount(
						authtypes.NewBaseAccountWithAddress(sdk.AccAddress{1}),
						minttypes.ModuleName,
						// includes minter permission
						authtypes.Minter,
					)).
					Once()
			},
			sdk.Coins{sdk.Coin{
				Denom:  "ukava",
				Amount: sdk.NewInt(-1000),
			}},
			"-1000ukava: invalid coins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewMockedTestData(t)
			tt.setupFn(td)

			require.NotPanics(t, func() {
				err := td.keeper.BurnCoins(td.ctx, tt.recipientModule, tt.mintAmount)

				if tt.wantError != "" {
					require.Error(t, err)
					require.EqualError(t, err, tt.wantError)
					return
				}

				require.NoError(t, err)
			})
		})
	}
}
