package bep3_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     app.TestApp
	handler sdk.Handler
	keeper  bep3.Keeper
	addrs   []sdk.AccAddress
}

func (suite *HandlerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetBep3Keeper()

	// Set up genesis state and initialize
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := []sdk.Coins{}
	for j := 0; j < 3; j++ {
		coins = append(coins, cs(c("bnb", 10000000000), c("ukava", 10000000000)))
	}
	authGS := app.NewAuthGenState(addrs, coins)
	tApp.InitializeFromGenesisStates(authGS, NewBep3GenStateMulti())

	suite.addrs = addrs
	suite.handler = bep3.NewHandler(keeper)
	suite.keeper = keeper
	suite.app = tApp
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) AddAtomicSwap() (cmn.HexBytes, cmn.HexBytes) {
	expireHeight := int64(360)
	amount := cs(c("bnb", int64(50000)))
	timestamp := ts(0)
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	// Create atomic swap and check err to confirm creation
	err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireHeight,
		suite.addrs[0], suite.addrs[1], TestSenderOtherChain, TestRecipientOtherChain,
		amount, amount.String())
	suite.Nil(err)

	swapID := types.CalculateSwapID(randomNumberHash, suite.addrs[0], TestSenderOtherChain)
	return swapID, randomNumber.Bytes()
}

func (suite *HandlerTestSuite) TestMsgCreateAtomicSwap() {
	amount := cs(c("bnb", int64(10000)))
	timestamp := ts(0)
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	msg := types.NewMsgCreateAtomicSwap(
		suite.addrs[0], suite.addrs[2], TestRecipientOtherChain, TestSenderOtherChain,
		randomNumberHash, timestamp, amount, amount.String(), int64(300), true)

	res := suite.handler(suite.ctx, msg)
	suite.True(res.IsOK())
}

func (suite *HandlerTestSuite) TestMsgClaimAtomicSwap() {
	// Attempt claim msg on fake atomic swap
	badRandomNumber, _ := types.GenerateSecureRandomNumber()
	badRandomNumberHash := types.CalculateRandomHash(badRandomNumber.Bytes(), ts(0))
	badSwapID := types.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := types.NewMsgClaimAtomicSwap(suite.addrs[0], badSwapID, badRandomNumber.Bytes())
	badRes := suite.handler(suite.ctx, badMsg)
	suite.False(badRes.IsOK())
	suite.True(strings.Contains(badRes.Log, fmt.Sprintf("AtomicSwap %s was not found", hex.EncodeToString(badSwapID))))

	// Add an atomic swap before attempting new claim msg
	swapID, randomNumber := suite.AddAtomicSwap()
	msg := types.NewMsgClaimAtomicSwap(suite.addrs[0], swapID, randomNumber)
	res := suite.handler(suite.ctx, msg)
	suite.True(res.IsOK())
}

func (suite *HandlerTestSuite) TestMsgRefundAtomicSwap() {
	// Attempt refund msg on fake atomic swap
	badRandomNumber, _ := types.GenerateSecureRandomNumber()
	badRandomNumberHash := types.CalculateRandomHash(badRandomNumber.Bytes(), ts(0))
	badSwapID := types.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := types.NewMsgRefundAtomicSwap(suite.addrs[0], badSwapID)
	badRes := suite.handler(suite.ctx, badMsg)
	suite.False(badRes.IsOK())
	suite.True(strings.Contains(badRes.Log, fmt.Sprintf("AtomicSwap %s was not found", hex.EncodeToString(badSwapID))))

	// Add an atomic swap and build refund msg
	swapID, _ := suite.AddAtomicSwap()
	msg := types.NewMsgRefundAtomicSwap(suite.addrs[0], swapID)

	// Attempt to refund active atomic swap
	res1 := suite.handler(suite.ctx, msg)
	suite.True(strings.Contains(res1.Log, "atomic swap is still active and cannot be refunded"))
	suite.False(res1.IsOK())

	// Expire the atomic swap with begin blocker and attempt refund
	laterCtx := suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400)
	bep3.BeginBlocker(laterCtx, suite.keeper)
	res2 := suite.handler(laterCtx, msg)
	suite.True(res2.IsOK())
}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res := suite.handler(suite.ctx, sdk.NewTestMsg())
	suite.False(res.IsOK())
	suite.True(strings.Contains(res.Log, "unrecognized bep3 message type"))
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
