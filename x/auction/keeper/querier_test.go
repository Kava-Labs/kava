package keeper_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/supply"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/cdp"
)

const (
	custom           = "custom"
	TestAuctionCount = 10
)

type QuerierTestSuite struct {
	suite.Suite

	keeper   keeper.Keeper
	app      app.TestApp
	auctions types.Auctions
	ctx      sdk.Context
	querier  sdk.Querier
}

func (suite *QuerierTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	buyer := addrs[0]
	modName := cdp.LiquidatorMacc

	// Set up seller account
	sellerAcc := supply.NewEmptyModuleAccount(modName, supply.Minter, supply.Burner)
	sellerAcc.SetCoins(cs(c("token1", 1000), c("token2", 1000), c("debt", 1000)))

	// Initialize genesis accounts
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(buyer, cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
			sellerAcc,
		}),
	)

	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = tApp.GetAuctionKeeper()

	// Populate with auctions
	randSrc := rand.New(rand.NewSource(int64(1234)))
	for j := 0; j < TestAuctionCount; j++ {
		lotAmount := simulation.RandIntBetween(randSrc, 10, 100)
		id, err := suite.keeper.StartSurplusAuction(suite.ctx, modName, c("token1", int64(lotAmount)), "token2")
		suite.NoError(err)

		auc, found := suite.keeper.GetAuction(suite.ctx, id)
		suite.True(found)
		suite.auctions = append(suite.auctions, auc)
	}

	suite.querier = keeper.NewQuerier(suite.keeper)
}

func (suite *QuerierTestSuite) TestQueryAuction() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAuction}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.QueryAuctionParams{AuctionID: types.DefaultNextAuctionID}), // get the first auction
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAuction}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes into type Auction
	var auction types.Auction
	suite.NoError(types.ModuleCdc.UnmarshalJSON(bz, &auction))

	// Check the returned auction
	suite.Equal(suite.auctions[0].GetID(), auction.GetID())
	suite.Equal(suite.auctions[0].GetInitiator(), auction.GetInitiator())
	suite.Equal(suite.auctions[0].GetLot(), auction.GetLot())
	suite.Equal(suite.auctions[0].GetBid(), auction.GetBid())
	suite.Equal(suite.auctions[0].GetEndTime(), auction.GetEndTime())

}

func (suite *QuerierTestSuite) TestQueryAuctions() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAuctions}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryAllAuctionParams(1, TestAuctionCount)),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAuctions}, query)
	suite.NoError(err)
	suite.NotNil(bz)

	// Unmarshal the bytes into type Auctions
	var auctions types.Auctions
	suite.NoError(types.ModuleCdc.UnmarshalJSON(bz, &auctions))

	// Check that each Auction has correct values
	if len(auctions) == 0 && len(suite.auctions) != 0 {
		suite.FailNow("no auctions returned") // skip the panic from indexing empty slice below
	}
	for i := 0; i < TestAuctionCount; i++ {
		suite.Equal(suite.auctions[i].GetID(), auctions[i].GetID())
		suite.Equal(suite.auctions[i].GetInitiator(), auctions[i].GetInitiator())
		suite.Equal(suite.auctions[i].GetLot(), auctions[i].GetLot())
		suite.Equal(suite.auctions[i].GetBid(), auctions[i].GetBid())
		suite.Equal(suite.auctions[i].GetEndTime(), auctions[i].GetEndTime())
	}
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}
