package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard/types"
)

func (suite *KeeperTestSuite) TestSendTimeLockedCoinsToAccount() {
	type accountArgs struct {
		addr                 sdk.AccAddress
		vestingAccountBefore bool
		vestingAccountAfter  bool
		coins                sdk.Coins
		periods              []vestingtypes.Period
		origVestingCoins     sdk.Coins
		startTime            int64
		endTime              int64
	}
	type args struct {
		accArgs                   accountArgs
		period                    vestingtypes.Period
		blockTime                 time.Time
		expectedAccountBalance    sdk.Coins
		expectedModAccountBalance sdk.Coins
		expectedPeriods           []vestingtypes.Period
		expectedStartTime         int64
		expectedEndTime           int64
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type testCase struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []testCase{
		{
			name: "send liquid coins to base account",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: false,
					vestingAccountAfter:  false,
					coins:                sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
				},
				period:                    vestingtypes.Period{Length: 0, Amount: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)))},
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)), sdk.NewCoin("hard", sdk.NewInt(100))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedPeriods:           []vestingtypes.Period{},
				expectedStartTime:         0,
				expectedEndTime:           0,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "send liquid coins to vesting account",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
					periods: []vestingtypes.Period{
						{Amount: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))), Length: 100},
					},
					origVestingCoins: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
					startTime:        time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix(),
					endTime:          time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix() + 100,
				},
				period:                    vestingtypes.Period{Length: 0, Amount: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(100)))},
				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000)), sdk.NewCoin("hard", sdk.NewInt(100))),
				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(900))),
				expectedPeriods: []vestingtypes.Period{
					{Amount: sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))), Length: 100},
				},
				expectedStartTime: time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix(),
				expectedEndTime:   time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix() + 100,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "insert period at beginning of schedule",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: []vestingtypes.Period{
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vestingtypes.Period{Length: 2, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(101, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: []vestingtypes.Period{
					{Length: 3, Amount: cs(c("hard", 6))},
					{Length: 2, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "insert period at beginning with new start time",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: []vestingtypes.Period{
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vestingtypes.Period{Length: 7, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(80, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: []vestingtypes.Period{
					{Length: 7, Amount: cs(c("hard", 6))},
					{Length: 18, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 80,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "insert period in middle of schedule",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: []vestingtypes.Period{
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vestingtypes.Period{Length: 7, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(101, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: []vestingtypes.Period{
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 3, Amount: cs(c("hard", 6))},
					{Length: 2, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "append to end of schedule",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: []vestingtypes.Period{
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vestingtypes.Period{Length: 7, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(125, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: []vestingtypes.Period{
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 12, Amount: cs(c("hard", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   132,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "add coins to existing period",
			args: args{
				accArgs: accountArgs{
					addr:                 sdk.AccAddress(crypto.AddressHash([]byte("test"))),
					vestingAccountBefore: true,
					vestingAccountAfter:  true,
					coins:                cs(c("bnb", 20)),
					periods: []vestingtypes.Period{
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))},
						{Length: 5, Amount: cs(c("bnb", 5))}},
					origVestingCoins: cs(c("bnb", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:                    vestingtypes.Period{Length: 5, Amount: cs(c("hard", 6))},
				blockTime:                 time.Unix(110, 0),
				expectedAccountBalance:    cs(c("bnb", 20), c("hard", 6)),
				expectedModAccountBalance: cs(c("hard", 994)),
				expectedPeriods: []vestingtypes.Period{
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5))},
					{Length: 5, Amount: cs(c("bnb", 5), c("hard", 6))},
					{Length: 5, Amount: cs(c("bnb", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// create new app with one funded account

			// Initialize test app and set context
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tc.args.blockTime})
			authGS := app.NewFundedGenStateWithCoins(
				tApp.AppCodec(),
				[]sdk.Coins{tc.args.accArgs.coins},
				[]sdk.AccAddress{tc.args.accArgs.addr},
			)
			loanToValue := sdk.MustNewDecFromStr("0.6")
			hardGS := types.NewGenesisState(types.NewParams(
				types.MoneyMarkets{
					types.NewMoneyMarket("usdx", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "usdx:usd", sdk.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
					types.NewMoneyMarket("ukava", types.NewBorrowLimit(false, sdk.NewDec(1000000000000000), loanToValue), "kava:usd", sdk.NewInt(1000000), types.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				},
				sdk.NewDec(10),
			), types.DefaultAccumulationTimes, types.DefaultDeposits, types.DefaultBorrows,
				types.DefaultTotalSupplied, types.DefaultTotalBorrowed, types.DefaultTotalReserves,
			)
			tApp.InitializeFromGenesisStates(authGS, app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(&hardGS)})
			if tc.args.accArgs.vestingAccountBefore {
				ak := tApp.GetAccountKeeper()
				acc := ak.GetAccount(ctx, tc.args.accArgs.addr)

				bacc := authtypes.NewBaseAccount(acc.GetAddress(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
				bva := vestingtypes.NewBaseVestingAccount(bacc, tc.args.accArgs.origVestingCoins, tc.args.accArgs.endTime)
				// TODO: check bal here?
				pva := vestingtypes.NewPeriodicVestingAccountRaw(bva, tc.args.accArgs.startTime, tc.args.accArgs.periods)
				ak.SetAccount(ctx, pva)
			}
			bankKeeper := tApp.GetBankKeeper()
			err := bankKeeper.MintCoins(ctx, types.ModuleAccountName, sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(1000))))
			suite.Require().NoError(err)

			keeper := tApp.GetHardKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper

			err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, types.ModuleAccountName, tc.args.accArgs.addr, tc.args.period.Amount, tc.args.period.Length)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				acc := suite.getAccount(tc.args.accArgs.addr)
				suite.Require().Equal(tc.args.expectedAccountBalance, bankKeeper.GetAllBalances(ctx, acc.GetAddress()))
				mAcc := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(tc.args.expectedModAccountBalance, bankKeeper.GetAllBalances(ctx, mAcc.GetAddress()))
				vacc, ok := acc.(*vestingtypes.PeriodicVestingAccount)
				if tc.args.accArgs.vestingAccountAfter {
					suite.Require().True(ok)
					suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
					suite.Require().Equal(tc.args.expectedStartTime, vacc.StartTime)
					suite.Require().Equal(tc.args.expectedEndTime, vacc.EndTime)
				} else {
					suite.Require().False(ok)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}

		})
	}
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
