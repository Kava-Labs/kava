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

func (suite *KeeperTestSuite) TestGetSetAtomicSwap() {
	heightSpan := int64(1000)
	expirationBlock := uint64(suite.ctx.BlockHeight()) + uint64(heightSpan)
	atomicSwap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHashes[0], int64(expirationBlock), timestamps[0], binanceAddrs[0], kavaAddrs[0], "", "", 0, types.Open)
	suite.keeper.SetAtomicSwap(suite.ctx, atomicSwap)

	s, found := suite.keeper.GetAtomicSwap(suite.ctx, atomicSwap.GetSwapID())
	suite.True(found)
	suite.Equal(atomicSwap, s)

	fakeSwapID := types.CalculateSwapID(atomicSwap.RandomNumberHash, kavaAddrs[1], "otheraddress")
	_, found = suite.keeper.GetAtomicSwap(suite.ctx, fakeSwapID)
	suite.False(found)
}

// TODO: RemoveAtomicSwap

func (suite *KeeperTestSuite) TestIterateAtomicSwaps() {
	atomicSwaps := atomicSwaps(4)
	for _, s := range atomicSwaps {
		suite.keeper.SetAtomicSwap(suite.ctx, s)
	}
	res := suite.keeper.GetAllAtomicSwaps(suite.ctx)
	suite.Equal(4, len(res))
}

// TODO: GetAllAtomicSwaps

// TODO: InsertIntoByBlockIndex
// TODO: RemoveFromByBlockIndex
// TODO: IterateAtomicSwapsByBlock

// TODO: SetAssetSupply
// TODO: GetAssetSupply
// TODO: IterateAssetSupplies
// TODO: GetAllAssetSupplies

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
