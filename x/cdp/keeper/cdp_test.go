package keeper_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
)

type CdpTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *CdpTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
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

func (suite *CdpTestSuite) TestAddCdp() {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	ak := suite.app.GetAccountKeeper()
	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc.SetCoins(cs(c("xrp", 200000000), c("btc", 500000000)))
	ak.SetAccount(suite.ctx, acc)
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 200000000), c("usdx", 26000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidCollateralRatio))
	err = suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 500000000), c("usdx", 26000000))
	suite.Error(err) // insufficient balance
	err = suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 200000000), c("xusd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrDebtNotSupported))

	acc2 := ak.NewAccountWithAddress(suite.ctx, addrs[1])
	acc2.SetCoins(cs(c("btc", 500000000000)))
	ak.SetAccount(suite.ctx, acc2)
	err = suite.keeper.AddCdp(suite.ctx, addrs[1], c("btc", 500000000000), c("usdx", 500000000001))
	suite.Require().True(errors.Is(err, types.ErrExceedsDebtLimit))

	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour * 2))
	pk := suite.app.GetPriceFeedKeeper()
	err = pk.SetCurrentPrices(ctx, "xrp:usd")
	suite.Error(err)
	ok := suite.keeper.UpdatePricefeedStatus(ctx, "xrp:usd")
	suite.False(ok)
	err = suite.keeper.AddCdp(ctx, addrs[0], c("xrp", 100000000), c("usdx", 10000000))
	suite.Require().True(errors.Is(err, types.ErrPricefeedDown))

	err = pk.SetCurrentPrices(suite.ctx, "xrp:usd")
	ok = suite.keeper.UpdatePricefeedStatus(suite.ctx, "xrp:usd")
	suite.True(ok)
	suite.NoError(err)
	err = suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 100000000), c("usdx", 10000000))
	suite.NoError(err)
	id := suite.keeper.GetNextCdpID(suite.ctx)
	suite.Equal(uint64(2), id)
	tp := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.Equal(i(10000000), tp)
	sk := suite.app.GetSupplyKeeper()
	macc := sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("debt", 10000000), c("xrp", 100000000)), macc.GetCoins())
	acc = ak.GetAccount(suite.ctx, addrs[0])
	suite.Equal(cs(c("usdx", 10000000), c("xrp", 100000000), c("btc", 500000000)), acc.GetCoins())

	err = suite.keeper.AddCdp(suite.ctx, addrs[0], c("btc", 500000000), c("usdx", 26667000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidCollateralRatio))

	err = suite.keeper.AddCdp(suite.ctx, addrs[0], c("btc", 500000000), c("usdx", 100000000))
	suite.NoError(err)
	id = suite.keeper.GetNextCdpID(suite.ctx)
	suite.Equal(uint64(3), id)
	tp = suite.keeper.GetTotalPrincipal(suite.ctx, "btc", "usdx")
	suite.Equal(i(100000000), tp)
	macc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("debt", 110000000), c("xrp", 100000000), c("btc", 500000000)), macc.GetCoins())
	acc = ak.GetAccount(suite.ctx, addrs[0])
	suite.Equal(cs(c("usdx", 110000000), c("xrp", 100000000)), acc.GetCoins())

	err = suite.keeper.AddCdp(suite.ctx, addrs[0], c("lol", 100), c("usdx", 10))
	suite.Require().True(errors.Is(err, types.ErrCollateralNotSupported))
	err = suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 100), c("usdx", 10))
	suite.Require().True(errors.Is(err, types.ErrCdpAlreadyExists))
}

func (suite *CdpTestSuite) TestGetSetDenomByte() {
	_, found := suite.keeper.GetDenomPrefix(suite.ctx, "lol")
	suite.False(found)
	db, found := suite.keeper.GetDenomPrefix(suite.ctx, "xrp")
	suite.True(found)
	suite.Equal(byte(0x20), db)
}

func (suite *CdpTestSuite) TestGetDebtDenom() {
	suite.Panics(func() { suite.keeper.SetDebtDenom(suite.ctx, "") })
	t := suite.keeper.GetDebtDenom(suite.ctx)
	suite.Equal("debt", t)
	suite.keeper.SetDebtDenom(suite.ctx, "lol")
	t = suite.keeper.GetDebtDenom(suite.ctx)
	suite.Equal("lol", t)
}

func (suite *CdpTestSuite) TestGetNextCdpID() {
	id := suite.keeper.GetNextCdpID(suite.ctx)
	suite.Equal(types.DefaultCdpStartingID, id)
}

func (suite *CdpTestSuite) TestGetSetCdp() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], c("xrp", 1), c("usdx", 1), tmtime.Canonical(time.Now()))
	err := suite.keeper.SetCDP(suite.ctx, cdp)
	suite.NoError(err)

	t, found := suite.keeper.GetCDP(suite.ctx, "xrp", types.DefaultCdpStartingID)
	suite.True(found)
	suite.Equal(cdp, t)
	_, found = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	suite.False(found)
	suite.keeper.DeleteCDP(suite.ctx, cdp)
	_, found = suite.keeper.GetCDP(suite.ctx, "btc", types.DefaultCdpStartingID)
	suite.False(found)
}

func (suite *CdpTestSuite) TestGetSetCdpId() {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], c("xrp", 1), c("usdx", 1), tmtime.Canonical(time.Now()))
	err := suite.keeper.SetCDP(suite.ctx, cdp)
	suite.NoError(err)
	suite.keeper.IndexCdpByOwner(suite.ctx, cdp)
	id, found := suite.keeper.GetCdpID(suite.ctx, addrs[0], "xrp")
	suite.True(found)
	suite.Equal(types.DefaultCdpStartingID, id)
	_, found = suite.keeper.GetCdpID(suite.ctx, addrs[0], "lol")
	suite.False(found)
	_, found = suite.keeper.GetCdpID(suite.ctx, addrs[1], "xrp")
	suite.False(found)
}

func (suite *CdpTestSuite) TestGetSetCdpByOwnerAndDenom() {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], c("xrp", 1), c("usdx", 1), tmtime.Canonical(time.Now()))
	err := suite.keeper.SetCDP(suite.ctx, cdp)
	suite.NoError(err)
	suite.keeper.IndexCdpByOwner(suite.ctx, cdp)
	t, found := suite.keeper.GetCdpByOwnerAndDenom(suite.ctx, addrs[0], "xrp")
	suite.True(found)
	suite.Equal(cdp, t)
	_, found = suite.keeper.GetCdpByOwnerAndDenom(suite.ctx, addrs[0], "lol")
	suite.False(found)
	_, found = suite.keeper.GetCdpByOwnerAndDenom(suite.ctx, addrs[1], "xrp")
	suite.False(found)
	suite.NotPanics(func() { suite.keeper.IndexCdpByOwner(suite.ctx, cdp) })
}

func (suite *CdpTestSuite) TestCalculateCollateralToDebtRatio() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], c("xrp", 3), c("usdx", 1), tmtime.Canonical(time.Now()))
	cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
	suite.Equal(sdk.MustNewDecFromStr("3.0"), cr)
	cdp = types.NewCDP(types.DefaultCdpStartingID, addrs[0], c("xrp", 1), c("usdx", 2), tmtime.Canonical(time.Now()))
	cr = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
	suite.Equal(sdk.MustNewDecFromStr("0.5"), cr)
}

func (suite *CdpTestSuite) TestSetCdpByCollateralRatio() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	cdp := types.NewCDP(types.DefaultCdpStartingID, addrs[0], c("xrp", 3), c("usdx", 1), tmtime.Canonical(time.Now()))
	cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
	suite.NotPanics(func() { suite.keeper.IndexCdpByCollateralRatio(suite.ctx, cdp.Collateral.Denom, cdp.ID, cr) })
}

func (suite *CdpTestSuite) TestIterateCdps() {
	cdps := cdps()
	for _, c := range cdps {
		err := suite.keeper.SetCDP(suite.ctx, c)
		suite.NoError(err)
		suite.keeper.IndexCdpByOwner(suite.ctx, c)
		cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
		suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c.Collateral.Denom, c.ID, cr)
	}
	t := suite.keeper.GetAllCdps(suite.ctx)
	suite.Equal(4, len(t))
}

func (suite *CdpTestSuite) TestIterateCdpsByDenom() {
	cdps := cdps()
	for _, c := range cdps {
		err := suite.keeper.SetCDP(suite.ctx, c)
		suite.NoError(err)
		suite.keeper.IndexCdpByOwner(suite.ctx, c)
		cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
		suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c.Collateral.Denom, c.ID, cr)
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

func (suite *CdpTestSuite) TestIterateCdpsByCollateralRatio() {
	cdps := cdps()
	for _, c := range cdps {
		err := suite.keeper.SetCDP(suite.ctx, c)
		suite.NoError(err)
		suite.keeper.IndexCdpByOwner(suite.ctx, c)
		cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
		suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c.Collateral.Denom, c.ID, cr)
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
	cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdps[0].Collateral, cdps[0].Principal)
	suite.keeper.RemoveCdpCollateralRatioIndex(suite.ctx, cdps[0].Collateral.Denom, cdps[0].ID, cr)
	xrpCdps = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("2.0").Add(sdk.SmallestDec()))
	suite.Equal(1, len(xrpCdps))
}

func (suite *CdpTestSuite) TestValidateCollateral() {
	c := sdk.NewCoin("xrp", sdk.NewInt(1))
	err := suite.keeper.ValidateCollateral(suite.ctx, c)
	suite.NoError(err)
	c = sdk.NewCoin("lol", sdk.NewInt(1))
	err = suite.keeper.ValidateCollateral(suite.ctx, c)
	suite.Require().True(errors.Is(err, types.ErrCollateralNotSupported))
}

func (suite *CdpTestSuite) TestValidatePrincipal() {
	d := sdk.NewCoin("usdx", sdk.NewInt(10000000))
	err := suite.keeper.ValidatePrincipalAdd(suite.ctx, d)
	suite.NoError(err)
	d = sdk.NewCoin("xusd", sdk.NewInt(1))
	err = suite.keeper.ValidatePrincipalAdd(suite.ctx, d)
	suite.Require().True(errors.Is(err, types.ErrDebtNotSupported))
	d = sdk.NewCoin("usdx", sdk.NewInt(1000000000001))
	err = suite.keeper.ValidateDebtLimit(suite.ctx, "xrp", d)
	suite.Require().True(errors.Is(err, types.ErrExceedsDebtLimit))
	d = sdk.NewCoin("usdx", sdk.NewInt(100000000))
	err = suite.keeper.ValidateDebtLimit(suite.ctx, "xrp", d)
	suite.NoError(err)
}

func (suite *CdpTestSuite) TestCalculateCollateralizationRatio() {
	c := cdps()[1]
	err := suite.keeper.SetCDP(suite.ctx, c)
	suite.NoError(err)
	suite.keeper.IndexCdpByOwner(suite.ctx, c)
	cr := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, c.Collateral, c.Principal)
	suite.keeper.IndexCdpByCollateralRatio(suite.ctx, c.Collateral.Denom, c.ID, cr)
	cr, err = suite.keeper.CalculateCollateralizationRatio(suite.ctx, c.Collateral, c.Principal, c.AccumulatedFees, "spot")
	suite.NoError(err)
	suite.Equal(d("2.5"), cr)
	c.AccumulatedFees = sdk.NewCoin("usdx", i(10000000))
	cr, err = suite.keeper.CalculateCollateralizationRatio(suite.ctx, c.Collateral, c.Principal, c.AccumulatedFees, "spot")
	suite.NoError(err)
	suite.Equal(d("1.25"), cr)
}

func (suite *CdpTestSuite) TestMintBurnDebtCoins() {
	cd := cdps()[1]
	err := suite.keeper.MintDebtCoins(suite.ctx, types.ModuleName, suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	suite.NoError(err)
	suite.Require().Panics(func() {
		_ = suite.keeper.MintDebtCoins(suite.ctx, "notamodule", suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	})

	sk := suite.app.GetSupplyKeeper()
	acc := sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("debt", 10000000)), acc.GetCoins())

	err = suite.keeper.BurnDebtCoins(suite.ctx, types.ModuleName, suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	suite.NoError(err)
	suite.Require().Panics(func() {
		_ = suite.keeper.BurnDebtCoins(suite.ctx, "notamodule", suite.keeper.GetDebtDenom(suite.ctx), cd.Principal)
	})
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(sdk.Coins(nil), acc.GetCoins())
}

func TestCdpTestSuite(t *testing.T) {
	suite.Run(t, new(CdpTestSuite))
}
