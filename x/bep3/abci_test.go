package bep3_test

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	TestSenderOtherChain    = "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7"
	TestRecipientOtherChain = "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7"
)

type ABCITestSuite struct {
	suite.Suite
	keeper   keeper.Keeper
	app      app.TestApp
	ctx      sdk.Context
	querier  sdk.Querier
	addrs    []sdk.AccAddress
	swapIDs  []cmn.HexBytes
	isSwapID map[string]bool
}

func (suite *ABCITestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	// Set up auth GenesisState
	_, addrs := app.GeneratePrivKeyAddressPairs(10)
	coins := []sdk.Coins{}
	for j := 0; j < 10; j++ {
		coins = append(coins, cs(c("bnb", 10000000000), c("ukava", 10000000000)))
	}
	authGS := app.NewAuthGenState(addrs, coins)

	tApp.InitializeFromGenesisStates(
		authGS,
		NewBep3GenStateMulti(),
	)

	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = tApp.GetBep3Keeper()
	suite.querier = keeper.NewQuerier(suite.keeper)
	suite.addrs = addrs

	// Create atomic swaps and save IDs
	var swapIDs []cmn.HexBytes
	isSwapID := make(map[string]bool)
	for i := 0; i < 10; i++ {
		// Set up atomic swap variables
		expireHeight := int64(360)
		amount := cs(c("bnb", int64(50000+i*100)))
		timestamp := ts(0)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

		// Create atomic swap and check err
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireHeight,
			suite.addrs[i], suite.addrs[i], TestSenderOtherChain, TestRecipientOtherChain,
			amount, amount.String())
		suite.Nil(err)

		// Calculate swap ID and save
		swapID := types.CalculateSwapID(randomNumberHash, suite.addrs[i], TestSenderOtherChain)
		swapIDs = append(swapIDs, swapID)
		isSwapID[hex.EncodeToString(swapID)] = true
	}
	suite.swapIDs = swapIDs
	suite.isSwapID = isSwapID
}

func (suite *ABCITestSuite) TestBeginBlocker() {
	// 1. Attempt to refund an atomic swap - should be rejected
	// 2. Claim an atomic swap - should be accepted and closed
	// 3. Move block time past expiration - all swaps should be closed
	// 4. Attempt to refund an atomic swap - should be accepted
	// 5. Move block time past deletion time - all swaps should be deleted

	// TODO:
	// Run the endblocker, simulating a block time 1ns before swap expiry
	// Check swap has not been closed yet
	// Run the endblocker, simulating a block time equal to swap expiry
	// Check swap has been deleted
}

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}
