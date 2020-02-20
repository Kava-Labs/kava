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

// TODO: implement keeper tests
// 	_, addrs := app.GeneratePrivKeyAddressPairs(1)
// 	ak := suite.app.GetAccountKeeper()
// 	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
// 	acc.SetCoins(sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000000000))))
// 	ak.SetAccount(suite.ctx, acc)

// 	err := suite.keeper.CreateAtomicSwap(suite.ctx, kavaAddrs[0], kavaAddrs[1], "", "0x9eB05a790e2De0a047a57a22199D8CccEA6d6D5A", randomNumberHashes[0], timestamps[0], coinsSingle, "99btc", 1000, false)
// 	suite.NoError(err)
// }

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
