package keeper_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

// Test suite used for all keeper tests
type PayoutTestSuite struct {
	suite.Suite

	keeper     keeper.Keeper
	hardKeeper hardkeeper.Keeper
	cdpKeeper  cdpkeeper.Keeper

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *PayoutTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *PayoutTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.hardKeeper = suite.app.GetHardKeeper()
	suite.cdpKeeper = suite.app.GetCDPKeeper()

	suite.ctx = suite.app.NewContext(true, tmprototypes.Header{Time: suite.genesisTime})

}

func (suite *PayoutTestSuite) SetupWithGenState(authBuilder app.AuthBankGenesisBuilder, incentBuilder testutil.IncentiveGenesisBuilder, hardBuilder testutil.HardGenesisBuilder) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authBuilder.BuildMarshalled(suite.app.AppCodec()),
		NewPricefeedGenStateMultiFromTime(suite.genesisTime),
		NewCDPGenStateMulti(),
		hardBuilder.BuildMarshalled(),
		incentBuilder.BuildMarshalled(),
	)
}

func (suite *PayoutTestSuite) getAccount(addr sdk.AccAddress) authtypes.AccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *PayoutTestSuite) getModuleAccount(name string) authtypes.ModuleAccountI {
	ak := suite.app.GetAccountKeeper()
	return ak.GetModuleAccount(suite.ctx, name)
}

func (suite *PayoutTestSuite) TestSendCoinsToPeriodicVestingAccount() {
	type accountArgs struct {
		periods          vestingtypes.Periods
		origVestingCoins sdk.Coins
		startTime        int64
		endTime          int64
	}
	type args struct {
		accArgs             accountArgs
		period              vestingtypes.Period
		ctxTime             time.Time
		mintModAccountCoins bool
		expectedPeriods     vestingtypes.Periods
		expectedStartTime   int64
		expectedEndTime     int64
	}
	type errArgs struct {
		expectErr bool
		contains  string
	}
	type testCase struct {
		name    string
		args    args
		errArgs errArgs
	}
	type testCases []testCase

	tests := testCases{
		{
			name: "insert period at beginning schedule",
			args: args{
				accArgs: accountArgs{
					periods: vestingtypes.Periods{
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vestingtypes.Period{Length: 2, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(101, 0),
				mintModAccountCoins: true,
				expectedPeriods: vestingtypes.Periods{
					vestingtypes.Period{Length: 3, Amount: cs(c("ukava", 6))},
					vestingtypes.Period{Length: 2, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "insert period at beginning with new start time",
			args: args{
				accArgs: accountArgs{
					periods: vestingtypes.Periods{
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vestingtypes.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(80, 0),
				mintModAccountCoins: true,
				expectedPeriods: vestingtypes.Periods{
					vestingtypes.Period{Length: 7, Amount: cs(c("ukava", 6))},
					vestingtypes.Period{Length: 18, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 80,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "insert period in middle of schedule",
			args: args{
				accArgs: accountArgs{
					periods: vestingtypes.Periods{
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vestingtypes.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(101, 0),
				mintModAccountCoins: true,
				expectedPeriods: vestingtypes.Periods{
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 3, Amount: cs(c("ukava", 6))},
					vestingtypes.Period{Length: 2, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "append to end of schedule",
			args: args{
				accArgs: accountArgs{
					periods: vestingtypes.Periods{
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vestingtypes.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(125, 0),
				mintModAccountCoins: true,
				expectedPeriods: vestingtypes.Periods{
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 12, Amount: cs(c("ukava", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   132,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "add coins to existing period",
			args: args{
				accArgs: accountArgs{
					periods: vestingtypes.Periods{
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(110, 0),
				mintModAccountCoins: true,
				expectedPeriods: vestingtypes.Periods{
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 11))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
				expectedStartTime: 100,
				expectedEndTime:   120,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
		{
			name: "insufficient mod account balance",
			args: args{
				accArgs: accountArgs{
					periods: vestingtypes.Periods{
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vestingtypes.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(125, 0),
				mintModAccountCoins: false,
				expectedPeriods: vestingtypes.Periods{
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 12, Amount: cs(c("ukava", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   132,
			},
			errArgs: errArgs{
				expectErr: true,
				contains:  "insufficient funds",
			},
		},
		{
			name: "add large period mid schedule",
			args: args{
				accArgs: accountArgs{
					periods: vestingtypes.Periods{
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vestingtypes.Period{Length: 50, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(110, 0),
				mintModAccountCoins: true,
				expectedPeriods: vestingtypes.Periods{
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vestingtypes.Period{Length: 40, Amount: cs(c("ukava", 6))}},
				expectedStartTime: 100,
				expectedEndTime:   160,
			},
			errArgs: errArgs{
				expectErr: false,
				contains:  "",
			},
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			authBuilder := app.NewAuthBankGenesisBuilder().WithSimplePeriodicVestingAccount(
				suite.addrs[0],
				tc.args.accArgs.origVestingCoins,
				tc.args.accArgs.periods,
				tc.args.accArgs.startTime,
			)
			if tc.args.mintModAccountCoins {
				authBuilder = authBuilder.WithSimpleModuleAccount(kavadisttypes.ModuleName, tc.args.period.Amount)
			}

			suite.genesisTime = tc.args.ctxTime
			suite.SetupApp()
			suite.app.InitializeFromGenesisStates(
				authBuilder.BuildMarshalled(suite.app.AppCodec()),
			)

			err := suite.keeper.SendTimeLockedCoinsToPeriodicVestingAccount(suite.ctx, kavadisttypes.ModuleName, suite.addrs[0], tc.args.period.Amount, tc.args.period.Length)

			if tc.errArgs.expectErr {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			} else {
				suite.Require().NoError(err)

				acc := suite.getAccount(suite.addrs[0])
				vacc, ok := acc.(*vestingtypes.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().ElementsMatch(tc.args.expectedPeriods, vacc.VestingPeriods)
				suite.Require().Equal(tc.args.expectedStartTime, vacc.StartTime)
				suite.Require().Equal(tc.args.expectedEndTime, vacc.EndTime)
			}
		})
	}
}

func (suite *PayoutTestSuite) TestSendCoinsToBaseAccount() {
	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleAccount(suite.addrs[1], cs(c("ukava", 400))).
		WithSimpleModuleAccount(kavadisttypes.ModuleName, cs(c("ukava", 600)))

	suite.genesisTime = time.Unix(100, 0)
	suite.SetupApp()
	suite.app.InitializeFromGenesisStates(
		authBuilder.BuildMarshalled(suite.app.AppCodec()),
	)

	// send coins to base account
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadisttypes.ModuleName, suite.addrs[1], cs(c("ukava", 100)), 5)
	suite.Require().NoError(err)
	acc := suite.getAccount(suite.addrs[1])
	vacc, ok := acc.(*vestingtypes.PeriodicVestingAccount)
	suite.True(ok)
	expectedPeriods := vestingtypes.Periods{
		vestingtypes.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	suite.ElementsMatch(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 100)), vacc.OriginalVesting)

	// TODO: BlockTime is 100, same as periodic vesting account start date so
	// it returns nil coins
	suite.Equal(cs(c("ukava", 500)), vacc.GetVestedCoins(suite.ctx.BlockTime()))
	suite.Equal(int64(105), vacc.EndTime)
	suite.Equal(int64(100), vacc.StartTime)
}

func (suite *PayoutTestSuite) TestSendCoinsToInvalidAccount() {
	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleModuleAccount(kavadisttypes.ModuleName, cs(c("ukava", 600)))

	suite.SetupApp()
	suite.app.InitializeFromGenesisStates(
		authBuilder.BuildMarshalled(suite.app.AppCodec()),
	)
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadisttypes.ModuleName, suite.addrs[2], cs(c("ukava", 100)), 5)
	suite.Require().ErrorIs(err, types.ErrInvalidAccountType)
	macc := suite.getModuleAccount(cdptypes.ModuleName)
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadisttypes.ModuleName, macc.GetAddress(), cs(c("ukava", 100)), 5)
	suite.Require().ErrorIs(err, types.ErrInvalidAccountType)
}

func (suite *PayoutTestSuite) TestGetPeriodLength() {
	type args struct {
		blockTime      time.Time
		multiplier     types.Multiplier
		expectedLength int64
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type periodTest struct {
		name    string
		args    args
		errArgs errArgs
	}
	testCases := []periodTest{
		{
			name: "first half of month",
			args: args{
				blockTime:      time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 5, 15, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "first half of month long lockup",
			args: args{
				blockTime:      time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 24, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2022, 11, 15, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 11, 2, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "second half of month",
			args: args{
				blockTime:      time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 7, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "second half of month long lockup",
			args: args{
				blockTime:      time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_LARGE, 24, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 31, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "end of feb",
			args: args{
				blockTime:      time.Date(2021, 2, 28, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 9, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2021, 2, 28, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "leap year",
			args: args{
				blockTime:      time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2020, 9, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "leap year long lockup",
			args: args{
				blockTime:      time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_LARGE, 24, sdk.MustNewDecFromStr("1")),
				expectedLength: time.Date(2022, 3, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 2, 29, 15, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "exactly half of month, is pushed to start of month + lockup",
			args: args{
				blockTime:      time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 7, 1, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "just before half of month",
			args: args{
				blockTime:      time.Date(2020, 12, 15, 13, 59, 59, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 6, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 6, 15, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 15, 13, 59, 59, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "just after start of month payout time, is pushed to mid month + lockup",
			args: args{
				blockTime:      time.Date(2020, 12, 1, 14, 0, 1, 0, time.UTC),
				multiplier:     types.NewMultiplier(types.MULTIPLIER_NAME_MEDIUM, 1, sdk.MustNewDecFromStr("0.333333")),
				expectedLength: time.Date(2021, 1, 15, 14, 0, 0, 0, time.UTC).Unix() - time.Date(2020, 12, 1, 14, 0, 1, 0, time.UTC).Unix(),
			},
			errArgs: errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.genesisTime = tc.args.blockTime
			suite.SetupApp()

			length, err := suite.keeper.GetPeriodLength(suite.ctx, tc.args.multiplier)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.args.expectedLength, length)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func TestPayoutTestSuite(t *testing.T) {
	suite.Run(t, new(PayoutTestSuite))
}
