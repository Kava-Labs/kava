package incentive_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite
	"github.com/kava-labs/kava/app"
  "github.com/kava-labs/kava/x/incentive"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     app.TestApp
	handler sdk.Handler
	keeper  incentive.Keeper
	addrs   []sdk.AccAddress
}

func (suite *HandlerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetIncentiveKeeper()

	// Set up genesis state and initialize
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := []sdk.Coins{}
	for j := 0; j < 3; j++ {
		coins = append(coins, cs(c("bnb", 10000000000), c("ukava", 10000000000)))
	}
	authGS := app.NewAuthGenState(addrs, coins)
	tApp.InitializeFromGenesisStates(authGS)

	suite.addrs = addrs
	suite.handler = bep3.NewHandler(keeper)
	suite.keeper = keeper
	suite.app = tApp
	suite.ctx = ctx
}


func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}