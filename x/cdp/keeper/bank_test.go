package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	keep "github.com/kava-labs/kava/x/cdp/keeper"
)

// Test the bank functionality of the CDP keeper
func TestKeeper_AddSubtractGetCoins(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	normalAddr := addrs[0]

	tests := []struct {
		name          string
		address       sdk.AccAddress
		shouldAdd     bool
		amount        sdk.Coins
		expectedCoins sdk.Coins
	}{
		{"addNormalAddress", normalAddr, true, cs(c("usdx", 53)), cs(c("usdx", 153), c("kava", 100))},
		{"subNormalAddress", normalAddr, false, cs(c("usdx", 53)), cs(c("usdx", 47), c("kava", 100))},
		{"addLiquidatorStable", keep.LiquidatorAccountAddress, true, cs(c("usdx", 53)), cs(c("usdx", 153))},
		{"subLiquidatorStable", keep.LiquidatorAccountAddress, false, cs(c("usdx", 53)), cs(c("usdx", 47))},
		{"addLiquidatorGov", keep.LiquidatorAccountAddress, true, cs(c("kava", 53)), cs(c("usdx", 100))},  // no change to balance
		{"subLiquidatorGov", keep.LiquidatorAccountAddress, false, cs(c("kava", 53)), cs(c("usdx", 100))}, // no change to balance
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup app with an account
			tApp := app.NewTestApp()
			tApp.InitializeFromGenesisStates(
				app.NewAuthGenState([]sdk.AccAddress{normalAddr}, []sdk.Coins{cs(c("usdx", 100), c("kava", 100))}),
			)

			// create a new context and setup the liquidator account
			ctx := tApp.NewContext(false, abci.Header{})
			keeper := tApp.GetCDPKeeper()
			keeper.SetLiquidatorModuleAccount(ctx, keep.LiquidatorModuleAccount{cs(c("usdx", 100))}) // set gov coin "balance" to zero

			// perform the test action
			var err sdk.Error
			if tc.shouldAdd {
				_, err = keeper.AddCoins(ctx, tc.address, tc.amount)
			} else {
				_, err = keeper.SubtractCoins(ctx, tc.address, tc.amount)
			}

			// check balances are as expected
			require.NoError(t, err)
			require.Equal(t, tc.expectedCoins, keeper.GetCoins(ctx, tc.address))
		})
	}
}
