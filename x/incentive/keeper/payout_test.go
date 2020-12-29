package keeper_test

// import (
// 	"errors"
// 	"strings"
// 	"time"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/x/auth"
// 	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

// 	abci "github.com/tendermint/tendermint/abci/types"

// 	"github.com/kava-labs/kava/app"
// 	"github.com/kava-labs/kava/x/cdp"
// 	"github.com/kava-labs/kava/x/incentive/types"
// 	"github.com/kava-labs/kava/x/kavadist"
// 	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
// 	"github.com/tendermint/tendermint/crypto"
// )

// func (suite *KeeperTestSuite) setupChain() {
// 	// creates a new app state with 4 funded addresses and 1 module account
// 	tApp := app.NewTestApp()
// 	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})
// 	_, addrs := app.GeneratePrivKeyAddressPairs(4)
// 	authGS := app.NewAuthGenState(
// 		addrs,
// 		[]sdk.Coins{
// 			cs(c("ukava", 400)),
// 			cs(c("ukava", 400)),
// 			cs(c("ukava", 400)),
// 			cs(c("ukava", 400)),
// 		})
// 	tApp.InitializeFromGenesisStates(
// 		authGS,
// 	)
// 	supplyKeeper := tApp.GetSupplyKeeper()
// 	macc := supplyKeeper.GetModuleAccount(ctx, kavadist.ModuleName)
// 	err := supplyKeeper.MintCoins(ctx, macc.GetName(), cs(c("ukava", 600)))
// 	suite.Require().NoError(err)

// 	// sets addrs[0] to be a periodic vesting account
// 	ak := tApp.GetAccountKeeper()
// 	acc := ak.GetAccount(ctx, addrs[0])
// 	bacc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
// 	periods := vesting.Periods{
// 		vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
// 		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
// 		vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
// 		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
// 	}
// 	bva, err2 := vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
// 	suite.Require().NoError(err2)
// 	pva := vesting.NewPeriodicVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods)
// 	ak.SetAccount(ctx, pva)

// 	// sets addrs[2] to be a validator vesting account
// 	acc = ak.GetAccount(ctx, addrs[2])
// 	bacc = auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
// 	bva, err2 = vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
// 	suite.Require().NoError(err2)
// 	vva := validatorvesting.NewValidatorVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods, sdk.ConsAddress{}, nil, 90)
// 	ak.SetAccount(ctx, vva)
// 	suite.app = tApp
// 	suite.keeper = tApp.GetIncentiveKeeper()
// 	suite.ctx = ctx
// 	suite.addrs = addrs
// }

// func (suite *KeeperTestSuite) setupExpiredClaims() {
// 	// creates a new app state with 4 funded addresses
// 	tApp := app.NewTestApp()
// 	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})
// 	_, addrs := app.GeneratePrivKeyAddressPairs(4)
// 	authGS := app.NewAuthGenState(
// 		addrs,
// 		[]sdk.Coins{
// 			cs(c("ukava", 400)),
// 			cs(c("ukava", 400)),
// 			cs(c("ukava", 400)),
// 			cs(c("ukava", 400)),
// 		})
// 	tApp.InitializeFromGenesisStates(
// 		authGS,
// 	)

// 	// creates two claim periods, one expired, and one that expires in the future
// 	cp1 := types.NewClaimPeriod("bnb", 1, time.Unix(90, 0), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})
// 	cp2 := types.NewClaimPeriod("xrp", 1, time.Unix(110, 0), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})
// 	suite.keeper = tApp.GetIncentiveKeeper()
// 	suite.keeper.SetClaimPeriod(ctx, cp1)
// 	suite.keeper.SetClaimPeriod(ctx, cp2)
// 	// creates one claim for the non-expired claim period and one claim for the expired claim period
// 	c1 := types.NewClaim(addrs[0], c("ukava", 1000000), "bnb", 1)
// 	c2 := types.NewClaim(addrs[0], c("ukava", 1000000), "xrp", 1)
// 	suite.keeper.SetClaim(ctx, c1)
// 	suite.keeper.SetClaim(ctx, c2)
// 	suite.app = tApp
// 	suite.ctx = ctx
// 	suite.addrs = addrs
// }

// func createPeriodicVestingAccount(origVesting sdk.Coins, periods vesting.Periods, startTime, endTime int64) (*vesting.PeriodicVestingAccount, error) {
// 	_, addr := app.GeneratePrivKeyAddressPairs(1)
// 	bacc := auth.NewBaseAccountWithAddress(addr[0])
// 	bacc.Coins = origVesting
// 	bva, err := vesting.NewBaseVestingAccount(&bacc, origVesting, endTime)
// 	if err != nil {
// 		return &vesting.PeriodicVestingAccount{}, err
// 	}
// 	pva := vesting.NewPeriodicVestingAccountRaw(bva, startTime, periods)
// 	err = pva.Validate()
// 	if err != nil {
// 		return &vesting.PeriodicVestingAccount{}, err
// 	}
// 	return pva, nil
// }

// func (suite *KeeperTestSuite) TestSendCoinsToPeriodicVestingAccount() {
// 	type accountArgs struct {
// 		periods          vesting.Periods
// 		origVestingCoins sdk.Coins
// 		startTime        int64
// 		endTime          int64
// 	}
// 	type args struct {
// 		accArgs             accountArgs
// 		period              vesting.Period
// 		ctxTime             time.Time
// 		mintModAccountCoins bool
// 		expectedPeriods     vesting.Periods
// 		expectedStartTime   int64
// 		expectedEndTime     int64
// 	}
// 	type errArgs struct {
// 		expectErr bool
// 		contains  string
// 	}
// 	type testCase struct {
// 		name    string
// 		args    args
// 		errArgs errArgs
// 	}
// 	type testCases []testCase

// 	tests := testCases{
// 		{
// 			name: "insert period at beginning schedule",
// 			args: args{
// 				accArgs: accountArgs{
// 					periods: vesting.Periods{
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 					origVestingCoins: cs(c("ukava", 20)),
// 					startTime:        100,
// 					endTime:          120,
// 				},
// 				period:              vesting.Period{Length: 2, Amount: cs(c("ukava", 6))},
// 				ctxTime:             time.Unix(101, 0),
// 				mintModAccountCoins: true,
// 				expectedPeriods: vesting.Periods{
// 					vesting.Period{Length: 3, Amount: cs(c("ukava", 6))},
// 					vesting.Period{Length: 2, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 				expectedStartTime: 100,
// 				expectedEndTime:   120,
// 			},
// 			errArgs: errArgs{
// 				expectErr: false,
// 				contains:  "",
// 			},
// 		},
// 		{
// 			name: "insert period at beginning with new start time",
// 			args: args{
// 				accArgs: accountArgs{
// 					periods: vesting.Periods{
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 					origVestingCoins: cs(c("ukava", 20)),
// 					startTime:        100,
// 					endTime:          120,
// 				},
// 				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
// 				ctxTime:             time.Unix(80, 0),
// 				mintModAccountCoins: true,
// 				expectedPeriods: vesting.Periods{
// 					vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
// 					vesting.Period{Length: 18, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 				expectedStartTime: 80,
// 				expectedEndTime:   120,
// 			},
// 			errArgs: errArgs{
// 				expectErr: false,
// 				contains:  "",
// 			},
// 		},
// 		{
// 			name: "insert period in middle of schedule",
// 			args: args{
// 				accArgs: accountArgs{
// 					periods: vesting.Periods{
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 					origVestingCoins: cs(c("ukava", 20)),
// 					startTime:        100,
// 					endTime:          120,
// 				},
// 				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
// 				ctxTime:             time.Unix(101, 0),
// 				mintModAccountCoins: true,
// 				expectedPeriods: vesting.Periods{
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 3, Amount: cs(c("ukava", 6))},
// 					vesting.Period{Length: 2, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 				expectedStartTime: 100,
// 				expectedEndTime:   120,
// 			},
// 			errArgs: errArgs{
// 				expectErr: false,
// 				contains:  "",
// 			},
// 		},
// 		{
// 			name: "append to end of schedule",
// 			args: args{
// 				accArgs: accountArgs{
// 					periods: vesting.Periods{
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 					origVestingCoins: cs(c("ukava", 20)),
// 					startTime:        100,
// 					endTime:          120,
// 				},
// 				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
// 				ctxTime:             time.Unix(125, 0),
// 				mintModAccountCoins: true,
// 				expectedPeriods: vesting.Periods{
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 12, Amount: cs(c("ukava", 6))}},
// 				expectedStartTime: 100,
// 				expectedEndTime:   132,
// 			},
// 			errArgs: errArgs{
// 				expectErr: false,
// 				contains:  "",
// 			},
// 		},
// 		{
// 			name: "add coins to existing period",
// 			args: args{
// 				accArgs: accountArgs{
// 					periods: vesting.Periods{
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 					origVestingCoins: cs(c("ukava", 20)),
// 					startTime:        100,
// 					endTime:          120,
// 				},
// 				period:              vesting.Period{Length: 5, Amount: cs(c("ukava", 6))},
// 				ctxTime:             time.Unix(110, 0),
// 				mintModAccountCoins: true,
// 				expectedPeriods: vesting.Periods{
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 11))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 				expectedStartTime: 100,
// 				expectedEndTime:   120,
// 			},
// 			errArgs: errArgs{
// 				expectErr: false,
// 				contains:  "",
// 			},
// 		},
// 		{
// 			name: "insufficient mod account balance",
// 			args: args{
// 				accArgs: accountArgs{
// 					periods: vesting.Periods{
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 					origVestingCoins: cs(c("ukava", 20)),
// 					startTime:        100,
// 					endTime:          120,
// 				},
// 				period:              vesting.Period{Length: 7, Amount: cs(c("ukava", 6))},
// 				ctxTime:             time.Unix(125, 0),
// 				mintModAccountCoins: false,
// 				expectedPeriods: vesting.Periods{
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 12, Amount: cs(c("ukava", 6))}},
// 				expectedStartTime: 100,
// 				expectedEndTime:   132,
// 			},
// 			errArgs: errArgs{
// 				expectErr: true,
// 				contains:  "insufficient funds",
// 			},
// 		},
// 		{
// 			name: "add large period mid schedule",
// 			args: args{
// 				accArgs: accountArgs{
// 					periods: vesting.Periods{
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 						vesting.Period{Length: 5, Amount: cs(c("ukava", 5))}},
// 					origVestingCoins: cs(c("ukava", 20)),
// 					startTime:        100,
// 					endTime:          120,
// 				},
// 				period:              vesting.Period{Length: 50, Amount: cs(c("ukava", 6))},
// 				ctxTime:             time.Unix(110, 0),
// 				mintModAccountCoins: true,
// 				expectedPeriods: vesting.Periods{
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 5, Amount: cs(c("ukava", 5))},
// 					vesting.Period{Length: 40, Amount: cs(c("ukava", 6))}},
// 				expectedStartTime: 100,
// 				expectedEndTime:   160,
// 			},
// 			errArgs: errArgs{
// 				expectErr: false,
// 				contains:  "",
// 			},
// 		},
// 	}
// 	for _, tc := range tests {
// 		suite.Run(tc.name, func() {
// 			// create the periodic vesting account
// 			pva, err := createPeriodicVestingAccount(tc.args.accArgs.origVestingCoins, tc.args.accArgs.periods, tc.args.accArgs.startTime, tc.args.accArgs.endTime)
// 			suite.Require().NoError(err)

// 			// setup store state with account and kavadist module account
// 			suite.ctx = suite.ctx.WithBlockTime(tc.args.ctxTime)
// 			ak := suite.app.GetAccountKeeper()
// 			ak.SetAccount(suite.ctx, pva)
// 			// mint module account coins if required
// 			if tc.args.mintModAccountCoins {
// 				sk := suite.app.GetSupplyKeeper()
// 				err = sk.MintCoins(suite.ctx, kavadist.ModuleName, tc.args.period.Amount)
// 				suite.Require().NoError(err)
// 			}

// 			err = suite.keeper.SendTimeLockedCoinsToPeriodicVestingAccount(suite.ctx, kavadist.ModuleName, pva.Address, tc.args.period.Amount, tc.args.period.Length)
// 			if tc.errArgs.expectErr {
// 				suite.Require().Error(err)
// 				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
// 			} else {
// 				suite.Require().NoError(err)

// 				acc := suite.getAccount(pva.Address)
// 				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
// 				suite.Require().True(ok)
// 				suite.Require().Equal(tc.args.expectedPeriods, vacc.VestingPeriods)
// 				suite.Require().Equal(tc.args.expectedStartTime, vacc.StartTime)
// 				suite.Require().Equal(tc.args.expectedEndTime, vacc.EndTime)
// 			}
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestSendCoinsToBaseAccount() {
// 	suite.setupChain()
// 	// send coins to base account
// 	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[1], cs(c("ukava", 100)), 5)
// 	suite.Require().NoError(err)
// 	acc := suite.getAccount(suite.addrs[1])
// 	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
// 	suite.True(ok)
// 	expectedPeriods := vesting.Periods{
// 		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
// 	}
// 	suite.Equal(expectedPeriods, vacc.VestingPeriods)
// 	suite.Equal(cs(c("ukava", 100)), vacc.OriginalVesting)
// 	suite.Equal(cs(c("ukava", 500)), vacc.Coins)
// 	suite.Equal(int64(105), vacc.EndTime)
// 	suite.Equal(int64(100), vacc.StartTime)

// }

// func (suite *KeeperTestSuite) TestSendCoinsToInvalidAccount() {
// 	suite.setupChain()
// 	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[2], cs(c("ukava", 100)), 5)
// 	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
// 	macc := suite.getModuleAccount(cdp.ModuleName)
// 	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, macc.GetAddress(), cs(c("ukava", 100)), 5)
// 	suite.Require().True(errors.Is(err, types.ErrInvalidAccountType))
// }

// func (suite *KeeperTestSuite) TestDeleteExpiredClaimPeriods() {
// 	suite.setupExpiredClaims() // creates new app state with one non-expired claim period (xrp) and one expired claim period (bnb) as well  as a claim that corresponds to each claim period

// 	// both claim periods are present
// 	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
// 	suite.True(found)
// 	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "xrp")
// 	suite.True(found)
// 	// both claims are present
// 	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
// 	suite.True(found)
// 	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "xrp", 1)
// 	suite.True(found)

// 	// expired claim period and associated claims should get deleted
// 	suite.NotPanics(func() {
// 		suite.keeper.DeleteExpiredClaimsAndClaimPeriods(suite.ctx)
// 	})
// 	// expired claim period and claim are not found
// 	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
// 	suite.False(found)
// 	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
// 	suite.False(found)
// 	// non-expired claim period and claim are found
// 	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "xrp")
// 	suite.True(found)
// 	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "xrp", 1)
// 	suite.True(found)

// }

// func (suite *KeeperTestSuite) TestPayoutClaim() {
// 	type args struct {
// 		claimOwner                sdk.AccAddress
// 		collateralType            string
// 		id                        uint64
// 		multiplier                types.MultiplierName
// 		blockTime                 time.Time
// 		rewards                   types.Rewards
// 		rewardperiods             types.RewardPeriods
// 		claimPeriods              types.ClaimPeriods
// 		claims                    types.Claims
// 		genIDs                    types.GenesisClaimPeriodIDs
// 		active                    bool
// 		validatorVesting          bool
// 		expectedAccountBalance    sdk.Coins
// 		expectedModAccountBalance sdk.Coins
// 		expectedVestingAccount    bool
// 		expectedVestingLength     int64
// 	}
// 	type errArgs struct {
// 		expectPass bool
// 		contains   string
// 	}
// 	type claimTest struct {
// 		name    string
// 		args    args
// 		errArgs errArgs
// 	}
// 	testCases := []claimTest{
// 		{
// 			"valid small claim",
// 			args{
// 				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
// 				collateralType:            "bnb-a",
// 				id:                        1,
// 				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
// 				rewards:                   types.Rewards{types.NewReward(true, "bnb-a", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)},
// 				rewardperiods:             types.RewardPeriods{types.NewRewardPeriod("bnb-a", time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), c("ukava", 1000), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24*2), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claimPeriods:              types.ClaimPeriods{types.NewClaimPeriod("bnb-a", 1, time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claims:                    types.Claims{types.NewClaim(sdk.AccAddress(crypto.AddressHash([]byte("test"))), sdk.NewCoin("ukava", sdk.NewInt(1000)), "bnb-a", 1)},
// 				genIDs:                    types.GenesisClaimPeriodIDs{types.GenesisClaimPeriodID{CollateralType: "bnb-a", ID: 2}},
// 				active:                    true,
// 				validatorVesting:          false,
// 				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
// 				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500))),
// 				expectedVestingAccount:    true,
// 				expectedVestingLength:     time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Unix() - time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix(),
// 				multiplier:                types.Small,
// 			},
// 			errArgs{
// 				expectPass: true,
// 				contains:   "",
// 			},
// 		},
// 		{
// 			"valid large claim",
// 			args{
// 				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
// 				collateralType:            "bnb-a",
// 				id:                        1,
// 				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
// 				rewards:                   types.Rewards{types.NewReward(true, "bnb-a", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)},
// 				rewardperiods:             types.RewardPeriods{types.NewRewardPeriod("bnb-a", time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), c("ukava", 1000), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24*2), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claimPeriods:              types.ClaimPeriods{types.NewClaimPeriod("bnb-a", 1, time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claims:                    types.Claims{types.NewClaim(sdk.AccAddress(crypto.AddressHash([]byte("test"))), sdk.NewCoin("ukava", sdk.NewInt(1000)), "bnb-a", 1)},
// 				genIDs:                    types.GenesisClaimPeriodIDs{types.GenesisClaimPeriodID{CollateralType: "bnb-a", ID: 2}},
// 				active:                    true,
// 				validatorVesting:          false,
// 				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
// 				expectedModAccountBalance: sdk.Coins(nil),
// 				expectedVestingAccount:    true,
// 				expectedVestingLength:     time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).AddDate(0, 12, 0).Unix() - time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Unix(),
// 				multiplier:                types.Large,
// 			},
// 			errArgs{
// 				expectPass: true,
// 				contains:   "",
// 			},
// 		},
// 		{
// 			"valid liquid claim",
// 			args{
// 				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
// 				collateralType:            "bnb-a",
// 				id:                        1,
// 				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
// 				rewards:                   types.Rewards{types.NewReward(true, "bnb-a", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)},
// 				rewardperiods:             types.RewardPeriods{types.NewRewardPeriod("bnb-a", time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), c("ukava", 1000), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24*2), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claimPeriods:              types.ClaimPeriods{types.NewClaimPeriod("bnb-a", 1, time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claims:                    types.Claims{types.NewClaim(sdk.AccAddress(crypto.AddressHash([]byte("test"))), sdk.NewCoin("ukava", sdk.NewInt(1000)), "bnb-a", 1)},
// 				genIDs:                    types.GenesisClaimPeriodIDs{types.GenesisClaimPeriodID{CollateralType: "bnb-a", ID: 2}},
// 				active:                    true,
// 				validatorVesting:          false,
// 				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
// 				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500))),
// 				expectedVestingAccount:    false,
// 				expectedVestingLength:     0,
// 				multiplier:                types.Small,
// 			},
// 			errArgs{
// 				expectPass: true,
// 				contains:   "",
// 			},
// 		},
// 		{
// 			"no matching claim",
// 			args{
// 				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
// 				collateralType:            "btcb-a",
// 				id:                        1,
// 				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
// 				rewards:                   types.Rewards{types.NewReward(true, "bnb-a", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)},
// 				rewardperiods:             types.RewardPeriods{types.NewRewardPeriod("bnb-a", time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), c("ukava", 1000), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24*2), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claimPeriods:              types.ClaimPeriods{types.NewClaimPeriod("bnb-a", 1, time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), types.Multipliers{types.NewMultiplier(types.Small, 0, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claims:                    types.Claims{types.NewClaim(sdk.AccAddress(crypto.AddressHash([]byte("test"))), sdk.NewCoin("ukava", sdk.NewInt(1000)), "bnb-a", 1)},
// 				genIDs:                    types.GenesisClaimPeriodIDs{types.GenesisClaimPeriodID{CollateralType: "bnb-a", ID: 2}},
// 				active:                    true,
// 				validatorVesting:          false,
// 				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
// 				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500))),
// 				expectedVestingAccount:    false,
// 				expectedVestingLength:     0,
// 				multiplier:                types.Small,
// 			},
// 			errArgs{
// 				expectPass: false,
// 				contains:   "no claim with input id found for owner and collateral type",
// 			},
// 		},
// 		{
// 			"validator vesting claim",
// 			args{
// 				claimOwner:                sdk.AccAddress(crypto.AddressHash([]byte("test"))),
// 				collateralType:            "bnb-a",
// 				id:                        1,
// 				blockTime:                 time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC),
// 				rewards:                   types.Rewards{types.NewReward(true, "bnb-a", c("ukava", 1000000000), time.Hour*7*24, types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))}, time.Hour*7*24)},
// 				rewardperiods:             types.RewardPeriods{types.NewRewardPeriod("bnb-a", time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), c("ukava", 1000), time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24*2), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claimPeriods:              types.ClaimPeriods{types.NewClaimPeriod("bnb-a", 1, time.Date(2020, 11, 1, 14, 0, 0, 0, time.UTC).Add(time.Hour*7*24), types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.5")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))})},
// 				claims:                    types.Claims{types.NewClaim(sdk.AccAddress(crypto.AddressHash([]byte("test"))), sdk.NewCoin("ukava", sdk.NewInt(1000)), "bnb-a", 1)},
// 				genIDs:                    types.GenesisClaimPeriodIDs{types.GenesisClaimPeriodID{CollateralType: "bnb-a", ID: 2}},
// 				active:                    true,
// 				validatorVesting:          true,
// 				expectedAccountBalance:    sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500)), sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
// 				expectedModAccountBalance: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500))),
// 				expectedVestingAccount:    false,
// 				expectedVestingLength:     0,
// 				multiplier:                types.Small,
// 			},
// 			errArgs{
// 				expectPass: false,
// 				contains:   "account type not supported",
// 			},
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(tc.name, func() {
// 			// create new app with one funded account
// 			config := sdk.GetConfig()
// 			app.SetBech32AddressPrefixes(config)
// 			// Initialize test app and set context
// 			tApp := app.NewTestApp()
// 			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tc.args.blockTime})
// 			authGS := app.NewAuthGenState(
// 				[]sdk.AccAddress{tc.args.claimOwner},
// 				[]sdk.Coins{
// 					sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000)), sdk.NewCoin("btcb", sdk.NewInt(1000))),
// 				})
// 			incentiveGS := types.NewGenesisState(types.NewParams(tc.args.active, tc.args.rewards), types.DefaultPreviousBlockTime, tc.args.rewardperiods, tc.args.claimPeriods, tc.args.claims, tc.args.genIDs)
// 			tApp.InitializeFromGenesisStates(authGS, app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(incentiveGS)})
// 			if tc.args.validatorVesting {
// 				ak := tApp.GetAccountKeeper()
// 				acc := ak.GetAccount(ctx, tc.args.claimOwner)
// 				bacc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
// 				bva, err := vesting.NewBaseVestingAccount(
// 					bacc,
// 					sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(20))), time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC).Unix()+100)
// 				suite.Require().NoError(err)
// 				vva := validatorvesting.NewValidatorVestingAccountRaw(
// 					bva,
// 					time.Date(2020, 10, 8, 14, 0, 0, 0, time.UTC).Unix(),
// 					vesting.Periods{
// 						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))},
// 						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))},
// 						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))},
// 						vesting.Period{Length: 25, Amount: cs(c("bnb", 5))}},
// 					sdk.ConsAddress(crypto.AddressHash([]byte("test"))),
// 					sdk.AccAddress{},
// 					95,
// 				)
// 				err = vva.Validate()
// 				suite.Require().NoError(err)
// 				ak.SetAccount(ctx, vva)
// 			}
// 			supplyKeeper := tApp.GetSupplyKeeper()
// 			supplyKeeper.MintCoins(ctx, types.IncentiveMacc, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000))))
// 			keeper := tApp.GetIncentiveKeeper()
// 			suite.app = tApp
// 			suite.ctx = ctx
// 			suite.keeper = keeper

// 			err := suite.keeper.PayoutClaim(suite.ctx, tc.args.claimOwner, tc.args.collateralType, tc.args.id, tc.args.multiplier)

// 			if tc.errArgs.expectPass {
// 				suite.Require().NoError(err)
// 				acc := suite.getAccount(tc.args.claimOwner)
// 				suite.Require().Equal(tc.args.expectedAccountBalance, acc.GetCoins())
// 				mAcc := suite.getModuleAccount(types.IncentiveMacc)
// 				suite.Require().Equal(tc.args.expectedModAccountBalance, mAcc.GetCoins())
// 				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
// 				if tc.args.expectedVestingAccount {
// 					suite.Require().True(ok)
// 					suite.Require().Equal(tc.args.expectedVestingLength, vacc.VestingPeriods[0].Length)
// 				} else {
// 					suite.Require().False(ok)
// 				}
// 				_, f := suite.keeper.GetClaim(ctx, tc.args.claimOwner, tc.args.collateralType, tc.args.id)
// 				suite.Require().False(f)
// 			} else {
// 				suite.Require().Error(err)
// 				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
// 			}
// 		})
// 	}
// }
