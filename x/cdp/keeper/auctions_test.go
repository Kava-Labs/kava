package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	auctiontypes "github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type AuctionTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

func (suite *AuctionTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()
	taddr := sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser1")))
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	authGS := app.NewFundedGenStateWithSameCoins(tApp.AppCodec(), cs(c("usdx", 21000000000)), []sdk.AccAddress{taddr})
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(tApp.AppCodec()),
		NewCDPGenStateMulti(tApp.AppCodec()),
	)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = []sdk.AccAddress{taddr}
	return
}

func (suite *AuctionTestSuite) TestNetDebtSurplus() {
	bk := suite.app.GetBankKeeper()
	ak := suite.app.GetAccountKeeper()

	err := bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100)))
	suite.NoError(err)
	err = bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 10)))
	suite.NoError(err)
	suite.NotPanics(func() { suite.keeper.NetSurplusAndDebt(suite.ctx) })
	acc := ak.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("debt", 90)), bk.GetAllBalances(suite.ctx, acc.GetAddress()))
}

func (suite *AuctionTestSuite) TestCollateralAuction() {
	bk := suite.app.GetBankKeeper()
	err := bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 21000000000), c("bnb", 190000000000)))
	suite.Require().NoError(err)
	testDeposit := types.NewDeposit(1, suite.addrs[0], c("bnb", 190000000000))
	err = suite.keeper.AuctionCollateral(suite.ctx, types.Deposits{testDeposit}, "bnb-a", i(21000000000), "usdx")
	suite.Require().NoError(err)
}

func (suite *AuctionTestSuite) TestSurplusAuction() {
	bk := suite.app.GetBankKeeper()
	ak := suite.app.GetAccountKeeper()

	err := bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 600000000000)))
	suite.NoError(err)
	err = bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100000000000)))
	suite.NoError(err)
	suite.keeper.RunSurplusAndDebtAuctions(suite.ctx)
	acc := ak.GetModuleAccount(suite.ctx, auctiontypes.ModuleName)
	suite.Equal(cs(c("usdx", 10000000000)), bk.GetAllBalances(suite.ctx, acc.GetAddress()))
	acc = ak.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("usdx", 490000000000)), bk.GetAllBalances(suite.ctx, acc.GetAddress()))
}

func (suite *AuctionTestSuite) TestDebtAuction() {
	bk := suite.app.GetBankKeeper()
	ak := suite.app.GetAccountKeeper()

	err := bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 100000000000)))
	suite.NoError(err)
	err = bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 200000000000)))
	suite.NoError(err)
	suite.keeper.RunSurplusAndDebtAuctions(suite.ctx)
	acc := ak.GetModuleAccount(suite.ctx, auctiontypes.ModuleName)
	suite.Equal(cs(c("debt", 10000000000)), bk.GetAllBalances(suite.ctx, acc.GetAddress()))
	acc = ak.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("debt", 90000000000)), bk.GetAllBalances(suite.ctx, acc.GetAddress()))
}

func (suite *AuctionTestSuite) TestGetTotalSurplus() {
	bk := suite.app.GetBankKeeper()

	// liquidator account has zero coins
	suite.Require().Equal(sdk.NewInt(0), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// mint some coins
	err := bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 100e6)))
	suite.Require().NoError(err)
	err = bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 200e6)))
	suite.Require().NoError(err)

	// liquidator account has 300e6 total usdx
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// mint some debt
	err = bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 500e6)))
	suite.Require().NoError(err)

	// liquidator account still has 300e6 total usdx -- debt balance is ignored
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// burn some usdx
	err = bk.BurnCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 50e6)))
	suite.Require().NoError(err)

	// liquidator usdx decreases
	suite.Require().Equal(sdk.NewInt(250e6), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))
}

func (suite *AuctionTestSuite) TestGetTotalDebt() {
	bk := suite.app.GetBankKeeper()

	// liquidator account has zero debt
	suite.Require().Equal(sdk.NewInt(0), suite.keeper.GetTotalSurplus(suite.ctx, types.LiquidatorMacc))

	// mint some debt
	err := bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100e6)))
	suite.Require().NoError(err)
	err = bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 200e6)))
	suite.Require().NoError(err)

	// liquidator account has 300e6 total debt
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalDebt(suite.ctx, types.LiquidatorMacc))

	// mint some usdx
	err = bk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 500e6)))
	suite.Require().NoError(err)

	// liquidator account still has 300e6 total debt -- usdx balance is ignored
	suite.Require().Equal(sdk.NewInt(300e6), suite.keeper.GetTotalDebt(suite.ctx, types.LiquidatorMacc))

	// burn some debt
	err = bk.BurnCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 50e6)))
	suite.Require().NoError(err)

	// liquidator debt decreases
	suite.Require().Equal(sdk.NewInt(250e6), suite.keeper.GetTotalDebt(suite.ctx, types.LiquidatorMacc))
}

func TestAuctionTestSuite(t *testing.T) {
	suite.Run(t, new(AuctionTestSuite))
}
