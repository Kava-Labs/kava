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
	_, addrs := app.GeneratePrivKeyAddressPairs(10)
	authGS := app.NewAuthGenState(
		addrs[0:2],
		[]sdk.Coins{
			cs(c("xrp", 500000000), c("btc", 500000000)),
			cs(c("xrp", 200000000))})
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
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 400000000)), cs(c("usdx", 10000000)))
	suite.NoError(err)
}

func (suite *DepositTestSuite) TestGetSetDeposit() {
	d, found := suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	suite.True(found)
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 400000000)))
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
	err := suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10000000)))
	suite.NoError(err)
	d, found := suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	suite.True(found)
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 410000000)))
	suite.True(d.Equals(td))
	ds := suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(1, len(ds))
	suite.True(ds[0].Equals(td))
	cd, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.Equal(cs(c("xrp", 410000000)), cd.Collateral)
	ak := suite.app.GetAccountKeeper()
	acc := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(i(90000000), acc.GetCoins().AmountOf("xrp"))

	err = suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("btc", 1)))
	suite.Equal(types.CodeCdpNotFound, err.Result().Code)

	err = suite.keeper.DepositCollateral(suite.ctx, suite.addrs[1], suite.addrs[0], cs(c("xrp", 1)))
	suite.Equal(types.CodeCdpNotFound, err.Result().Code)

	err = suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[1], cs(c("xrp", 10000000)))
	suite.NoError(err)
	d, found = suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[1])
	suite.True(found)
	td = types.NewDeposit(uint64(1), suite.addrs[1], cs(c("xrp", 10000000)))
	suite.True(d.Equals(td))
	ds = suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(2, len(ds))
	suite.True(ds[1].Equals(td))
}

func (suite *DepositTestSuite) TestWithdrawCollateral() {
	err := suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 321000000)))
	suite.Equal(types.CodeInvalidCollateralRatio, err.Result().Code)
	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[1], suite.addrs[0], cs(c("xrp", 10000000)))
	suite.Equal(types.CodeCdpNotFound, err.Result().Code)

	cd, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	cd.AccumulatedFees = cs(c("usdx", 1))
	suite.keeper.SetCDP(suite.ctx, cd)
	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 320000000)))
	suite.Equal(types.CodeInvalidCollateralRatio, err.Result().Code)

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10000000)))
	suite.NoError(err)
	dep, _ := suite.keeper.GetDeposit(suite.ctx, uint64(1), suite.addrs[0])
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 390000000)))
	suite.True(dep.Equals(td))
	ak := suite.app.GetAccountKeeper()
	acc := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(i(110000000), acc.GetCoins().AmountOf("xrp"))

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[1], cs(c("xrp", 10000000)))
	suite.Equal(types.CodeDepositNotFound, err.Result().Code)
}

func TestDepositTestSuite(t *testing.T) {
	suite.Run(t, new(DepositTestSuite))
}
