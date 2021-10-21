package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper       keeper.Keeper
	supplyKeeper types.SupplyKeeper
	app          app.TestApp
	ctx          sdk.Context
}

var (
	testPeriods = types.Periods{
		types.Period{
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
	coins := []sdk.Coins{sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000000000)))}
	authGS := app.NewAuthGenState(
		addrs, coins)

	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	params := types.NewParams(true, testPeriods)
	gs := app.GenesisState{types.ModuleName: types.ModuleCdc.MustMarshalJSON(types.NewGenesisState(params, types.DefaultPreviousBlockTime))}
	tApp.InitializeFromGenesisStates(
		authGS,
		gs,
	)
	keeper := tApp.GetKavadistKeeper()
	sk := tApp.GetSupplyKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.supplyKeeper = sk

}

func (suite *KeeperTestSuite) TestMintExpiredPeriod() {
	initialSupply := suite.supplyKeeper.GetSupply(suite.ctx).GetTotal().AmountOf(types.GovDenom)
	suite.NotPanics(func() { suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)) })
	ctx := suite.ctx.WithBlockTime(time.Date(2022, 1, 1, 0, 7, 0, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)
	suite.Equal(initialSupply, finalSupply)
}

func (suite *KeeperTestSuite) TestMintPeriodNotStarted() {
	initialSupply := suite.supplyKeeper.GetSupply(suite.ctx).GetTotal().AmountOf(types.GovDenom)
	suite.NotPanics(func() { suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)) })
	ctx := suite.ctx.WithBlockTime(time.Date(2019, 1, 1, 0, 7, 0, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)
	suite.Equal(initialSupply, finalSupply)
}

func (suite *KeeperTestSuite) TestMintOngoingPeriod() {
	initialSupply := suite.supplyKeeper.GetSupply(suite.ctx).GetTotal().AmountOf(types.GovDenom)
	suite.NotPanics(func() {
		suite.keeper.SetPreviousBlockTime(suite.ctx, time.Date(2020, time.March, 1, 1, 0, 1, 0, time.UTC))
	})
	ctx := suite.ctx.WithBlockTime(time.Date(2021, 2, 28, 23, 59, 59, 0, time.UTC))
	err := suite.keeper.MintPeriodInflation(ctx)
	suite.NoError(err)
	finalSupply := suite.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)
	suite.True(finalSupply.GT(initialSupply))
	mAcc := suite.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
	mAccSupply := mAcc.GetCoins().AmountOf(types.GovDenom)
	suite.True(mAccSupply.Equal(finalSupply.Sub(initialSupply)))
	// expect that inflation is ~10%
	expectedSupply := sdk.NewDecFromInt(initialSupply).Mul(sdk.MustNewDecFromStr("1.1"))
	supplyError := sdk.OneDec().Sub((sdk.NewDecFromInt(finalSupply).Quo(expectedSupply))).Abs()
	suite.True(supplyError.LTE(sdk.MustNewDecFromStr("0.001")))
}

func (suite *KeeperTestSuite) TestMintPeriodTransition() {
	initialSupply := suite.supplyKeeper.GetSupply(suite.ctx).GetTotal().AmountOf(types.GovDenom)
	params := suite.keeper.GetParams(suite.ctx)
	periods := types.Periods{
		testPeriods[0],
		types.Period{
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
	finalSupply := suite.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)
	suite.True(finalSupply.GT(initialSupply))
}

func (suite *KeeperTestSuite) TestMintNotActive() {
	initialSupply := suite.supplyKeeper.GetSupply(suite.ctx).GetTotal().AmountOf(types.GovDenom)
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
	finalSupply := suite.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)
	suite.Equal(initialSupply, finalSupply)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
