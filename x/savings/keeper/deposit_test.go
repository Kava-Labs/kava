package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/savings/types"
)

func (suite *KeeperTestSuite) TestDeposit() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, delegator := addrs[0], addrs[1]

	valAddr := sdk.ValAddress(valAccAddr)
	initialBalance := sdk.NewInt(1e9)

	bkavaDenom := fmt.Sprintf("bkava-%s", valAddr.String())
	invalidBkavaDenom := fmt.Sprintf("bkava-%s", sdk.ValAddress(addrs[2]).String())

	type args struct {
		allowedDenoms             []string
		depositor                 sdk.AccAddress
		initialDepositorBalance   sdk.Coins
		depositAmount             sdk.Coins
		numberDeposits            int
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		expectedDepositCoins      sdk.Coins
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type depositTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []depositTest{
		{
			"valid",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(900)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid multi deposit",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
				numberDeposits:            2,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(800)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(200))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid bkava",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava", "bkava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdk.NewInt(100))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdk.NewInt(900)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdk.NewInt(100))),
				expectedDepositCoins:      sdk.NewCoins(sdk.NewCoin(bkavaDenom, sdk.NewInt(100))),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid deposit denom",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("fake", sdk.NewInt(100))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
			},
			errArgs{
				expectPass: false,
				contains:   "invalid deposit denom",
			},
		},
		{
			"invalid bkava",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava", "bkava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin(invalidBkavaDenom, sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin(invalidBkavaDenom, sdk.NewInt(100))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
			},
			errArgs{
				expectPass: false,
				contains:   "invalid deposit denom",
			},
		},
		{
			"insufficient funds",
			args{
				allowedDenoms:             []string{"bnb", "btcb", "ukava"},
				depositor:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
				initialDepositorBalance:   sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				depositAmount:             sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10000))),
				numberDeposits:            1,
				expectedAccountBalance:    sdk.Coins{},
				expectedModAccountBalance: sdk.Coins{},
				expectedDepositCoins:      sdk.Coins{},
			},
			errArgs{
				expectPass: false,
				contains:   "insufficient funds",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account

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

			// Create validator and delegate for bkava
			suite.CreateAccountWithAddress(valAccAddr, cs(c("ukava", 100e10)))
			suite.CreateAccountWithAddress(delegator, cs(c("ukava", 100e10)))

			suite.CreateNewUnbondedValidator(valAddr, initialBalance)
			suite.CreateDelegation(valAddr, delegator, initialBalance)
			staking.EndBlocker(suite.ctx, suite.app.GetStakingKeeper())

			// run the test
			var err error
			for i := 0; i < tc.args.numberDeposits; i++ {
				err = suite.keeper.Deposit(suite.ctx, tc.args.depositor, tc.args.depositAmount)
			}

			// verify results
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.depositor)
				suite.Require().Equal(tc.args.expectedAccountBalance, suite.getAccountCoins(acc))
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, suite.getAccountCoins(mAcc))
				dep, f := suite.keeper.GetDeposit(suite.ctx, tc.args.depositor)
				suite.Require().True(f)
				suite.Require().Equal(tc.args.expectedDepositCoins, dep.Amount)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
