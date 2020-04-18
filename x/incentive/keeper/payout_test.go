package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
	abci "github.com/tendermint/tendermint/abci/types"
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
	err := supplyKeeper.MintCoins(ctx, macc.GetName(), cs(c("ukava", 500)))
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

func (suite *KeeperTestSuite) TestSendCoinsToPeriodicVestingAccount() {
	suite.setupChain()

	type args struct {
		coins  sdk.Coins
		length int64
	}

	type errArgs struct {
		expectErr bool
		code      sdk.CodeType
	}

	type vestingAccountTest struct {
		name                    string
		blockTime               time.Time
		args                    args
		errArgs                 errArgs
		expectedPeriods         vesting.Periods
		expectedOriginalVesting sdk.Coins
		expectedCoins           sdk.Coins
		expectedStartTime       int64
		expectedEndTime         int64
	}

	type vestingAccountTests []vestingAccountTest

	testCases := vestingAccountTests{
		vestingAccountTest{
			name:      "insert period into an existing vesting schedule",
			blockTime: time.Unix(100, 0),
			args:      args{coins: cs(c("ukava", 100)), length: 5},
			errArgs:   errArgs{expectErr: false, code: sdk.CodeType(0)},
			expectedPeriods: vesting.Periods{
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
			},
			expectedOriginalVesting: cs(c("ukava", 500)),
			expectedCoins:           cs(c("ukava", 500)),
			expectedStartTime:       int64(100),
			expectedEndTime:         int64(116),
		},
		vestingAccountTest{
			name:      "append period to the end of an existing vesting schedule",
			blockTime: time.Unix(100, 0),
			args:      args{coins: cs(c("ukava", 100)), length: 17},
			errArgs:   errArgs{expectErr: false, code: sdk.CodeType(0)},
			expectedPeriods: vesting.Periods{
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
			},
			expectedOriginalVesting: cs(c("ukava", 600)),
			expectedCoins:           cs(c("ukava", 600)),
			expectedStartTime:       int64(100),
			expectedEndTime:         int64(117),
		},
		vestingAccountTest{
			name:      "append period to the end of a completed vesting schedule",
			blockTime: time.Unix(120, 0),
			args:      args{coins: cs(c("ukava", 100)), length: 5},
			errArgs:   errArgs{expectErr: false, code: sdk.CodeType(0)},
			expectedPeriods: vesting.Periods{
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
			},
			expectedOriginalVesting: cs(c("ukava", 700)),
			expectedCoins:           cs(c("ukava", 700)),
			expectedStartTime:       int64(100),
			expectedEndTime:         int64(125),
		},
		vestingAccountTest{
			name:      "prepend period to to an upcoming vesting schedule",
			blockTime: time.Unix(90, 0),
			args:      args{coins: cs(c("ukava", 100)), length: 5},
			errArgs:   errArgs{expectErr: false, code: sdk.CodeType(0)},
			expectedPeriods: vesting.Periods{
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
			},
			expectedOriginalVesting: cs(c("ukava", 800)),
			expectedCoins:           cs(c("ukava", 800)),
			expectedStartTime:       int64(90),
			expectedEndTime:         int64(125),
		},
		vestingAccountTest{
			name:      "add period that coincides with an existing end time",
			blockTime: time.Unix(90, 0),
			args:      args{coins: cs(c("ukava", 100)), length: 11},
			errArgs:   errArgs{expectErr: false, code: sdk.CodeType(0)},
			expectedPeriods: vesting.Periods{
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(6), Amount: cs(c("ukava", 200))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
				vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
			},
			expectedOriginalVesting: cs(c("ukava", 900)),
			expectedCoins:           cs(c("ukava", 900)),
			expectedStartTime:       int64(90),
			expectedEndTime:         int64(125),
		},
		vestingAccountTest{
			name:                    "insufficient module account balance",
			blockTime:               time.Unix(90, 0),
			args:                    args{coins: cs(c("ukava", 1000)), length: 11},
			errArgs:                 errArgs{expectErr: true, code: types.CodeInsufficientBalance},
			expectedPeriods:         vesting.Periods{},
			expectedOriginalVesting: sdk.Coins{},
			expectedCoins:           sdk.Coins{},
			expectedStartTime:       int64(0),
			expectedEndTime:         int64(0),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.ctx = suite.ctx.WithBlockTime(tc.blockTime)
			err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], tc.args.coins, tc.args.length)
			if tc.errArgs.expectErr {
				suite.Equal(tc.errArgs.code, err.Result().Code)
			} else {
				suite.Require().NoError(err)
				acc := suite.getAccount(suite.addrs[0])
				vacc, ok := acc.(*vesting.PeriodicVestingAccount)
				suite.True(ok)
				suite.Equal(tc.expectedPeriods, vacc.VestingPeriods)
				suite.Equal(tc.expectedOriginalVesting, vacc.OriginalVesting)
				suite.Equal(tc.expectedCoins, vacc.Coins)
				suite.Equal(tc.expectedStartTime, vacc.StartTime)
				suite.Equal(tc.expectedEndTime, vacc.EndTime)
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
	suite.Equal(types.CodeInvalidAccountType, err.Result().Code)
	macc := suite.getModuleAccount(cdp.ModuleName)
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, macc.GetAddress(), cs(c("ukava", 100)), 5)
	suite.Equal(types.CodeInvalidAccountType, err.Result().Code)
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
	err := suite.keeper.PayoutClaim(suite.ctx, suite.addrs[0], "bnb", 1)
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
	suite.Equal(types.CodeClaimNotFound, err.Result().Code)
	// addrs[0] has an xrp claim, but there is not corresponding claim period
	err = suite.keeper.PayoutClaim(suite.ctx, suite.addrs[0], "xrp", 1)
	suite.Equal(types.CodeClaimPeriodNotFound, err.Result().Code)
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
