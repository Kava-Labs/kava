package keeper_test

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/savings/types"
)

func (suite *KeeperTestSuite) TestWithdraw() {
	type args struct {
		allowedDenoms             []string
		depositor                 sdk.AccAddress
		initialDepositorBalance   sdk.Coins
		initialModAccountBalance  sdk.Coins
		depositAmount             sdk.Coins
		withdrawAmount            sdk.Coins
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		expectedDepositCoins      sdk.Coins
	}
	type errArgs struct {
		expectPass   bool
		expectDelete bool
		contains     string
	}
	type withdrawTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []withdrawTest{
		{
			"valid: partial withdraw",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				initialModAccountBalance:  sdk.Coins(nil),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(900)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
			},
			errArgs{
				expectPass:   true,
				expectDelete: false,
				contains:     "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
			authGS := app.NewFundedGenStateWithCoins(
				tApp.AppCodec(),
				[]sdk.Coins{tc.args.initialDepositorBalance},
				[]sdk.AccAddress{tc.args.depositor},
			)
			savingsGS := types.NewGenesisState(types.NewParams(tc.args.allowedDenoms))

			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{types.ModuleName: tApp.AppCodec().MustMarshalJSON(&savingsGS)},
			)
			keeper := tApp.GetSavingsKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			// Mint coins to savings module account
			bankKeeper := tApp.GetBankKeeper()
			// err := bankKeeper.MintCoins(ctx, types.ModuleAccountName, tc.args.initialModAccountBalance)
			// suite.Require().NoError(err)

			err := suite.keeper.Deposit(suite.ctx, tc.args.depositor, tc.args.depositAmount)
			suite.Require().NoError(err)

			err = suite.keeper.Withdraw(suite.ctx, tc.args.depositor, tc.args.withdrawAmount)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				// Check depositor's account balance
				acc := suite.getAccount(tc.args.depositor)
				suite.Require().Equal(tc.args.expectedAccountBalance, bankKeeper.GetAllBalances(ctx, acc.GetAddress()))
				// Check savings module account balance
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().True(tc.args.expectedModAccountBalance.IsEqual(bankKeeper.GetAllBalances(ctx, mAcc.GetAddress())))
				// Check deposit
				testDeposit, f := suite.keeper.GetDeposit(suite.ctx, tc.args.depositor)
				if tc.errArgs.expectDelete {
					suite.Require().False(f)
				} else {
					suite.Require().True(f)
					suite.Require().Equal(tc.args.expectedDepositCoins, testDeposit.Amount)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
