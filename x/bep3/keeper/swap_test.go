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

type AtomicSwapTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *AtomicSwapTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	// TODO: is this required
	tApp.InitializeFromGenesisStates(
		NewBep3GenStateMulti(),
	)
	keeper := tApp.GetBep3Keeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	return
}

// TODO: transition to table test
func (suite *AtomicSwapTestSuite) TestCreateAtomicSwap() {
	// Create two accounts
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	ak := suite.app.GetAccountKeeper()
	// Set up acc1, fund with tokens
	acc1 := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc1.SetCoins(sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000000000))))
	ak.SetAccount(suite.ctx, acc1)
	// Set up acc2
	acc2 := ak.NewAccountWithAddress(suite.ctx, addrs[1])
	ak.SetAccount(suite.ctx, acc2)

	// Set up params TODO: move to common_test
	timestamp := time.Now().Unix()
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
	heightSpan := int64(360)
	sender := acc1.GetAddress()
	recipient := acc2.GetAddress()
	senderOtherChain := binanceAddrs[0].String()
	recipientOtherChain := binanceAddrs[1].String()

	// Create an atomic swap
	err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, heightSpan, sender, recipient,
		senderOtherChain, recipientOtherChain, coinsSingle, coinsSingle.String())
	suite.NoError(err)
}

// func (suite *AtomicSwapTestSuite) TestClaimAtomicSwap() {}

// func (suite *AtomicSwapTestSuite) TestRefundAtomicSwap() {}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
