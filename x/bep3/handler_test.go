package bep3_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
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
	tApp.InitializeFromGenesisStates(authGS, NewBep3GenStateMulti(addrs[0]))

	suite.addrs = addrs
	suite.handler = bep3.NewHandler(keeper)
	suite.keeper = keeper
	suite.app = tApp
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) AddAtomicSwap() (tmbytes.HexBytes, tmbytes.HexBytes) {
	expireHeight := int64(360)
	amount := cs(c("bnb", int64(50000)))
	timestamp := ts(0)
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	// Create atomic swap and check err to confirm creation
	err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireHeight,
		suite.addrs[0], suite.addrs[1], TestSenderOtherChain, TestRecipientOtherChain,
		amount, amount.String(), true)
	suite.Nil(err)

	swapID := bep3.CalculateSwapID(randomNumberHash, suite.addrs[0], TestSenderOtherChain)
	return swapID, randomNumber.Bytes()
}

func (suite *HandlerTestSuite) TestMsgCreateAtomicSwap() {
	amount := cs(c("bnb", int64(10000)))
	timestamp := ts(0)
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	msg := bep3.NewMsgCreateAtomicSwap(
		suite.addrs[0], suite.addrs[2], TestRecipientOtherChain, TestSenderOtherChain,
		randomNumberHash, timestamp, amount, amount.String(), int64(300), true)

	res, err := suite.handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *HandlerTestSuite) TestMsgClaimAtomicSwap() {
	// Attempt claim msg on fake atomic swap
	badRandomNumber, _ := bep3.GenerateSecureRandomNumber()
	badRandomNumberHash := bep3.CalculateRandomHash(badRandomNumber.Bytes(), ts(0))
	badSwapID := bep3.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := bep3.NewMsgClaimAtomicSwap(suite.addrs[0], badSwapID, badRandomNumber.Bytes())
	badRes, err := suite.handler(suite.ctx, badMsg)
	suite.Require().Error(err)
	suite.True(strings.Contains(badRes.Log, fmt.Sprintf("AtomicSwap %s was not found", hex.EncodeToString(badSwapID))))

	// Add an atomic swap before attempting new claim msg
	swapID, randomNumber := suite.AddAtomicSwap()
	msg := bep3.NewMsgClaimAtomicSwap(suite.addrs[0], swapID, randomNumber)
	res, err := suite.handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *HandlerTestSuite) TestMsgRefundAtomicSwap() {
	// Attempt refund msg on fake atomic swap
	badRandomNumber, _ := bep3.GenerateSecureRandomNumber()
	badRandomNumberHash := bep3.CalculateRandomHash(badRandomNumber.Bytes(), ts(0))
	badSwapID := bep3.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := bep3.NewMsgRefundAtomicSwap(suite.addrs[0], badSwapID)
	badRes, err := suite.handler(suite.ctx, badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap and build refund msg
	swapID, _ := suite.AddAtomicSwap()
	msg := bep3.NewMsgRefundAtomicSwap(suite.addrs[0], swapID)

	// Attempt to refund active atomic swap
	res1, err := suite.handler(suite.ctx, msg)
	suite.True(strings.Contains(res1.Log, "atomic swap is still active and cannot be refunded"))
	suite.Require().Error(err)

	// Expire the atomic swap with begin blocker and attempt refund
	laterCtx := suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400)
	bep3.BeginBlocker(laterCtx, suite.keeper)
	res2, err := suite.handler(laterCtx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res2)
}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, sdk.NewTestMsg())
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
