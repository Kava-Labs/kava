package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetBep3Keeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	return
}

func (suite *KeeperTestSuite) TestGetSetHtlt() {
	htlt := types.NewHTLT(binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHashes[0], timestamps[0], coinsSingle, "bnb50000", 80000, false)
	swapID, err := types.CalculateSwapID(htlt.RandomNumberHash, htlt.From, htlt.SenderOtherChain)
	suite.NoError(err)
	suite.keeper.SetHTLT(suite.ctx, htlt, swapID)

	h, found := suite.keeper.GetHTLT(suite.ctx, swapID)
	suite.True(found)
	suite.Equal(htlt, h)

	fakeSwapID, err := types.CalculateSwapID(htlt.RandomNumberHash, kavaAddrs[1], "otheraddress")
	suite.NoError(err)
	_, found = suite.keeper.GetHTLT(suite.ctx, fakeSwapID)
	suite.False(found)

	suite.keeper.DeleteHTLT(suite.ctx, swapID)
	_, found = suite.keeper.GetHTLT(suite.ctx, swapID)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestIterateHtlts() {
	htlts := htlts()
	for _, h := range htlts {
		swapID, err := types.CalculateSwapID(h.RandomNumberHash, h.From, h.SenderOtherChain)
		suite.NoError(err)
		suite.keeper.SetHTLT(suite.ctx, h, swapID)
	}
	res := suite.keeper.GetAllHtlts(suite.ctx)
	suite.Equal(4, len(res))
}

func TestHtltTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func htlts() types.HTLTs {
	var htlts types.HTLTs
	h1 := types.NewHTLT(binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHashes[0], timestamps[0], coinsSingle, "bnb50000", 50500, false)
	h2 := types.NewHTLT(binanceAddrs[1], kavaAddrs[1], "", "", randomNumberHashes[1], timestamps[1], coinsSingle, "bnb50000", 61500, false)
	h3 := types.NewHTLT(binanceAddrs[2], kavaAddrs[2], "", "", randomNumberHashes[2], timestamps[2], coinsSingle, "bnb50000", 72500, false)
	h4 := types.NewHTLT(binanceAddrs[3], kavaAddrs[3], "", "", randomNumberHashes[3], timestamps[3], coinsSingle, "bnb50000", 83500, false)
	htlts = append(htlts, h1, h2, h3, h4)
	return htlts
}
