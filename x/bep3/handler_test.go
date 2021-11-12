package bep3_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	testdata "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     app.TestApp
	handler sdk.Handler
	keeper  keeper.Keeper
	addrs   []sdk.AccAddress
}

func (suite *HandlerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetBep3Keeper()

	cdc := tApp.AppCodec()

	// Set up genesis state and initialize
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := sdk.NewCoins(c("bnb", 10000000000), c("ukava", 10000000000))
	authGS := app.NewFundedGenStateWithSameCoins(tApp.AppCodec(), coins, addrs)
	tApp.InitializeFromGenesisStates(authGS, NewBep3GenStateMulti(cdc, addrs[0]))

	suite.addrs = addrs
	suite.handler = bep3.NewHandler(keeper)
	suite.keeper = keeper
	suite.app = tApp
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) AddAtomicSwap() (tmbytes.HexBytes, tmbytes.HexBytes) {
	expireHeight := types.DefaultMinBlockLock
	amount := cs(c("bnb", int64(50000)))
	timestamp := ts(0)
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

	// Create atomic swap and check err to confirm creation
	err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireHeight,
		suite.addrs[0], suite.addrs[1], TestSenderOtherChain, TestRecipientOtherChain,
		amount, true)
	suite.Nil(err)

	swapID := types.CalculateSwapID(randomNumberHash, suite.addrs[0], TestSenderOtherChain)
	return swapID, randomNumber[:]
}

func (suite *HandlerTestSuite) TestMsgCreateAtomicSwap() {
	amount := cs(c("bnb", int64(10000)))
	timestamp := ts(0)
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

	msg := types.NewMsgCreateAtomicSwap(
		suite.addrs[0].String(), suite.addrs[2].String(), TestRecipientOtherChain,
		TestSenderOtherChain, randomNumberHash, timestamp, amount,
		types.DefaultMinBlockLock)

	res, err := suite.handler(suite.ctx, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *HandlerTestSuite) TestMsgClaimAtomicSwap() {
	// Attempt claim msg on fake atomic swap
	badRandomNumber, _ := types.GenerateSecureRandomNumber()
	badRandomNumberHash := types.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := types.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := types.NewMsgClaimAtomicSwap(suite.addrs[0].String(), badSwapID, badRandomNumber[:])
	badRes, err := suite.handler(suite.ctx, &badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap before attempting new claim msg
	swapID, randomNumber := suite.AddAtomicSwap()
	msg := types.NewMsgClaimAtomicSwap(suite.addrs[0].String(), swapID, randomNumber)
	res, err := suite.handler(suite.ctx, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *HandlerTestSuite) TestMsgRefundAtomicSwap() {
	// Attempt refund msg on fake atomic swap
	badRandomNumber, _ := types.GenerateSecureRandomNumber()
	badRandomNumberHash := types.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := types.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := types.NewMsgRefundAtomicSwap(suite.addrs[0].String(), badSwapID)
	badRes, err := suite.handler(suite.ctx, &badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap and build refund msg
	swapID, _ := suite.AddAtomicSwap()
	msg := types.NewMsgRefundAtomicSwap(suite.addrs[0].String(), swapID)

	// Attempt to refund active atomic swap
	res1, err := suite.handler(suite.ctx, &msg)
	suite.Require().Error(err)
	suite.Require().Nil(res1)

	// Expire the atomic swap with begin blocker and attempt refund
	laterCtx := suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400)
	bep3.BeginBlocker(laterCtx, suite.keeper)
	res2, err := suite.handler(laterCtx, &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res2)
}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, testdata.NewTestMsg())
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
