package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type SavingsTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

func (suite *SavingsTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("usdx", 100000)), cs(c("usdx", 50000)), cs(c("usdx", 50000)),
		},
	)
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	sk := tApp.GetSupplyKeeper()
	macc := sk.GetModuleAccount(ctx, types.LiquidatorMacc)
	err := sk.MintCoins(ctx, macc.GetName(), cs(c("usdx", 10000)))
	suite.NoError(err)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.keeper = keeper
	suite.ctx = ctx
	suite.addrs = addrs
}

func (suite *SavingsTestSuite) TestApplySavingsRate() {
	err := suite.keeper.ApplySavingsRate(suite.ctx, "usdx")
	suite.NoError(err)
	ak := suite.app.GetAccountKeeper()
	acc0 := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(cs(c("usdx", 104750)), acc0.GetCoins())
	acc1 := ak.GetAccount(suite.ctx, suite.addrs[1])
	suite.Equal(cs(c("usdx", 52375)), acc1.GetCoins())
	acc2 := ak.GetAccount(suite.ctx, suite.addrs[2])
	suite.Equal(cs(c("usdx", 52375)), acc2.GetCoins())
}

func (suite *SavingsTestSuite) TestGetSetPreviousDistributionTime() {
	now := tmtime.Now()

	_, f := suite.keeper.GetPreviousSavingsDistribution(suite.ctx)
	suite.False(f)

	suite.NotPanics(func() { suite.keeper.SetPreviousSavingsDistribution(suite.ctx, now) })

	pdt, f := suite.keeper.GetPreviousSavingsDistribution(suite.ctx)
	suite.True(f)
	suite.Equal(now, pdt)

}

func TestSavingsTestSuite(t *testing.T) {
	suite.Run(t, new(SavingsTestSuite))
}
