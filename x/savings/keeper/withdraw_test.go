package keeper_test

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/savings/types"
)

func (suite *KeeperTestSuite) TestWithdraw() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, delegator := addrs[0], addrs[1]

	valAddr := sdk.ValAddress(valAccAddr)
	initialBalance := sdkmath.NewInt(1e9)

	bkavaDenom := fmt.Sprintf("bkava-%s", valAddr.String())

	type args struct {
		allowedDenoms             []string
		depositor                 sdk.AccAddress
		initialDepositorBalance   sdk.Coins
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
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(1000)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(900)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(100))),
			},
			errArgs{
				expectPass:   true,
				expectDelete: false,
				contains:     "",
			},
		},
		{
			"valid: partial bkava",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava", "bkava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdkmath.NewInt(1000)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdkmath.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdkmath.NewInt(100))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdkmath.NewInt(900)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdkmath.NewInt(100))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdkmath.NewInt(100))),
			},
			errArgs{
				expectPass:   true,
				expectDelete: false,
				contains:     "",
			},
		},
		{
			"valid: full withdraw",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(1000)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(1000)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
			},
			errArgs{
				expectPass:   true,
				expectDelete: true,
				contains:     "",
			},
		},
		{
			"valid: withdraw exceeds deposit but is adjusted to match max deposit",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(1000)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(300))),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(1000)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
			},
			errArgs{
				expectPass:   true,
				expectDelete: true,
				contains:     "",
			},
		},
		{
			"invalid: withdraw non-supplied coin type",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(1000)), sdk.NewCoin("btcb", sdkmath.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
				withdrawAmount:            sdk.NewCoins(sdk.NewCoin("btcb", sdkmath.NewInt(200))),
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(200))),
				expectedDepositCoins:      sdk.Coins{},
			},
			errArgs{
				expectPass:   false,
				expectDelete: false,
				contains:     "invalid withdraw denom",
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
			savingsGS := types.NewGenesisState(
				types.NewParams(tc.args.allowedDenoms),
				types.Deposits{},
			)

			stakingParams := stakingtypes.DefaultParams()
			stakingParams.BondDenom = "ukava"

			tApp.InitializeFromGenesisStates(authGS,
				app.GenesisState{types.ModuleName: tApp.AppCodec().MustMarshalJSON(&savingsGS)},
				app.GenesisState{stakingtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(stakingtypes.NewGenesisState(stakingParams, nil, nil))},
			)
			keeper := tApp.GetSavingsKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper
			bankKeeper := tApp.GetBankKeeper()

			// Create validator and delegate for bkava
			suite.CreateAccountWithAddress(valAccAddr, cs(c("ukava", 100e10)))
			suite.CreateAccountWithAddress(delegator, cs(c("ukava", 100e10)))

			suite.CreateNewUnbondedValidator(valAddr, initialBalance)
			suite.CreateDelegation(valAddr, delegator, initialBalance)
			staking.EndBlocker(suite.ctx, suite.app.GetStakingKeeper())

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
