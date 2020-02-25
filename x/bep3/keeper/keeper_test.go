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
	atomicSwap := types.NewAtomicSwap(coinsSingle, randomNumberHashes[0], int64(expirationBlock), timestamps[0], binanceAddrs[0], kavaAddrs[0], "", 0, types.Open)
	suite.keeper.SetAtomicSwap(suite.ctx, atomicSwap)

	s, found := suite.keeper.GetAtomicSwap(suite.ctx, atomicSwap.GetSwapID())
	suite.True(found)
	suite.Equal(atomicSwap, s)

	fakeSwapID := types.CalculateSwapID(atomicSwap.RandomNumberHash, kavaAddrs[1], "otheraddress")
	_, found = suite.keeper.GetAtomicSwap(suite.ctx, fakeSwapID)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestIterateAtomicSwaps() {
	atomicSwaps := atomicSwaps(4)
	for _, s := range atomicSwaps {
		suite.keeper.SetAtomicSwap(suite.ctx, s)
	}
	res := suite.keeper.GetAllAtomicSwaps(suite.ctx)
	suite.Equal(4, len(res))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func atomicSwaps(count int) types.AtomicSwaps {
	var atomicSwaps types.AtomicSwaps

	var swapIDs [][]byte
	for i := 0; i < count; i++ {
		swapID := types.CalculateSwapID(randomNumberHashes[i], binanceAddrs[i], "")
		swapIDs = append(swapIDs, swapID)
	}
	s1 := types.NewAtomicSwap(coinsSingle, randomNumberHashes[0], int64(100), timestamps[0], binanceAddrs[0], kavaAddrs[0], "", 0, types.Open)
	s2 := types.NewAtomicSwap(coinsSingle, randomNumberHashes[1], int64(275), timestamps[1], binanceAddrs[1], kavaAddrs[1], "", 0, types.Open)
	s3 := types.NewAtomicSwap(coinsSingle, randomNumberHashes[2], int64(325), timestamps[2], binanceAddrs[2], kavaAddrs[2], "", 0, types.Open)
	s4 := types.NewAtomicSwap(coinsSingle, randomNumberHashes[3], int64(500), timestamps[3], binanceAddrs[3], kavaAddrs[3], "", 0, types.Open)
	atomicSwaps = append(atomicSwaps, s1, s2, s3, s4)
	return atomicSwaps
}
