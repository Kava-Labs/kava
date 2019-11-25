package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Test the bank functionality of the CDP keeper
func TestKeeper_AddSubtractGetCoins(t *testing.T) {
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
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
		{"addLiquidatorStable", LiquidatorAccountAddress, true, cs(c("usdx", 53)), cs(c("usdx", 153))},
		{"subLiquidatorStable", LiquidatorAccountAddress, false, cs(c("usdx", 53)), cs(c("usdx", 47))},
		{"addLiquidatorGov", LiquidatorAccountAddress, true, cs(c("kava", 53)), cs(c("usdx", 100))},  // no change to balance
		{"subLiquidatorGov", LiquidatorAccountAddress, false, cs(c("kava", 53)), cs(c("usdx", 100))}, // no change to balance
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup keeper
			mapp, keeper, _, _ := setUpMockAppWithoutGenesis()
			// initialize an account with coins
			genAcc := auth.BaseAccount{
				Address: normalAddr,
				Coins:   cs(c("usdx", 100), c("kava", 100)),
			}
			mock.SetGenesis(mapp, []authexported.Account{&genAcc})

			// create a new context and setup the liquidator account
			header := abci.Header{Height: mapp.LastBlockHeight() + 1}
			mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
			ctx := mapp.BaseApp.NewContext(false, header)
			keeper.setLiquidatorModuleAccount(ctx, LiquidatorModuleAccount{cs(c("usdx", 100))}) // set gov coin "balance" to zero

			// perform the test action
			var err sdk.Error
			if tc.shouldAdd {
				_, err = keeper.AddCoins(ctx, tc.address, tc.amount)
			} else {
				_, err = keeper.SubtractCoins(ctx, tc.address, tc.amount)
			}

			mapp.EndBlock(abci.RequestEndBlock{})
			mapp.Commit()

			// check balances are as expected
			require.NoError(t, err)
			require.Equal(t, tc.expectedCoins, keeper.GetCoins(ctx, tc.address))
		})
	}
}
