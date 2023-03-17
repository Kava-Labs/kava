package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *keeperTestSuite) TestMintExpiredPeriod() {
	initialSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().NotPanics(func() { suite.Keeper.SetPreviousBlockTime(suite.Ctx, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)) })
	ctx := suite.Ctx.WithBlockTime(time.Date(2022, 1, 1, 0, 7, 0, 0, time.UTC))
	err := suite.Keeper.MintPeriodInflation(ctx)
	suite.Require().NoError(err)
	finalSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().Equal(initialSupply, finalSupply)
}

func (suite *keeperTestSuite) TestMintPeriodNotStarted() {
	initialSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().NotPanics(func() { suite.Keeper.SetPreviousBlockTime(suite.Ctx, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)) })
	ctx := suite.Ctx.WithBlockTime(time.Date(2019, 1, 1, 0, 7, 0, 0, time.UTC))
	err := suite.Keeper.MintPeriodInflation(ctx)
	suite.Require().NoError(err)
	finalSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().Equal(initialSupply, finalSupply)
}

func (suite *keeperTestSuite) TestMintOngoingPeriod() {
	initialSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().NotPanics(func() {
		suite.Keeper.SetPreviousBlockTime(suite.Ctx, time.Date(2020, time.March, 1, 1, 0, 1, 0, time.UTC))
	})
	ctx := suite.Ctx.WithBlockTime(time.Date(2021, 2, 28, 23, 59, 59, 0, time.UTC))
	err := suite.Keeper.MintPeriodInflation(ctx)
	suite.Require().NoError(err)
	finalSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().True(finalSupply.Amount.GT(initialSupply.Amount))
	mAcc := suite.AccountKeeper.GetModuleAccount(ctx, types.ModuleName)
	mAccSupply := suite.BankKeeper.GetAllBalances(ctx, mAcc.GetAddress()).AmountOf(types.GovDenom)
	suite.Require().True(mAccSupply.Equal(finalSupply.Amount.Sub(initialSupply.Amount)))
	// expect that inflation is ~10%
	expectedSupply := sdk.NewDecFromInt(initialSupply.Amount).Mul(sdk.MustNewDecFromStr("1.1"))
	supplyError := sdk.OneDec().Sub((sdk.NewDecFromInt(finalSupply.Amount).Quo(expectedSupply))).Abs()
	suite.Require().True(supplyError.LTE(sdk.MustNewDecFromStr("0.001")))
}

func (suite *keeperTestSuite) TestMintPeriodTransition() {
	initialSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	params := suite.Keeper.GetParams(suite.Ctx)
	periods := []types.Period{
		suite.TestPeriods[0],
		{
			Start:     time.Date(2021, time.March, 1, 1, 0, 0, 0, time.UTC),
			End:       time.Date(2022, time.March, 1, 1, 0, 0, 0, time.UTC),
			Inflation: sdk.MustNewDecFromStr("1.000000003022265980"),
		},
	}
	params.Periods = periods
	suite.Require().NotPanics(func() {
		suite.Keeper.SetParams(suite.Ctx, params)
	})
	suite.Require().NotPanics(func() {
		suite.Keeper.SetPreviousBlockTime(suite.Ctx, time.Date(2020, time.March, 1, 1, 0, 1, 0, time.UTC))
	})
	ctx := suite.Ctx.WithBlockTime(time.Date(2021, 3, 10, 0, 0, 0, 0, time.UTC))
	err := suite.Keeper.MintPeriodInflation(ctx)
	suite.Require().NoError(err)
	finalSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().True(finalSupply.Amount.GT(initialSupply.Amount))
}

func (suite *keeperTestSuite) TestMintNotActive() {
	initialSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	params := suite.Keeper.GetParams(suite.Ctx)
	params.Active = false
	suite.Require().NotPanics(func() {
		suite.Keeper.SetParams(suite.Ctx, params)
	})
	suite.Require().NotPanics(func() {
		suite.Keeper.SetPreviousBlockTime(suite.Ctx, time.Date(2020, time.March, 1, 1, 0, 1, 0, time.UTC))
	})
	ctx := suite.Ctx.WithBlockTime(time.Date(2021, 2, 28, 23, 59, 59, 0, time.UTC))
	err := suite.Keeper.MintPeriodInflation(ctx)
	suite.Require().NoError(err)
	finalSupply := suite.BankKeeper.GetSupply(suite.Ctx, types.GovDenom)
	suite.Require().Equal(initialSupply, finalSupply)
}

func (suite *keeperTestSuite) TestInfraMinting() {
	type args struct {
		startTime           time.Time
		endTime             time.Time
		infraPeriods        types.Periods
		expectedFinalSupply sdk.Coin
		marginOfError       sdk.Dec
	}

	type errArgs struct {
		expectPass bool
		contains   string
	}

	type test struct {
		name    string
		args    args
		errArgs errArgs
	}

	testCases := []test{
		{
			"5% apy one year",
			args{
				startTime:           time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC),
				endTime:             time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC),
				infraPeriods:        types.Periods{types.NewPeriod(time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC), time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC), sdk.MustNewDecFromStr("1.000000001547125958"))},
				expectedFinalSupply: sdk.NewCoin(types.GovDenom, sdk.NewInt(1050000000000)),
				marginOfError:       sdk.MustNewDecFromStr("0.0001"),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"5% apy 10 seconds",
			args{
				startTime:           time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC),
				endTime:             time.Date(2022, time.October, 1, 1, 0, 10, 0, time.UTC),
				infraPeriods:        types.Periods{types.NewPeriod(time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC), time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC), sdk.MustNewDecFromStr("1.000000001547125958"))},
				expectedFinalSupply: sdk.NewCoin(types.GovDenom, sdk.NewInt(1000000015471)),
				marginOfError:       sdk.MustNewDecFromStr("0.0001"),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		params := types.NewParams(true, types.DefaultPeriods, types.NewInfraParams(tc.args.infraPeriods, types.DefaultInfraParams.PartnerRewards, types.DefaultInfraParams.CoreRewards))
		ctx := suite.Ctx.WithBlockTime(tc.args.startTime)
		suite.Keeper.SetParams(ctx, params)
		suite.Require().NotPanics(func() {
			suite.Keeper.SetPreviousBlockTime(ctx, tc.args.startTime)
		})

		// Delete initial genesis tokens to start with a clean slate
		suite.App.DeleteGenesisValidator(suite.T(), suite.Ctx)
		suite.App.DeleteGenesisValidatorCoins(suite.T(), suite.Ctx)

		ctx = suite.Ctx.WithBlockTime(tc.args.endTime)
		err := suite.Keeper.MintPeriodInflation(ctx)
		suite.Require().NoError(err)

		finalSupply := suite.BankKeeper.GetSupply(ctx, types.GovDenom)
		marginHigh := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Add(tc.args.marginOfError))
		marginLow := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Sub(tc.args.marginOfError))
		suite.Require().Truef(
			sdk.NewDecFromInt(finalSupply.Amount).LTE(marginHigh),
			"final supply %s is not <= %s high margin",
			finalSupply.Amount.String(),
			marginHigh.String(),
		)
		suite.Require().Truef(
			sdk.NewDecFromInt(finalSupply.Amount).GTE(marginLow),
			"final supply %s is not >= %s low margin",
			finalSupply.Amount.String(),
		)

	}

}

func (suite *keeperTestSuite) TestInfraPayoutCore() {

	type args struct {
		startTime               time.Time
		endTime                 time.Time
		infraPeriods            types.Periods
		expectedFinalSupply     sdk.Coin
		expectedBalanceIncrease sdk.Coin
		marginOfError           sdk.Dec
	}

	type errArgs struct {
		expectPass bool
		contains   string
	}

	type test struct {
		name    string
		args    args
		errArgs errArgs
	}

	testCases := []test{
		{
			"5% apy one year",
			args{
				startTime:               time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC),
				endTime:                 time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC),
				infraPeriods:            types.Periods{types.NewPeriod(time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC), time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC), sdk.MustNewDecFromStr("1.000000001547125958"))},
				expectedFinalSupply:     sdk.NewCoin(types.GovDenom, sdk.NewInt(1050000000000)),
				expectedBalanceIncrease: sdk.NewCoin(types.GovDenom, sdk.NewInt(50000000000)),
				marginOfError:           sdk.MustNewDecFromStr("0.0001"),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		coreReward := types.NewCoreReward(suite.Addrs[0], sdk.OneDec())
		params := types.NewParams(true, types.DefaultPeriods, types.NewInfraParams(tc.args.infraPeriods, types.DefaultInfraParams.PartnerRewards, types.CoreRewards{coreReward}))
		ctx := suite.Ctx.WithBlockTime(tc.args.startTime)
		suite.Keeper.SetParams(ctx, params)
		suite.Require().NotPanics(func() {
			suite.Keeper.SetPreviousBlockTime(ctx, tc.args.startTime)
		})

		// Delete initial genesis tokens to start with a clean slate
		suite.App.DeleteGenesisValidator(suite.T(), suite.Ctx)
		suite.App.DeleteGenesisValidatorCoins(suite.T(), suite.Ctx)

		initialBalance := suite.BankKeeper.GetBalance(ctx, suite.Addrs[0], types.GovDenom)
		ctx = suite.Ctx.WithBlockTime(tc.args.endTime)
		err := suite.Keeper.MintPeriodInflation(ctx)
		suite.Require().NoError(err)
		finalSupply := suite.BankKeeper.GetSupply(ctx, types.GovDenom)
		marginHigh := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Add(tc.args.marginOfError))
		marginLow := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Sub(tc.args.marginOfError))
		suite.Require().True(sdk.NewDecFromInt(finalSupply.Amount).LTE(marginHigh))
		suite.Require().True(sdk.NewDecFromInt(finalSupply.Amount).GTE(marginLow))

		finalBalance := suite.BankKeeper.GetBalance(ctx, suite.Addrs[0], types.GovDenom)
		suite.Require().Equal(tc.args.expectedBalanceIncrease, finalBalance.Sub(initialBalance))

	}

}

func (suite *keeperTestSuite) TestInfraPayoutPartner() {

	type args struct {
		startTime               time.Time
		endTime                 time.Time
		infraPeriods            types.Periods
		expectedFinalSupply     sdk.Coin
		expectedBalanceIncrease sdk.Coin
		marginOfError           sdk.Dec
	}

	type errArgs struct {
		expectPass bool
		contains   string
	}

	type test struct {
		name    string
		args    args
		errArgs errArgs
	}

	testCases := []test{
		{
			"5% apy one year",
			args{
				startTime:               time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC),
				endTime:                 time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC),
				infraPeriods:            types.Periods{types.NewPeriod(time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC), time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC), sdk.MustNewDecFromStr("1.000000001547125958"))},
				expectedFinalSupply:     sdk.NewCoin(types.GovDenom, sdk.NewInt(1050000000000)),
				expectedBalanceIncrease: sdk.NewCoin(types.GovDenom, sdk.NewInt(63072000)),
				marginOfError:           sdk.MustNewDecFromStr("0.0001"),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		partnerReward := types.NewPartnerReward(suite.Addrs[0], sdk.NewCoin(types.GovDenom, sdk.NewInt(2)))
		params := types.NewParams(true, types.DefaultPeriods, types.NewInfraParams(tc.args.infraPeriods, types.PartnerRewards{partnerReward}, types.DefaultInfraParams.CoreRewards))
		ctx := suite.Ctx.WithBlockTime(tc.args.startTime)
		suite.Keeper.SetParams(ctx, params)
		suite.Require().NotPanics(func() {
			suite.Keeper.SetPreviousBlockTime(ctx, tc.args.startTime)
		})

		// Delete initial genesis tokens to start with a clean slate
		suite.App.DeleteGenesisValidator(suite.T(), suite.Ctx)
		suite.App.DeleteGenesisValidatorCoins(suite.T(), suite.Ctx)

		initialBalance := suite.BankKeeper.GetBalance(ctx, suite.Addrs[0], types.GovDenom)
		ctx = suite.Ctx.WithBlockTime(tc.args.endTime)
		err := suite.Keeper.MintPeriodInflation(ctx)
		suite.Require().NoError(err)
		finalSupply := suite.BankKeeper.GetSupply(ctx, types.GovDenom)
		marginHigh := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Add(tc.args.marginOfError))
		marginLow := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Sub(tc.args.marginOfError))
		suite.Require().True(sdk.NewDecFromInt(finalSupply.Amount).LTE(marginHigh))
		suite.Require().True(sdk.NewDecFromInt(finalSupply.Amount).GTE(marginLow))

		finalBalance := suite.BankKeeper.GetBalance(ctx, suite.Addrs[0], types.GovDenom)
		suite.Require().Equal(tc.args.expectedBalanceIncrease, finalBalance.Sub(initialBalance))

	}

}

func (suite *keeperTestSuite) TestInfraPayoutE2E() {

	type balance struct {
		address sdk.AccAddress
		amount  sdk.Coins
	}

	type balances []balance

	type args struct {
		periods             types.Periods
		startTime           time.Time
		endTime             time.Time
		infraPeriods        types.Periods
		coreRewards         types.CoreRewards
		partnerRewards      types.PartnerRewards
		expectedFinalSupply sdk.Coin
		expectedBalances    balances
		marginOfError       sdk.Dec
	}

	type errArgs struct {
		expectPass bool
		contains   string
	}

	type test struct {
		name    string
		args    args
		errArgs errArgs
	}

	_, addrs := app.GeneratePrivKeyAddressPairs(3)

	testCases := []test{
		{
			"5% apy one year",
			args{
				periods:             types.Periods{types.NewPeriod(time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC), time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC), sdk.MustNewDecFromStr("1.000000001547125958"))},
				startTime:           time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC),
				endTime:             time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC),
				infraPeriods:        types.Periods{types.NewPeriod(time.Date(2022, time.October, 1, 1, 0, 0, 0, time.UTC), time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC), sdk.MustNewDecFromStr("1.000000001547125958"))},
				coreRewards:         types.CoreRewards{types.NewCoreReward(addrs[1], sdk.OneDec())},
				partnerRewards:      types.PartnerRewards{types.NewPartnerReward(addrs[2], sdk.NewCoin("ukava", sdk.NewInt(2)))},
				expectedFinalSupply: sdk.NewCoin(types.GovDenom, sdk.NewInt(1102500000000)),
				expectedBalances: balances{
					balance{addrs[1], sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(52436928000)))},
					balance{addrs[2], sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(63072000)))},
				},
				marginOfError: sdk.MustNewDecFromStr("0.0001"),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		params := types.NewParams(true, tc.args.periods, types.NewInfraParams(tc.args.infraPeriods, tc.args.partnerRewards, tc.args.coreRewards))
		ctx := suite.Ctx.WithBlockTime(tc.args.startTime)
		suite.Keeper.SetParams(ctx, params)
		suite.Require().NotPanics(func() {
			suite.Keeper.SetPreviousBlockTime(ctx, tc.args.startTime)
		})

		// Delete initial genesis tokens to start with a clean slate
		suite.App.DeleteGenesisValidator(suite.T(), suite.Ctx)
		suite.App.DeleteGenesisValidatorCoins(suite.T(), suite.Ctx)

		ctx = suite.Ctx.WithBlockTime(tc.args.endTime)
		err := suite.Keeper.MintPeriodInflation(ctx)
		suite.Require().NoError(err)
		finalSupply := suite.BankKeeper.GetSupply(ctx, types.GovDenom)
		marginHigh := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Add(tc.args.marginOfError))
		marginLow := sdk.NewDecFromInt(tc.args.expectedFinalSupply.Amount).Mul(sdk.OneDec().Sub(tc.args.marginOfError))
		suite.Require().True(sdk.NewDecFromInt(finalSupply.Amount).LTE(marginHigh))
		suite.Require().True(sdk.NewDecFromInt(finalSupply.Amount).GTE(marginLow))

		for _, bal := range tc.args.expectedBalances {
			finalBalance := suite.BankKeeper.GetAllBalances(ctx, bal.address)
			suite.Require().Equal(bal.amount, finalBalance)
		}
	}
}
