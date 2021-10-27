package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper        keeper.Keeper
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	app           app.TestApp
	ctx           sdk.Context
}

var (
	testPeriods = []types.Period{
		{
			Start:     time.Date(2020, time.March, 1, 1, 0, 0, 0, time.UTC),
			End:       time.Date(2021, time.March, 1, 1, 0, 0, 0, time.UTC),
			Inflation: sdk.MustNewDecFromStr("1.000000003022265980"),
		},
	}
)

func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	coins := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000000000)))
	authGS := app.NewFundedGenStateWithSameCoins(tApp.AppCodec(), coins, addrs)

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	params := types.NewParams(true, testPeriods)
	gs := app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(types.NewGenesisState(params, types.DefaultPreviousBlockTime))}
	tApp.InitializeFromGenesisStates(
		authGS,
		gs,
	)
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = tApp.GetKavadistKeeper()
	suite.bankKeeper = tApp.GetBankKeeper()
	suite.accountKeeper = tApp.GetAccountKeeper()
}

func (suite *KeeperTestSuite) TestMintExpiredPeriod() {
	initialSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.NotPanics(func() { suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)) })
	ctx := suite.ctx.WithBlockTime(time.Date(2022, 1, 1, 0, 7, 0, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.Equal(initialSupply, finalSupply)
}

func (suite *KeeperTestSuite) TestMintPeriodNotStarted() {
	initialSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.NotPanics(func() { suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)) })
	ctx := suite.ctx.WithBlockTime(time.Date(2019, 1, 1, 0, 7, 0, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.Equal(initialSupply, finalSupply)
}

func (suite *KeeperTestSuite) TestMintOngoingPeriod() {
	initialSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.NotPanics(func() {
		suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2020, time.March, 1, 1, 0, 1, 0, time.UTC))
	})
	ctx := suite.ctx.WithBlockTime(time.Date(2021, 2, 28, 23, 59, 59, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.True(finalSupply.Amount.GT(initialSupply.Amount))
	mAcc := suite.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	mAccSupply := suite.bankKeeper.GetAllBalances(ctx, mAcc.GetAddress()).AmountOf(types.GovDenom)
	suite.True(mAccSupply.Equal(finalSupply.Amount.Sub(initialSupply.Amount)))
	// expect that inflation is ~10%
	expectedSupply := sdk.NewDecFromInt(initialSupply.Amount).Mul(sdk.MustNewDecFromStr("1.1"))
	supplyError := sdk.OneDec().Sub((sdk.NewDecFromInt(finalSupply.Amount).Quo(expectedSupply))).Abs()
	suite.True(supplyError.LTE(sdk.MustNewDecFromStr("0.001")))
}

func (suite *KeeperTestSuite) TestMintPeriodTransition() {
	initialSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	params := suite.keeper.GetParams(suite.ctx)
	periods := []types.Period{
		testPeriods[0],
		{
			Start:     time.Date(2021, time.March, 1, 1, 0, 0, 0, time.UTC),
			End:       time.Date(2022, time.March, 1, 1, 0, 0, 0, time.UTC),
			Inflation: sdk.MustNewDecFromStr("1.000000003022265980"),
		},
	}
	params.Periods = periods
	suite.NotPanics(func() {
		suite.keeper.SetParams(suite.ctx, params)
	})
	suite.NotPanics(func() {
		suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2020, time.March, 1, 1, 0, 1, 0, time.UTC))
	})
	ctx := suite.ctx.WithBlockTime(time.Date(2021, 3, 10, 0, 0, 0, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.True(finalSupply.Amount.GT(initialSupply.Amount))
}

func (suite *KeeperTestSuite) TestMintNotActive() {
	initialSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	params := suite.keeper.GetParams(suite.ctx)
	params.Active = false
	suite.NotPanics(func() {
		suite.keeper.SetParams(suite.ctx, params)
	})
	suite.NotPanics(func() {
		suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2020, time.March, 1, 1, 0, 1, 0, time.UTC))
	})
	ctx := suite.ctx.WithBlockTime(time.Date(2021, 2, 28, 23, 59, 59, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.bankKeeper.GetSupply(suite.ctx, types.GovDenom)
	suite.Equal(initialSupply, finalSupply)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
