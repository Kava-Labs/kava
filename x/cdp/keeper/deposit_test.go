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

type DepositTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

func (suite *DepositTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("xrp", 500), c("btc", 5)),
			cs(c("xrp", 200))})
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.keeper = keeper
	suite.ctx = ctx
	suite.addrs = addrs
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 400)), cs(c("usdx", 10)))
	suite.NoError(err)
}

func (suite *DepositTestSuite) TestGetSetDeposit() {
	d, found := suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	suite.True(found)
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 400)))
	suite.True(d.Equals(td))
	ds := suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(1, len(ds))
	suite.True(ds[0].Equals(td))
	suite.keeper.DeleteDeposit(suite.ctx, uint64(1), suite.addrs[0])
	_, found = suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	suite.False(found)
	ds = suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(0, len(ds))
}

func (suite *DepositTestSuite) TestDepositCollateral() {
	err := suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10)))
	suite.NoError(err)
	d, found := suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	suite.True(found)
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 410)))
	suite.True(d.Equals(td))
	ds := suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(1, len(ds))
	suite.True(ds[0].Equals(td))
	cd, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.Equal(cs(c("xrp", 410)), cd.Collateral)

	err = suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("btc", 1)))
	suite.Equal(sdk.CodeType(7), err.Result().Code)

	err = suite.keeper.DepositCollateral(suite.ctx, suite.addrs[1], suite.addrs[0], cs(c("xrp", 1)))
	suite.Equal(sdk.CodeType(7), err.Result().Code)

	err = suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[1], cs(c("xrp", 10)))
	suite.NoError(err)
	d, found = suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[1])
	suite.True(found)
	td = types.NewDeposit(uint64(1), suite.addrs[1], cs(c("xrp", 10)))
	suite.True(d.Equals(td))
	ds = suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(2, len(ds))
	suite.True(ds[1].Equals(td))
}

func (suite *DepositTestSuite) TestWithdrawCollateral() {
	err := suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 321)))
	suite.Equal(sdk.CodeType(6), err.Result().Code)
	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[1], suite.addrs[0], cs(c("xrp", 10)))
	suite.Equal(sdk.CodeType(7), err.Result().Code)

	d, _ := suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	d.InLiquidation = true
	suite.keeper.SetDeposit(suite.ctx, d, uint64(1))

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10)))
	suite.Equal(sdk.CodeType(11), err.Result().Code)

	d, _ = suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	d.InLiquidation = false
	suite.keeper.SetDeposit(suite.ctx, d, uint64(1))
	cd, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	cd.AccumulatedFees = cs(c("usdx", 1))
	suite.keeper.SetCDP(suite.ctx, cd)
	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 320)))
	suite.Equal(sdk.CodeType(6), err.Result().Code)

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10)))
	suite.NoError(err)
	d, _ = suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 390)))
	suite.True(d.Equals(td))

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[1], cs(c("xrp", 10)))
	suite.Equal(sdk.CodeType(8), err.Result().Code)
}

func TestDepositTestSuite(t *testing.T) {
	suite.Run(t, new(DepositTestSuite))
}
