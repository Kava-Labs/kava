package keeper_test

import (
	"errors"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

func (suite *KeeperTestSuite) setupChain() {
	// creates a new app state with 4 funded addresses and 1 module account
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
		})
	tApp.InitializeFromGenesisStates(
		authGS,
	)
	supplyKeeper := tApp.GetSupplyKeeper()
	macc := supplyKeeper.GetModuleAccount(ctx, kavadist.ModuleName)
	err := supplyKeeper.MintCoins(ctx, macc.GetName(), cs(c("ukava", 600)))
	suite.Require().NoError(err)

	// sets addrs[0] to be a periodic vesting account
	ak := tApp.GetAccountKeeper()
	acc := ak.GetAccount(ctx, addrs[0])
	bacc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	periods := vesting.Periods{
		vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	bva, err2 := vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
	suite.Require().NoError(err2)
	pva := vesting.NewPeriodicVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods)
	ak.SetAccount(ctx, pva)

	// sets addrs[2] to be a validator vesting account
	acc = ak.GetAccount(ctx, addrs[2])
	bacc = auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	bva, err2 = vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
	suite.Require().NoError(err2)
	vva := validatorvesting.NewValidatorVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods, sdk.ConsAddress{}, nil, 90)
	ak.SetAccount(ctx, vva)
	suite.app = tApp
	suite.keeper = tApp.GetIncentiveKeeper()
	suite.ctx = ctx
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) setupExpiredClaims() {
	// creates a new app state with 4 funded addresses
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
		})
	tApp.InitializeFromGenesisStates(
		authGS,
	)

	// creates two claim periods, one expired, and one that expires in the future
	cp1 := types.NewClaimPeriod("bnb", 1, time.Unix(90, 0), time.Hour*8766)
	cp2 := types.NewClaimPeriod("xrp", 1, time.Unix(110, 0), time.Hour*8766)
	suite.keeper = tApp.GetIncentiveKeeper()
	suite.keeper.SetClaimPeriod(ctx, cp1)
	suite.keeper.SetClaimPeriod(ctx, cp2)
	// creates one claim for the non-expired claim period and one claim for the expired claim period
	c1 := types.NewClaim(addrs[0], c("ukava", 1000000), "bnb", 1)
	c2 := types.NewClaim(addrs[0], c("ukava", 1000000), "xrp", 1)
	suite.keeper.SetClaim(ctx, c1)
	suite.keeper.SetClaim(ctx, c2)
	suite.app = tApp
	suite.ctx = ctx
	suite.addrs = addrs
}

func createPeriodicVestingAccount(origVesting sdk.Coins, periods vesting.Periods, startTime, endTime int64) (*vesting.PeriodicVestingAccount, error) {
	_, addr := app.GeneratePrivKeyAddressPairs(1)
	bacc := auth.NewBaseAccountWithAddress(addr[0])
	bacc.Coins = origVesting
	bva, err := vesting.NewBaseVestingAccount(&bacc, origVesting, endTime)
	if err != nil {
		return &vesting.PeriodicVestingAccount{}, err
	}
	pva := vesting.NewPeriodicVestingAccountRaw(bva, startTime, periods)
	err = pva.Validate()
	if err != nil {
		return &vesting.PeriodicVestingAccount{}, err
	}
	return pva, nil
}

func (suite *KeeperTestSuite) TestSendCoinsToPeriodicVestingAccount() {
	type accountArgs struct {
		periods          vesting.Periods
		origVestingCoins sdk.Coins
		startTime        int64
		endTime          int64
	}
	type args struct {
		accArgs             accountArgs
		period              vesting.Period
		ctxTime             time.Time
		mintModAccountCoins bool
		expectedPeriods     vesting.Periods
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
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 2, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(101, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 3, Amount: cs(c("ukava", 6))},
					vesting.Period{Length: 2, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
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
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(80, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
					vesting.Period{Length: 18, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
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
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(101, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 3, Amount: cs(c("ukava", 6))},
					vesting.Period{Length: 2, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
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
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(125, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 12, Amount: cs(c("ukava", 6))}},
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
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 5, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(110, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 11))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
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
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(125, 0),
				mintModAccountCoins: false,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 12, Amount: cs(c("ukava", 6))}},
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
					periods: vesting.Periods{
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
					origVestingCoins: cs(c("ukava", 20)),
					startTime:        100,
					endTime:          120,
				},
				period:              vesting.Period{Length: 50, Amount: cs(c("ukava", 6))},
				ctxTime:             time.Unix(110, 0),
				mintModAccountCoins: true,
				expectedPeriods: vesting.Periods{
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
					vesting.Period{Length: 40, Amount: cs(c("ukava", 6))}},
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
			// create the periodic vesting account
			pva, err := createPeriodicVestingAccount(tc.args.accArgs.origVestingCoins, tc.args.accArgs.periods, tc.args.accArgs.startTime, tc.args.accArgs.endTime)
			suite.Require().NoError(err)

			// setup store state with account and kavadist module account
			suite.ctx = suite.ctx.WithBlockTime(tc.args.ctxTime)
			ak := suite.app.GetAccountKeeper()
			ak.SetAccount(suite.ctx, pva)
			// mint module account coins if required
			if tc.args.mintModAccountCoins {
				sk := suite.app.GetSupplyKeeper()
				err = sk.MintCoins(suite.ctx, kavadist.ModuleName, tc.args.period.Amount)
				suite.Require().NoError(err)
			}

			err = suite.keeper.SendTimeLockedCoinsToPeriodicVestingAccount(suite.ctx, kavadist.ModuleName, pva.Address, tc.args.period.Amount, tc.args.period.Length)
			if tc.errArgs.expectErr {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			} else {
				suite.Require().NoError(err)

				acc := suite.getAccount(pva.Address)
				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				suite.Require().True(ok)
				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
				suite.Require().Equal(tc.args.expectedStartTime, vacc.StartTime)
				suite.Require().Equal(tc.args.expectedEndTime, vacc.EndTime)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSendCoinsToBaseAccount() {
	suite.setupChain()
	// send coins to base account
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[1], cs(c("ukava", 100)), 5)
	suite.Require().NoError(err)
	acc := suite.getAccount(suite.addrs[1])
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	expectedPeriods := vesting.Periods{
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 100)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 500)), vacc.Coins)
	suite.Equal(int64(105), vacc.EndTime)
	suite.Equal(int64(100), vacc.StartTime)

}

func (suite *KeeperTestSuite) TestSendCoinsToInvalidAccount() {
	suite.setupChain()
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[2], cs(c("ukava", 100)), 5)
	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
	macc := suite.getModuleAccount(cdp.ModuleName)
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, macc.GetAddress(), cs(c("ukava", 100)), 5)
	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
}

func (suite *KeeperTestSuite) TestPayoutClaim() {
	suite.setupChain() // adds 3 accounts - 1 periodic vesting account, 1 base account, and 1 validator vesting account

	// add 2 claims that correspond to an existing claim period and one claim that has no corresponding claim period
	cp1 := types.NewClaimPeriod("bnb", 1, suite.ctx.BlockTime().Add(time.Hour*168), time.Hour*8766)
	suite.keeper.SetClaimPeriod(suite.ctx, cp1)
	// valid claim for addrs[0]
	c1 := types.NewClaim(suite.addrs[0], c("ukava", 100), "bnb", 1)
	// invalid claim for addrs[0]
	c2 := types.NewClaim(suite.addrs[0], c("ukava", 100), "xrp", 1)
	// valid claim for addrs[1]
	c3 := types.NewClaim(suite.addrs[1], c("ukava", 100), "bnb", 1)
	suite.keeper.SetClaim(suite.ctx, c1)
	suite.keeper.SetClaim(suite.ctx, c2)
	suite.keeper.SetClaim(suite.ctx, c3)

	// existing claim with corresponding claim period successfully claimed by existing periodic vesting account
	err := suite.keeper.PayoutClaim(suite.ctx.WithBlockTime(time.Unix(3700, 0)), suite.addrs[0], "bnb", 1)
	suite.Require().NoError(err)
	acc := suite.getAccount(suite.addrs[0])
	// account is a periodic vesting account
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	// vesting balance is correct
	suite.Equal(cs(c("ukava", 500)), vacc.OriginalVesting)

	// existing claim with corresponding claim period successfully claimed by base account
	err = suite.keeper.PayoutClaim(suite.ctx, suite.addrs[1], "bnb", 1)
	suite.Require().NoError(err)
	acc = suite.getAccount(suite.addrs[1])
	// account has become a periodic vesting account
	vacc, ok = acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	// vesting balance is correct
	suite.Equal(cs(c("ukava", 100)), vacc.OriginalVesting)

	// addrs[3] has no claims
	err = suite.keeper.PayoutClaim(suite.ctx, suite.addrs[3], "bnb", 1)
	suite.Require().True(errors.Is(err, types.ErrClaimNotFound))
	// addrs[0] has an xrp claim, but there is not corresponding claim period
	err = suite.keeper.PayoutClaim(suite.ctx, suite.addrs[0], "xrp", 1)
	suite.Require().True(errors.Is(err, types.ErrClaimPeriodNotFound))
}

func (suite *KeeperTestSuite) TestDeleteExpiredClaimPeriods() {
	suite.setupExpiredClaims() // creates new app state with one non-expired claim period (xrp) and one expired claim period (bnb) as well  as a claim that corresponds to each claim period

	// both claim periods are present
	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.True(found)
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "xrp")
	suite.True(found)
	// both claims are present
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.True(found)
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "xrp", 1)
	suite.True(found)

	// expired claim period and associated claims should get deleted
	suite.NotPanics(func() {
		suite.keeper.DeleteExpiredClaimsAndClaimPeriods(suite.ctx)
	})
	// expired claim period and claim are not found
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.False(found)
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.False(found)
	// non-expired claim period and claim are found
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "xrp")
	suite.True(found)
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "xrp", 1)
	suite.True(found)

}
