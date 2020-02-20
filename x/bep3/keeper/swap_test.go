package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
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
	keeper := tApp.GetBep3Keeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	return
}

// TODO: test panicked: UnmarshalJSON cannot decode empty bytes
// func (suite *AtomicSwapTestSuite) TestCreateAtomicSwap() {
// 	_, addrs := app.GeneratePrivKeyAddressPairs(1)
// 	ak := suite.app.GetAccountKeeper()
// 	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
// 	acc.SetCoins(sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000000000))))
// 	ak.SetAccount(suite.ctx, acc)

// 	expectedSwapIDBytes, err := types.CalculateSwapID(randomNumberHashes[1], binanceAddrs[0], "")
// 	suite.NoError(err)
// 	expectedSwapID := types.BytesToHexEncodedString(expectedSwapIDBytes)

// 	swapID, err := suite.keeper.CreateAtomicSwap(suite.ctx, binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHashes[1], timestamps[1], coinsSingle, "50000bnb", 80000, false)
// 	suite.NoError(err)
// 	suite.Equal(swapID, expectedSwapID)
// }

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
