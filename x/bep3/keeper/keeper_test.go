package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
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
	fakeSwapID := types.CalculateSwapID(atomicSwap.RandomNumberHash, TestUser2, "otheraddress")
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
	suite.keeper.IterateAtomicSwapsByBlock(suite.ctx, atomicSwap.ExpireHeight+1, func(id []byte) bool {
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
	suite.keeper.IterateAtomicSwapsByBlock(suite.ctx, atomicSwap.ExpireHeight+1, func(id []byte) bool {
		swapIDsPre = append(swapIDsPre, id)
		return false
	})
	suite.Equal(len(swapIDsPre), 1)

	suite.keeper.RemoveFromByBlockIndex(suite.ctx, atomicSwap)

	// Check stored data not in block index
	var swapIDsPost [][]byte
	suite.keeper.IterateAtomicSwapsByBlock(suite.ctx, atomicSwap.ExpireHeight+1, func(id []byte) bool {
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
		randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

		atomicSwap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
			uint64(blockCtx.BlockHeight()), timestamp, TestUser1, TestUser2,
			TestSenderOtherChain, TestRecipientOtherChain, 0, types.Open,
			true, types.Incoming)

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
		randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

		atomicSwap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
			uint64(suite.ctx.BlockHeight()), timestamp, TestUser1, TestUser2,
			TestSenderOtherChain, TestRecipientOtherChain, 100, types.Open,
			true, types.Incoming)

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
	denom := "bnb"
	// Put asset supply in store
	assetSupply := types.NewAssetSupply(c(denom, 0), c(denom, 0), c(denom, 50000))
	suite.keeper.SetAssetSupply(suite.ctx, assetSupply, denom)

	// Check asset in store
	storedAssetSupply, found := suite.keeper.GetAssetSupply(suite.ctx, denom)
	suite.True(found)
	suite.Equal(assetSupply, storedAssetSupply)

	// Check fake asset supply not in store
	fakeDenom := "xyz"
	_, found = suite.keeper.GetAssetSupply(suite.ctx, fakeDenom)
	suite.False(found)
}

func (suite *KeeperTestSuite) TestGetAllAssetSupplies() {

	// Put asset supply in store
	assetSupply := types.NewAssetSupply(c("bnb", 0), c("bnb", 0), c("bnb", 50000))
	suite.keeper.SetAssetSupply(suite.ctx, assetSupply, "bnb")
	assetSupply = types.NewAssetSupply(c("inc", 0), c("inc", 0), c("inc", 50000))
	suite.keeper.SetAssetSupply(suite.ctx, assetSupply, "inc")

	supplies := suite.keeper.GetAllAssetSupplies(suite.ctx)
	suite.Equal(2, len(supplies))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
