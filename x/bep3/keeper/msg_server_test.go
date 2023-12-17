package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	app       app.TestApp
	msgServer types.MsgServer
	keeper    keeper.Keeper
	addrs     []sdk.AccAddress
}

func (suite *MsgServerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	cdc := tApp.AppCodec()

	// Set up genesis state and initialize
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := sdk.NewCoins(c("bnb", 10000000000), c("ukava", 10000000000))
	authGS := app.NewFundedGenStateWithSameCoins(tApp.AppCodec(), coins, addrs)
	tApp.InitializeFromGenesisStates(authGS, NewBep3GenStateMulti(cdc, addrs[0]))

	suite.addrs = addrs
	suite.keeper = tApp.GetBep3Keeper()
	suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)
	suite.app = tApp
	suite.ctx = ctx
}

func (suite *MsgServerTestSuite) AddAtomicSwap() (tmbytes.HexBytes, tmbytes.HexBytes) {
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

func (suite *MsgServerTestSuite) TestMsgCreateAtomicSwap() {
	amount := cs(c("bnb", int64(10000)))
	timestamp := ts(0)
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

	msg := types.NewMsgCreateAtomicSwap(
		suite.addrs[0].String(), suite.addrs[2].String(), TestRecipientOtherChain,
		TestSenderOtherChain, randomNumberHash, timestamp, amount,
		types.DefaultMinBlockLock)

	res, err := suite.msgServer.CreateAtomicSwap(sdk.WrapSDKContext(suite.ctx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *MsgServerTestSuite) TestMsgClaimAtomicSwap() {
	// Attempt claim msg on fake atomic swap
	badRandomNumber, _ := types.GenerateSecureRandomNumber()
	badRandomNumberHash := types.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := types.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := types.NewMsgClaimAtomicSwap(suite.addrs[0].String(), badSwapID, badRandomNumber[:])
	badRes, err := suite.msgServer.ClaimAtomicSwap(sdk.WrapSDKContext(suite.ctx), &badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap before attempting new claim msg
	swapID, randomNumber := suite.AddAtomicSwap()
	msg := types.NewMsgClaimAtomicSwap(suite.addrs[0].String(), swapID, randomNumber)
	res, err := suite.msgServer.ClaimAtomicSwap(sdk.WrapSDKContext(suite.ctx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
}

func (suite *MsgServerTestSuite) TestMsgRefundAtomicSwap() {
	// Attempt refund msg on fake atomic swap
	badRandomNumber, _ := types.GenerateSecureRandomNumber()
	badRandomNumberHash := types.CalculateRandomHash(badRandomNumber[:], ts(0))
	badSwapID := types.CalculateSwapID(badRandomNumberHash, suite.addrs[0], TestSenderOtherChain)
	badMsg := types.NewMsgRefundAtomicSwap(suite.addrs[0].String(), badSwapID)
	badRes, err := suite.msgServer.RefundAtomicSwap(sdk.WrapSDKContext(suite.ctx), &badMsg)
	suite.Require().Error(err)
	suite.Require().Nil(badRes)

	// Add an atomic swap and build refund msg
	swapID, _ := suite.AddAtomicSwap()
	msg := types.NewMsgRefundAtomicSwap(suite.addrs[0].String(), swapID)

	// Attempt to refund active atomic swap
	res1, err := suite.msgServer.RefundAtomicSwap(sdk.WrapSDKContext(suite.ctx), &msg)
	suite.Require().Error(err)
	suite.Require().Nil(res1)

	// Expire the atomic swap with begin blocker and attempt refund
	laterCtx := suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400)
	bep3.BeginBlocker(laterCtx, suite.keeper)
	res2, err := suite.msgServer.RefundAtomicSwap(sdk.WrapSDKContext(laterCtx), &msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res2)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}
