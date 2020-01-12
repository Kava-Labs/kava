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
	d, found := suite.keeper.GetDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
	suite.True(found)
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 400000000)))
	suite.True(d.Equals(td))
	ds := suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(1, len(ds))
	suite.True(ds[0].Equals(td))
	suite.keeper.DeleteDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
	_, found = suite.keeper.GetDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
	suite.False(found)
	ds = suite.keeper.GetDeposits(suite.ctx, uint64(1))
	suite.Equal(0, len(ds))
}

func (suite *DepositTestSuite) TestDepositCollateral() {
	err := suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10000000)))
	suite.NoError(err)
	d, found := suite.keeper.GetDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
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
	d, found = suite.keeper.GetDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[1])
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

	d, _ := suite.keeper.GetDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
	d.InLiquidation = true
	suite.keeper.DeleteDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
	suite.keeper.SetDeposit(suite.ctx, d)
	_, f := suite.keeper.GetDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
	suite.False(f)

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10000000)))
	suite.Equal(types.CodeCdpNotAvailable, err.Result().Code)

	d, f = suite.keeper.GetDeposit(suite.ctx, types.StatusLiquidated, uint64(1), suite.addrs[0])
	suite.True(f)
	suite.keeper.DeleteDeposit(suite.ctx, types.StatusLiquidated, uint64(1), suite.addrs[0])
	d.InLiquidation = false
	suite.keeper.SetDeposit(suite.ctx, d)
	_, f = suite.keeper.GetDeposit(suite.ctx, types.StatusLiquidated, uint64(1), suite.addrs[0])
	suite.False(f)

	cd, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	cd.AccumulatedFees = cs(c("usdx", 1))
	suite.keeper.SetCDP(suite.ctx, cd)
	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 320000000)))
	suite.Equal(types.CodeInvalidCollateralRatio, err.Result().Code)

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10000000)))
	suite.NoError(err)
	d, _ = suite.keeper.GetDeposit(suite.ctx, types.StatusNil, uint64(1), suite.addrs[0])
	td := types.NewDeposit(uint64(1), suite.addrs[0], cs(c("xrp", 390000000)))
	suite.True(d.Equals(td))
	ak := suite.app.GetAccountKeeper()
	acc := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(i(110000000), acc.GetCoins().AmountOf("xrp"))

	err = suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[1], cs(c("xrp", 10000000)))
	suite.Equal(types.CodeDepositNotFound, err.Result().Code)
}

func (suite *DepositTestSuite) TestIterateLiquidatedDeposits() {
	for j := 0; j < 10; j++ {
		d := types.NewDeposit(uint64(j+2), suite.addrs[j], cs(c("xrp", 1000000)))
		if j%2 == 0 {
			d.InLiquidation = true
		}
		suite.keeper.SetDeposit(suite.ctx, d)
	}
	ds := suite.keeper.GetAllLiquidatedDeposits(suite.ctx)
	for _, d := range ds {
		suite.True(d.InLiquidation)
	}
	suite.Equal(5, len(ds))
}
func TestDepositTestSuite(t *testing.T) {
	suite.Run(t, new(DepositTestSuite))
}
