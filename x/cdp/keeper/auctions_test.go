package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type AuctionTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *AuctionTestSuite) SetupTest() {
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

func (suite *AuctionTestSuite) TestNetDebtSurplus() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100)))
	suite.NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 10)))
	suite.NoError(err)
	suite.NotPanics(func() { suite.keeper.NetSurplusAndDebt(suite.ctx) })
	acc := sk.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("debt", 90)), acc.GetCoins())
}

func (suite *AuctionTestSuite) TestSurplusAuction() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 600000000000)))
	suite.NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 100000000000)))
	suite.NoError(err)
	suite.keeper.RunSurplusAndDebtAuctions(suite.ctx)
	acc := sk.GetModuleAccount(suite.ctx, auction.ModuleName)
	suite.Equal(cs(c("usdx", 10000000000)), acc.GetCoins())
	acc = sk.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("usdx", 490000000000)), acc.GetCoins())
}

func (suite *AuctionTestSuite) TestDebtAuction() {
	sk := suite.app.GetSupplyKeeper()
	err := sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("usdx", 100000000000)))
	suite.NoError(err)
	err = sk.MintCoins(suite.ctx, types.LiquidatorMacc, cs(c("debt", 200000000000)))
	suite.NoError(err)
	suite.keeper.RunSurplusAndDebtAuctions(suite.ctx)
	acc := sk.GetModuleAccount(suite.ctx, auction.ModuleName)
	suite.Equal(cs(c("debt", 10000000000)), acc.GetCoins())
	acc = sk.GetModuleAccount(suite.ctx, types.LiquidatorMacc)
	suite.Equal(cs(c("debt", 90000000000)), acc.GetCoins())
}

func TestAuctionTestSuite(t *testing.T) {
	suite.Run(t, new(AuctionTestSuite))
}
