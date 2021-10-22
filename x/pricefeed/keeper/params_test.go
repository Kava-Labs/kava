package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	addrs  []string
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmprototypes.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(),
	)
	suite.keeper = tApp.GetPriceFeedKeeper()
	suite.ctx = ctx

	_, addrs := app.GeneratePrivKeyAddressPairs(10)
	var strAddrs []string
	for _, addr := range addrs {
		strAddrs = append(strAddrs, addr.String())
	}

	suite.addrs = strAddrs
}

func (suite *KeeperTestSuite) TestGetSetOracles() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Equal([]string(nil), params.Markets[0].Oracles)

	params.Markets[0].Oracles = suite.addrs
	suite.NotPanics(func() { suite.keeper.SetParams(suite.ctx, params) })
	params = suite.keeper.GetParams(suite.ctx)
	suite.Equal(suite.addrs, params.Markets[0].Oracles)

	addr, err := suite.keeper.GetOracle(suite.ctx, params.Markets[0].MarketId, suite.addrs[0])
	suite.NoError(err)
	suite.Equal(suite.addrs[0], addr)
}

func (suite *KeeperTestSuite) TestGetAuthorizedAddresses() {
	_, oracles := app.GeneratePrivKeyAddressPairs(5)
	var strOracles []string
	for _, addr := range oracles {
		strOracles = append(strOracles, addr.String())
	}

	params := types.Params{
		Markets: []types.Market{
			{MarketId: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: strOracles[:3], Active: true},
			{MarketId: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: strOracles[2:], Active: true},
			{MarketId: "xrp:usd:30", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: nil, Active: true},
		},
	}
	suite.keeper.SetParams(suite.ctx, params)

	actualOracles := suite.keeper.GetAuthorizedAddresses(suite.ctx)

	suite.Require().ElementsMatch(strOracles, actualOracles)
}
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
