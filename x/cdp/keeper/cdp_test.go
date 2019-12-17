package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

var (
	BeforeTestMulti = []string{
		"TestIterateCdpsByDenom",
		"TestIterateCdpsByCollateralRatio",
		"TestIterateCdps",
		"TestValidatePrincipal",
		"TestCalculateCollateralizationRatio",
		"TestAddCdp",
	}
)

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenState(sdk.DefaultBondDenom, d("1.0")),
		NewCDPGenState(sdk.DefaultBondDenom, d("1.5")))
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.keeper = keeper
	suite.ctx = ctx
}

func (suite *KeeperTestSuite) BeforeTest(suiteName, testName string) {
	for _, tn := range BeforeTestMulti {
		if testName == tn {
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			tApp.InitializeFromGenesisStates(
				NewPricefeedGenStateMulti(),
				NewCDPGenStateMulti(),
			)
			keeper := tApp.GetCDPKeeper()
			suite.app = tApp
			suite.ctx = ctx
			suite.keeper = keeper
			return
		}

	}
}

func (suite *KeeperTestSuite) TestAddCdp() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.app.GetAccountKeeper()
	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc.SetCoins(cs(c("xrp", 200), c("btc", 5)))
	ak.SetAccount(suite.ctx, acc)
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 200)), cs(c("usdx", 26)))
	suite.Equal(sdk.CodeType(6), err.Result().Code)
	err = suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 500)), cs(c("usdx", 26)))
	suite.Equal(sdk.CodeType(10), err.Result().Code)
	err = suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 200)), cs(c("xusd", 10)))
	suite.Equal(sdk.CodeType(4), err.Result().Code)

	err = suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 100)), cs(c("usdx", 10)))
	suite.NoError(err)
	tp := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.Equal(i(10), tp)
	sk := suite.app.GetSupplyKeeper()
	macc := sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("debt", 10), c("xrp", 100)), macc.GetCoins())
	acc = ak.GetAccount(suite.ctx, addrs[0])
	suite.Equal(cs(c("usdx", 10), c("xrp", 100), c("btc", 5)), acc.GetCoins())

	err = suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("btc", 5)), cs(c("usdx", 26667)))
	suite.Equal(sdk.CodeType(6), err.Result().Code)

	err = suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("btc", 5)), cs(c("usdx", 100)))
	suite.NoError(err)
	tp = suite.keeper.GetTotalPrincipal(suite.ctx, "btc", "usdx")
	suite.Equal(i(100), tp)
	macc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("debt", 110), c("xrp", 100), c("btc", 5)), macc.GetCoins())
	acc = ak.GetAccount(suite.ctx, addrs[0])
	suite.Equal(cs(c("usdx", 110), c("xrp", 100)), acc.GetCoins())

	err = suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("lol", 100)), cs(c("usdx", 10)))
	suite.Equal(sdk.CodeType(3), err.Result().Code)
	err = suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 100)), cs(c("usdx", 10)))
	suite.Equal(sdk.CodeType(1), err.Result().Code)
}

func (suite *KeeperTestSuite) TestGetSetDenomByte() {
	_, found := suite.keeper.GetDenomPrefix(suite.ctx, "xrp")
	suite.False(found)
	suite.keeper.SetParams(suite.ctx, params())
	db, found := suite.keeper.GetDenomPrefix(suite.ctx, "xrp")
	suite.True(found)
	suite.Equal(byte(0x20), db)
}

func (suite *KeeperTestSuite) TestGetDebtDenom() {
	suite.Panics(func() { suite.keeper.SetDebtDenom(suite.ctx, "") })
	t := suite.keeper.GetDebtDenom(suite.ctx)
	suite.Equal("debt", t)
	suite.keeper.SetDebtDenom(suite.ctx, "lol")
	t = suite.keeper.GetDebtDenom(suite.ctx)
	suite.Equal("lol", t)
}

func (suite *KeeperTestSuite) TestGetNextCdpID() {
	id := suite.keeper.GetNextCdpID(suite.ctx)
	suite.Equal(types.DefaultCdpStartingID, id)
}

func (suite *KeeperTestSuite) TestGetSetCdp() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), tmtime.Canonical(time.Now()))
	suite.keeper.SetCDP(suite.ctx, cdp)
	t, found := suite.keeper.GetCDP(suite.ctx, sdk.DefaultBondDenom, types.DefaultCdpStartingID)
	suite.True(found)
	suite.Equal(cdp, t)
	_, found = suite.keeper.GetCDP(suite.ctx, sdk.DefaultBondDenom, uint64(2))
	suite.False(found)
	suite.keeper.DeleteCDP(suite.ctx, cdp)
	_, found = suite.keeper.GetCDP(suite.ctx, sdk.DefaultBondDenom, types.DefaultCdpStartingID)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestGetSetCdpId() {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), tmtime.Canonical(time.Now()))
	suite.keeper.SetCDP(suite.ctx, cdp)
	suite.keeper.IndexCdpByOwner(suite.ctx, cdp)
	id, found := suite.keeper.GetCdpID(suite.ctx, addrs[0], sdk.DefaultBondDenom)
	suite.True(found)
	suite.Equal(types.DefaultCdpStartingID, id)
	_, found = suite.keeper.GetCdpID(suite.ctx, addrs[0], "lol")
	suite.False(found)
	_, found = suite.keeper.GetCdpID(suite.ctx, addrs[1], sdk.DefaultBondDenom)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestGetSetCdpByOwnerAndDenom() {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), tmtime.Canonical(time.Now()))
	suite.keeper.SetCDP(suite.ctx, cdp)
	suite.keeper.IndexCdpByOwner(suite.ctx, cdp)
	t, found := suite.keeper.GetCdpByOwnerAndDenom(suite.ctx, addrs[0], sdk.DefaultBondDenom)
	suite.True(found)
	suite.Equal(cdp, t)
	_, found = suite.keeper.GetCdpByOwnerAndDenom(suite.ctx, addrs[0], "lol")
	suite.False(found)
	_, found = suite.keeper.GetCdpByOwnerAndDenom(suite.ctx, addrs[1], sdk.DefaultBondDenom)
	suite.False(found)
	suite.NotPanics(func() { suite.keeper.IndexCdpByOwner(suite.ctx, cdp) })
}

func (suite *KeeperTestSuite) TestCalculateCollateralToDebtRatio() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(3))), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), tmtime.Canonical(time.Now()))
	cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
	suite.Equal(sdk.MustNewDecFromStr("3.0"), cr)
	cdp = types.NewCDP(types.DefaultCdpStartingID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2))), tmtime.Canonical(time.Now()))
	cr = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
	suite.Equal(sdk.MustNewDecFromStr("0.5"), cr)
	cdp = types.NewCDP(types.DefaultCdpStartingID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(3))), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2)), sdk.NewCoin("usdx", sdk.NewInt(1))), tmtime.Canonical(time.Now()))
	cr = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
	suite.Equal(sdk.MustNewDecFromStr("1"), cr)
}

func (suite *KeeperTestSuite) TestSetCdpByCollateralRatio() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(3))), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))), tmtime.Canonical(time.Now()))
	cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
	suite.NotPanics(func() { suite.keeper.IndexCdpByCollateralRatio(suite.ctx, cdp, cr) })
}

func (suite *KeeperTestSuite) TestIterateCdps() {
	cdps := cdps()
	for _, c := range cdps {
		suite.keeper.SetCDP(suite.ctx, c)
		suite.keeper.IndexCdpByOwner(suite.ctx, c)
		cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
		suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c, cr)
	}
	t := suite.keeper.GetAllCdps(suite.ctx)
	suite.Equal(4, len(t))
}

func (suite *KeeperTestSuite) TestIterateCdpsByDenom() {
	cdps := cdps()
	for _, c := range cdps {
		suite.keeper.SetCDP(suite.ctx, c)
		suite.keeper.IndexCdpByOwner(suite.ctx, c)
		cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
		suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c, cr)
	}
	xrpCdps := suite.keeper.GetAllCdpsByDenom(suite.ctx, "xrp")
	suite.Equal(3, len(xrpCdps))
	btcCdps := suite.keeper.GetAllCdpsByDenom(suite.ctx, "btc")
	suite.Equal(1, len(btcCdps))
	suite.keeper.DeleteCDP(suite.ctx, cdps[0])
	suite.keeper.RemoveCdpOwnerIndex(suite.ctx, cdps[0])
	xrpCdps = suite.keeper.GetAllCdpsByDenom(suite.ctx, "xrp")
	suite.Equal(2, len(xrpCdps))
	suite.keeper.DeleteCDP(suite.ctx, cdps[1])
	suite.keeper.RemoveCdpOwnerIndex(suite.ctx, cdps[1])
	ids, found := suite.keeper.GetCdpIdsByOwner(suite.ctx, cdps[1].Owner)
	suite.True(found)
	suite.Equal(1, len(ids))
	suite.Equal(uint64(3), ids[0])
}

func (suite *KeeperTestSuite) TestIterateCdpsByCollateralRatio() {
	cdps := cdps()
	for _, c := range cdps {
		suite.keeper.SetCDP(suite.ctx, c)
		suite.keeper.IndexCdpByOwner(suite.ctx, c)
		cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
		suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c, cr)
	}
	xrpCdps := suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("1.25"))
	suite.Equal(0, len(xrpCdps))
	xrpCdps = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("1.25").Add(sdk.SmallestDec()))
	suite.Equal(1, len(xrpCdps))
	xrpCdps = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("2.0").Add(sdk.SmallestDec()))
	suite.Equal(2, len(xrpCdps))
	xrpCdps = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("100.0").Add(sdk.SmallestDec()))
	suite.Equal(3, len(xrpCdps))
	suite.keeper.DeleteCDP(suite.ctx, cdps[0])
	suite.keeper.RemoveCdpOwnerIndex(suite.ctx, cdps[0])
	suite.keeper.RemoveCdpLiquidationRatioIndex(suite.ctx, cdps[0])
	xrpCdps = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("2.0").Add(sdk.SmallestDec()))
	suite.Equal(1, len(xrpCdps))
}

func (suite *KeeperTestSuite) TestValidateCollateral() {
	c := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
	err := suite.keeper.ValidateCollateral(suite.ctx, c)
	suite.NoError(err)
	c = sdk.NewCoins(sdk.NewCoin("lol", sdk.NewInt(1)))
	err = suite.keeper.ValidateCollateral(suite.ctx, c)
	suite.Error(err)
	c = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)), sdk.NewCoin("xrp", sdk.NewInt(1)))
	err = suite.keeper.ValidateCollateral(suite.ctx, c)
	suite.Error(err)
}

func (suite *KeeperTestSuite) TestValidatePrincipal() {
	d := sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1)))
	err := suite.keeper.ValidatePrincipal(suite.ctx, d)
	suite.NoError(err)
	d = sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1)), sdk.NewCoin("susd", sdk.NewInt(1)))
	err = suite.keeper.ValidatePrincipal(suite.ctx, d)
	suite.NoError(err)
	d = sdk.NewCoins(sdk.NewCoin("xusd", sdk.NewInt(1)))
	err = suite.keeper.ValidatePrincipal(suite.ctx, d)
	suite.Error(err)
	d = sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(1000000000001)))
	err = suite.keeper.ValidatePrincipal(suite.ctx, d)
	suite.Error(err)
}

func (suite *KeeperTestSuite) TestCalculateCollateralizationRatio() {
	c := cdps()[1]
	suite.keeper.SetCDP(suite.ctx, c)
	suite.keeper.IndexCdpByOwner(suite.ctx, c)
	cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
	suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c, cr)
	cr, err := suite.keeper.CalculateCollateralizationRatio(suite.ctx, c.Collateral, c.Principal, c.AccumulatedFees)
	suite.NoError(err)
	suite.Equal(d("2.5"), cr)
	c.AccumulatedFees = sdk.NewCoins(sdk.NewCoin("usdx", i(10)))
	cr, err = suite.keeper.CalculateCollateralizationRatio(suite.ctx, c.Collateral, c.Principal, c.AccumulatedFees)
	suite.NoError(err)
	suite.Equal(d("1.25"), cr)
}

func (suite *KeeperTestSuite) TestMintBurnDebtCoins() {
	cd := cdps()[1]
	err := suite.keeper.MintDebtCoins(suite.ctx, types.ModuleName, suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	suite.NoError(err)
	err = suite.keeper.MintDebtCoins(suite.ctx, "notamodule", suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	suite.Error(err)
	sk := suite.app.GetSupplyKeeper()
	acc := sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("debt", 10)), acc.GetCoins())

	err = suite.keeper.BurnDebtCoins(suite.ctx, types.ModuleName, suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	suite.NoError(err)
	err = suite.keeper.BurnDebtCoins(suite.ctx, "notamodule", suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	suite.Error(err)
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(sdk.Coins(nil), acc.GetCoins())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
