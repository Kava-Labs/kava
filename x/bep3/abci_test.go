package bep3_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type ABCITestSuite struct {
	suite.Suite
	keeper        keeper.Keeper
	app           app.TestApp
	ctx           sdk.Context
	addrs         []sdk.AccAddress
	swapIDs       []cmn.HexBytes
	randomNumbers []cmn.HexBytes
}

func (suite *ABCITestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	// Set up auth GenesisState
	_, addrs := app.GeneratePrivKeyAddressPairs(11)
	coins := []sdk.Coins{}
	for j := 0; j < 11; j++ {
		coins = append(coins, cs(c("bnb", 10000000000), c("ukava", 10000000000)))
	}
	authGS := app.NewAuthGenState(addrs, coins)
	// Initialize test app
	tApp.InitializeFromGenesisStates(authGS, NewBep3GenStateMulti())

	suite.ctx = ctx
	suite.app = tApp
	suite.addrs = addrs
	suite.ResetKeeper()
}

func (suite *ABCITestSuite) ResetKeeper() {
	suite.keeper = suite.app.GetBep3Keeper()

	var swapIDs []cmn.HexBytes
	var randomNumbers []cmn.HexBytes
	for i := 0; i < 10; i++ {
		// Set up atomic swap variables
		expireHeight := int64(360)
		amount := cs(c("bnb", int64(50000+i*100)))
		timestamp := ts(0)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

		// Create atomic swap and check err to confirm creation
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireHeight,
			suite.addrs[i], suite.addrs[i], TestSenderOtherChain, TestRecipientOtherChain,
			amount, amount.String())
		suite.Nil(err)

		// Store swap's calculated ID and secret random number
		swapID := types.CalculateSwapID(randomNumberHash, suite.addrs[i], TestSenderOtherChain)
		suite.swapIDs = append(swapIDs, swapID)
		suite.randomNumbers = append(randomNumbers, randomNumber.Bytes())
	}
}

func (suite *ABCITestSuite) TestBeginBlocker() {
	testCases := []struct {
		name            string
		firstCtx        sdk.Context
		secondCtx       sdk.Context
		expectedStatus  types.SwapStatus
		expectInStorage bool
	}{
		{
			name:            "normal",
			firstCtx:        suite.ctx,
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 10),
			expectedStatus:  types.Open,
			expectInStorage: true,
		},
		{
			name:            "after expiration",
			firstCtx:        suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400),
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 410),
			expectedStatus:  types.Expired,
			expectInStorage: true,
		},
		{
			name:            "after completion",
			firstCtx:        suite.ctx,
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 10),
			expectedStatus:  types.Completed,
			expectInStorage: true,
		},
		{
			name:            "after deletion",
			firstCtx:        suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400),
			secondCtx:       suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400 + SwapLongtermStorageDuration),
			expectedStatus:  types.NULL,
			expectInStorage: false,
		},
	}

	for _, tc := range testCases {
		// Reset keeper and run the initial begin blocker
		suite.ResetKeeper()
		bep3.BeginBlocker(tc.firstCtx, suite.keeper)

		switch tc.expectedStatus {
		case types.Completed:
			for i, swapID := range suite.swapIDs {
				err := suite.keeper.ClaimAtomicSwap(tc.firstCtx, suite.addrs[10], swapID, suite.randomNumbers[i])
				suite.Nil(err)
			}
		case types.NULL:
			for _, swapID := range suite.swapIDs {
				err := suite.keeper.RefundAtomicSwap(tc.firstCtx, suite.addrs[10], swapID)
				suite.Nil(err)
			}
		}

		// Run the second begin blocker
		bep3.BeginBlocker(tc.secondCtx, suite.keeper)

		// Check each swap's availibility and status
		for _, swapID := range suite.swapIDs {
			storedSwap, found := suite.keeper.GetAtomicSwap(tc.secondCtx, swapID)
			if tc.expectInStorage {
				suite.True(found)
			} else {
				suite.False(found)
			}
			suite.Equal(tc.expectedStatus, storedSwap.Status)
		}
	}
}

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}
