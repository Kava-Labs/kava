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
	BeforeTestMulti = []string{"TestIterateCdpsByDenom", "TestIterateCdpsByCollateralRatio"}
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

func (suite *KeeperTestSuite) TestGetSetDenomByte() {
	_, found := suite.keeper.GetDenomPrefix(suite.ctx, "xrp")
	suite.False(found)
	suite.keeper.SetParams(suite.ctx, params())
	db, found := suite.keeper.GetDenomPrefix(suite.ctx, "xrp")
	suite.True(found)
	suite.Equal(byte(0x20), db)
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
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
