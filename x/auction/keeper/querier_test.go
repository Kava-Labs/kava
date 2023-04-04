package keeper_test

import (
	"math/rand"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/testutil"
	"github.com/kava-labs/kava/x/auction/types"
)

const (
	custom           = "custom"
	TestAuctionCount = 10
)

type querierTestSuite struct {
	testutil.Suite

	auctions    []types.Auction
	legacyAmino *codec.LegacyAmino
	querier     sdk.Querier
}

func (suite *querierTestSuite) SetupTest() {
	suite.Suite.SetupTest(10)
	// Populate with auctions
	for j := 0; j < TestAuctionCount; j++ {
		var id uint64
		var err error
		lotAmount := int64(rand.Intn(100-10) + 10)

		// Add coins required for auction creation to module account
		suite.AddCoinsToNamedModule(suite.ModAcc.Name, cs(c("token1", lotAmount), c("usdx", 20), c("debt", 10)))

		ownerAddrIndex := rand.Intn(9-1) + 1
		if ownerAddrIndex%2 == 0 {
			id, err = suite.Keeper.StartSurplusAuction(suite.Ctx, suite.ModAcc.Name, c("token1", lotAmount), "token2")
		} else {
			id, err = suite.Keeper.StartCollateralAuction(suite.Ctx, suite.ModAcc.Name, c("token1", lotAmount), c("usdx", int64(20)),
				[]sdk.AccAddress{suite.Addrs[ownerAddrIndex]}, []sdkmath.Int{sdkmath.NewInt(lotAmount)}, c("debt", int64(10)))
		}
		suite.NoError(err)

		auc, found := suite.Keeper.GetAuction(suite.Ctx, id)
		suite.True(found)
		suite.auctions = append(suite.auctions, auc)
	}
	suite.legacyAmino = suite.App.LegacyAmino()
	suite.querier = keeper.NewQuerier(suite.Keeper, suite.legacyAmino)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(querierTestSuite))
}

func (suite *querierTestSuite) assertQuerierResponse(expected interface{}, actual []byte) {
	expectedJson, err := suite.legacyAmino.MarshalJSONIndent(expected, "", "  ")
	suite.Require().NoError(err)
	suite.Require().Equal(string(expectedJson), string(actual))
}

func (suite *querierTestSuite) TestQueryParams() {
	bz, err := suite.querier(suite.Ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Require().NoError(err)
	suite.Require().NotNil(bz)

	var params types.Params
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(bz, &params))

	expectedParams := suite.Keeper.GetParams(suite.Ctx)
	suite.Require().Equal(expectedParams, params)
}

func (suite *querierTestSuite) TestQueryAuction() {
	ctx := suite.Ctx.WithIsCheckTx(false)

	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAuction}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(types.NewQueryAuctionParams(suite.auctions[0].GetID())),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAuction}, query)
	suite.NoError(err)
	suite.NotNil(bz)
	suite.assertQuerierResponse(suite.auctions[0], bz)
}

func (suite *querierTestSuite) TestQueryAuctions() {
	ctx := suite.Ctx.WithIsCheckTx(false)

	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAuctions}, "/"),
		Data: suite.legacyAmino.MustMarshalJSON(
			types.NewQueryAllAuctionParams(1, TestAuctionCount, "", "", "", nil),
		),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAuctions}, query)
	suite.Require().NoError(err)
	suite.Require().NotNil(bz)

	suite.assertQuerierResponse(suite.Keeper.GetAllAuctions(suite.Ctx), bz)
}

func (suite *querierTestSuite) TestQueryNextAuctionID() {
	bz, err := suite.querier(suite.Ctx, []string{types.QueryNextAuctionID}, abci.RequestQuery{})
	suite.Require().NoError(err)
	suite.Require().NotNil(bz)

	var nextAuctionID uint64
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(bz, &nextAuctionID))

	expectedID, _ := suite.Keeper.GetNextAuctionID(suite.Ctx)
	suite.Require().Equal(expectedID, nextAuctionID)
}
