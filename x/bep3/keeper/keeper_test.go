package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const LongtermStorageDuration = 86400

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	suite.ResetChain()
	return
}

func (suite *KeeperTestSuite) ResetChain() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetBep3Keeper()

	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
}

func (suite *KeeperTestSuite) TestGetSetAtomicSwap() {
	suite.ResetChain()

	// Set new atomic swap
	atomicSwap := atomicSwap(suite.ctx, 1)
	suite.keeper.SetAtomicSwap(suite.ctx, atomicSwap)

	// Check atomic swap in store
	s, found := suite.keeper.GetAtomicSwap(suite.ctx, atomicSwap.GetSwapID())
	suite.True(found)
	suite.Equal(atomicSwap, s)

	// Check fake atomic swap not in store
	fakeSwapID := types.CalculateSwapID(atomicSwap.RandomNumberHash, kavaAddrs[1], "otheraddress")
	_, found = suite.keeper.GetAtomicSwap(suite.ctx, fakeSwapID)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestRemoveAtomicSwap() {
	suite.ResetChain()

	// Set new atomic swap
	atomicSwap := atomicSwap(suite.ctx, 1)
	suite.keeper.SetAtomicSwap(suite.ctx, atomicSwap)

	// Check atomic swap in store
	s, found := suite.keeper.GetAtomicSwap(suite.ctx, atomicSwap.GetSwapID())
	suite.True(found)
	suite.Equal(atomicSwap, s)

	suite.keeper.RemoveAtomicSwap(suite.ctx, atomicSwap.GetSwapID())

	// Check atomic swap not in store
	_, found = suite.keeper.GetAtomicSwap(suite.ctx, atomicSwap.GetSwapID())
	suite.False(found)
}
func (suite *KeeperTestSuite) TestIterateAtomicSwaps() {
	suite.ResetChain()

	// Set atomic swaps
	atomicSwaps := atomicSwaps(suite.ctx, 4)
	for _, s := range atomicSwaps {
		suite.keeper.SetAtomicSwap(suite.ctx, s)
	}

	// Read each atomic swap from the store
	var readAtomicSwaps types.AtomicSwaps
	suite.keeper.IterateAtomicSwaps(suite.ctx, func(a types.AtomicSwap) bool {
		readAtomicSwaps = append(readAtomicSwaps, a)
		return false
	})

	// Check expected values
	suite.Equal(len(atomicSwaps), len(readAtomicSwaps))
}

func (suite *KeeperTestSuite) TestGetAllAtomicSwaps() {
	suite.ResetChain()

	// Set atomic swaps
	atomicSwaps := atomicSwaps(suite.ctx, 4)
	for _, s := range atomicSwaps {
		suite.keeper.SetAtomicSwap(suite.ctx, s)
	}

	// Get and check atomic swaps
	res := suite.keeper.GetAllAtomicSwaps(suite.ctx)
	suite.Equal(4, len(res))
}

func (suite *KeeperTestSuite) TestInsertIntoByBlockIndex() {
	suite.ResetChain()

	// Set new atomic swap in by block index
	atomicSwap := atomicSwap(suite.ctx, 1)
	suite.keeper.InsertIntoByBlockIndex(suite.ctx, atomicSwap)

	// Block index lacks getter methods, must use iteration to get count of swaps in store
	var swapIDs [][]byte
	suite.keeper.IterateAtomicSwapsByBlock(suite.ctx, uint64(atomicSwap.ExpireHeight+1), func(id []byte) bool {
		swapIDs = append(swapIDs, id)
		return false
	})
	suite.Equal(len(swapIDs), 1)

	// Marshal the expected swapID
	cdc := suite.app.Codec()
	res, _ := cdc.MarshalBinaryBare(atomicSwap.GetSwapID())
	expectedSwapID := res[1:]

	suite.Equal(expectedSwapID, swapIDs[0])
}

func (suite *KeeperTestSuite) TestRemoveFromByBlockIndex() {
	suite.ResetChain()

	// Set new atomic swap in by block index
	atomicSwap := atomicSwap(suite.ctx, 1)
	suite.keeper.InsertIntoByBlockIndex(suite.ctx, atomicSwap)

	// Check stored data in block index
	var swapIDsPre [][]byte
	suite.keeper.IterateAtomicSwapsByBlock(suite.ctx, uint64(atomicSwap.ExpireHeight+1), func(id []byte) bool {
		swapIDsPre = append(swapIDsPre, id)
		return false
	})
	suite.Equal(len(swapIDsPre), 1)

	suite.keeper.RemoveFromByBlockIndex(suite.ctx, atomicSwap)

	// Check stored data not in block index
	var swapIDsPost [][]byte
	suite.keeper.IterateAtomicSwapsByBlock(suite.ctx, uint64(atomicSwap.ExpireHeight+1), func(id []byte) bool {
		swapIDsPost = append(swapIDsPost, id)
		return false
	})
	suite.Equal(len(swapIDsPost), 0)
}

func (suite *KeeperTestSuite) TestIterateAtomicSwapsByBlock() {
	suite.ResetChain()

	type args struct {
		blockCtx sdk.Context
		swap     types.AtomicSwap
	}

	var testCases []args
	for i := 0; i < 8; i++ {
		// Set up context 100 blocks apart
		blockCtx := suite.ctx.WithBlockHeight(int64(i) * 100)

		// Initialize a new atomic swap (different randomNumberHash = different swap IDs)
		timestamp := tmtime.Now().Add(time.Duration(i) * time.Minute).Unix()
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

		atomicSwap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
			blockCtx.BlockHeight(), timestamp, kavaAddrs[0], kavaAddrs[1],
			binanceAddrs[0].String(), binanceAddrs[1].String(), 0, types.Open, true)

		// Insert into block index
		suite.keeper.InsertIntoByBlockIndex(blockCtx, atomicSwap)
		// Add to local block index
		testCases = append(testCases, args{blockCtx, atomicSwap})
	}

	// Set up the expected swap IDs for a given cutoff block
	cutoffBlock := int64(450)
	var expectedSwapIDs [][]byte
	for _, tc := range testCases {
		if tc.blockCtx.BlockHeight() < cutoffBlock || tc.blockCtx.BlockHeight() == cutoffBlock {
			expectedSwapIDs = append(expectedSwapIDs, tc.swap.GetSwapID())
		}
	}

	// Read the swap IDs from store for a given cutoff block
	var readSwapIDs [][]byte
	suite.keeper.IterateAtomicSwapsByBlock(suite.ctx, uint64(cutoffBlock), func(id []byte) bool {
		readSwapIDs = append(readSwapIDs, id)
		return false
	})

	suite.Equal(expectedSwapIDs, readSwapIDs)
}

func (suite *KeeperTestSuite) TestInsertIntoLongtermStorage() {
	suite.ResetChain()

	// Set atomic swap in longterm storage
	atomicSwap := atomicSwap(suite.ctx, 1)
	atomicSwap.ClosedBlock = suite.ctx.BlockHeight()
	suite.keeper.InsertIntoLongtermStorage(suite.ctx, atomicSwap)

	// Longterm storage lacks getter methods, must use iteration to get count of swaps in store
	var swapIDs [][]byte
	suite.keeper.IterateAtomicSwapsLongtermStorage(suite.ctx, uint64(atomicSwap.ClosedBlock+LongtermStorageDuration), func(id []byte) bool {
		swapIDs = append(swapIDs, id)
		return false
	})
	suite.Equal(len(swapIDs), 1)

	// Marshal the expected swapID
	cdc := suite.app.Codec()
	res, _ := cdc.MarshalBinaryBare(atomicSwap.GetSwapID())
	expectedSwapID := res[1:]

	suite.Equal(expectedSwapID, swapIDs[0])
}

func (suite *KeeperTestSuite) TestRemoveFromLongtermStorage() {
	suite.ResetChain()

	// Set atomic swap in longterm storage
	atomicSwap := atomicSwap(suite.ctx, 1)
	atomicSwap.ClosedBlock = suite.ctx.BlockHeight()
	suite.keeper.InsertIntoLongtermStorage(suite.ctx, atomicSwap)

	// Longterm storage lacks getter methods, must use iteration to get count of swaps in store
	var swapIDs [][]byte
	suite.keeper.IterateAtomicSwapsLongtermStorage(suite.ctx, uint64(atomicSwap.ClosedBlock+LongtermStorageDuration), func(id []byte) bool {
		swapIDs = append(swapIDs, id)
		return false
	})
	suite.Equal(len(swapIDs), 1)

	suite.keeper.RemoveFromLongtermStorage(suite.ctx, atomicSwap)

	// Check stored data not in block index
	var swapIDsPost [][]byte
	suite.keeper.IterateAtomicSwapsLongtermStorage(suite.ctx, uint64(atomicSwap.ClosedBlock+LongtermStorageDuration), func(id []byte) bool {
		swapIDsPost = append(swapIDsPost, id)
		return false
	})
	suite.Equal(len(swapIDsPost), 0)
}

func (suite *KeeperTestSuite) TestIterateAtomicSwapsLongtermStorage() {
	suite.ResetChain()

	// Set up atomic swaps with stagged closed blocks
	var swaps types.AtomicSwaps
	for i := 0; i < 8; i++ {
		timestamp := tmtime.Now().Unix()
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

		atomicSwap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
			suite.ctx.BlockHeight(), timestamp, kavaAddrs[0], kavaAddrs[1],
			binanceAddrs[0].String(), binanceAddrs[1].String(), 100, types.Open, true)

		// Set closed block staggered by 100 blocks and insert into longterm storage
		atomicSwap.ClosedBlock = int64(i) * 100
		suite.keeper.InsertIntoLongtermStorage(suite.ctx, atomicSwap)
		// Add to local longterm storage
		swaps = append(swaps, atomicSwap)
	}

	// Set up the expected swap IDs for a given cutoff block.
	cutoffBlock := int64(LongtermStorageDuration + 350)
	var expectedSwapIDs [][]byte
	for _, swap := range swaps {
		if swap.ClosedBlock+LongtermStorageDuration < cutoffBlock ||
			swap.ClosedBlock+LongtermStorageDuration == cutoffBlock {
			expectedSwapIDs = append(expectedSwapIDs, swap.GetSwapID())
		}
	}

	// Read the swap IDs from store for a given cutoff block
	var readSwapIDs [][]byte
	suite.keeper.IterateAtomicSwapsLongtermStorage(suite.ctx, uint64(cutoffBlock), func(id []byte) bool {
		readSwapIDs = append(readSwapIDs, id)
		return false
	})

	// At the cutoff block, iteration should return half of the swap IDs
	suite.Equal(len(swaps)/2, len(expectedSwapIDs))
	suite.Equal(len(swaps)/2, len(readSwapIDs))
	// Should be the same IDs
	suite.Equal(expectedSwapIDs, readSwapIDs)
}

func (suite *KeeperTestSuite) TestGetSetAssetSupply() {
	suite.ResetChain()

	// Set new asset supply
	asset := c("bnb", 50000)
	suite.keeper.SetAssetSupply(suite.ctx, asset, []byte(asset.Denom))

	// Check asset in store
	assetSupply, found := suite.keeper.GetAssetSupply(suite.ctx, []byte(asset.Denom))
	suite.True(found)
	suite.Equal(asset, assetSupply)

	// Check fake asset not in store
	fakeAsset := c("xyz", 50000)
	_, found = suite.keeper.GetAssetSupply(suite.ctx, []byte(fakeAsset.Denom))
	suite.False(found)
}

func (suite *KeeperTestSuite) TestIterateAssetSupplies() {
	suite.ResetChain()

	// Set asset supplies
	assetSupplies := []sdk.Coin{c("test1", 25000), c("test2", 50000), c("test3", 100000)}
	for _, asset := range assetSupplies {
		suite.keeper.SetAssetSupply(suite.ctx, asset, []byte(asset.Denom))
	}

	// Read each asset supply from the store
	var readAssetSupplies []sdk.Coin
	suite.keeper.IterateAssetSupplies(suite.ctx, func(c sdk.Coin) bool {
		readAssetSupplies = append(readAssetSupplies, c)
		return false
	})

	// Check expected values
	suite.Equal(assetSupplies, readAssetSupplies)
}

func (suite *KeeperTestSuite) TestGetAllAssetSupplies() {
	suite.ResetChain()

	// Set asset supplies
	assetSupplies := []sdk.Coin{c("test1", 25000), c("test2", 50000), c("test3", 100000)}
	for _, asset := range assetSupplies {
		suite.keeper.SetAssetSupply(suite.ctx, asset, []byte(asset.Denom))
	}

	// Get all asset supplies
	res := suite.keeper.GetAllAssetSupplies(suite.ctx)
	suite.Equal(3, len(res))
}

func (suite *KeeperTestSuite) TestGetSetInSwapSupply() {
	suite.ResetChain()

	// Set new asset supply
	asset := c("bnb", 50000)
	suite.keeper.SetInSwapSupply(suite.ctx, asset, []byte(asset.Denom))

	// Check in swap supply in store
	inSwapSupply, found := suite.keeper.GetInSwapSupply(suite.ctx, []byte(asset.Denom))
	suite.True(found)
	suite.Equal(asset, inSwapSupply)

	// Check fake asset not in store
	fakeAsset := c("xyz", 50000)
	_, found = suite.keeper.GetInSwapSupply(suite.ctx, []byte(fakeAsset.Denom))
	suite.False(found)
}

func (suite *KeeperTestSuite) TestIterateInSwapSupplies() {
	suite.ResetChain()

	// Set in swap supplies
	inSwapSupplies := []sdk.Coin{c("test1", 25000), c("test2", 50000), c("test3", 100000)}
	for _, asset := range inSwapSupplies {
		suite.keeper.SetInSwapSupply(suite.ctx, asset, []byte(asset.Denom))
	}

	// Read each in swap supply from the store
	var readInSwapSupplies []sdk.Coin
	suite.keeper.IterateInSwapSupplies(suite.ctx, func(c sdk.Coin) bool {
		readInSwapSupplies = append(readInSwapSupplies, c)
		return false
	})

	// Check expected values
	suite.Equal(inSwapSupplies, readInSwapSupplies)
}

func (suite *KeeperTestSuite) TestGetAllInSwapSupplies() {
	suite.ResetChain()

	// Set in swap supplies
	inSwapSupplies := []sdk.Coin{c("test1", 25000), c("test2", 50000), c("test3", 100000)}
	for _, asset := range inSwapSupplies {
		suite.keeper.SetInSwapSupply(suite.ctx, asset, []byte(asset.Denom))
	}

	// Get all in swap supplies
	res := suite.keeper.GetAllInSwapSupplies(suite.ctx)
	suite.Equal(3, len(res))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
