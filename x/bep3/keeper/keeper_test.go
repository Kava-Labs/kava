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
	swapID, err := types.CalculateSwapID(randomNumberHashes[0], binanceAddrs[0], "")
	suite.NoError(err)

	heightSpan := int64(1000)
	expirationBlock := uint64(suite.ctx.BlockHeight()) + uint64(heightSpan)
	atomicSwap := types.NewAtomicSwap(swapID, binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHashes[0], timestamps[0], coinsSingle, "50000bnb", false, expirationBlock)
	suite.keeper.SetAtomicSwap(suite.ctx, atomicSwap)

	s, found := suite.keeper.GetAtomicSwap(suite.ctx, swapID)
	suite.True(found)
	suite.Equal(atomicSwap, s)

	fakeSwapID, err := types.CalculateSwapID(atomicSwap.RandomNumberHash, kavaAddrs[1], "otheraddress")
	suite.NoError(err)
	_, found = suite.keeper.GetAtomicSwap(suite.ctx, fakeSwapID)
	suite.False(found)

	suite.keeper.DeleteAtomicSwap(suite.ctx, swapID)
	_, found = suite.keeper.GetAtomicSwap(suite.ctx, swapID)
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
		swapID, _ := types.CalculateSwapID(randomNumberHashes[i], binanceAddrs[i], "")
		swapIDs = append(swapIDs, swapID)
	}
	s1 := types.NewAtomicSwap(swapIDs[0], binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHashes[0], timestamps[0], coinsSingle, "50000bnb", false, uint64(50500+1000))
	s2 := types.NewAtomicSwap(swapIDs[1], binanceAddrs[1], kavaAddrs[1], "", "", randomNumberHashes[1], timestamps[1], coinsSingle, "50000bnb", false, uint64(61500+1000))
	s3 := types.NewAtomicSwap(swapIDs[2], binanceAddrs[2], kavaAddrs[2], "", "", randomNumberHashes[2], timestamps[2], coinsSingle, "50000bnb", false, uint64(72500+1000))
	s4 := types.NewAtomicSwap(swapIDs[3], binanceAddrs[3], kavaAddrs[3], "", "", randomNumberHashes[3], timestamps[3], coinsSingle, "50000bnb", false, uint64(83500+1000))
	atomicSwaps = append(atomicSwaps, s1, s2, s3, s4)
	return atomicSwaps
}
