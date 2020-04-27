package incentive_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/kavadist"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }

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
	suite.handler = incentive.NewHandler(keeper)
	suite.keeper = keeper
	suite.app = tApp
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) addClaim() {
	supplyKeeper := suite.app.GetSupplyKeeper()
	macc := supplyKeeper.GetModuleAccount(suite.ctx, kavadist.ModuleName)
	err := supplyKeeper.MintCoins(suite.ctx, macc.GetName(), cs(c("ukava", 1000000)))
	suite.Require().NoError(err)
	cp := incentive.NewClaimPeriod("bnb", 1, suite.ctx.BlockTime().Add(time.Hour*168), time.Hour*8766)
	suite.NotPanics(func() {
		suite.keeper.SetClaimPeriod(suite.ctx, cp)
	})
	c1 := incentive.NewClaim(suite.addrs[0], c("ukava", 1000000), "bnb", 1)
	suite.NotPanics(func() {
		suite.keeper.SetClaim(suite.ctx, c1)
	})
}

func (suite *HandlerTestSuite) TestMsgClaimReward() {
	suite.addClaim()
	msg := incentive.NewMsgClaimReward(suite.addrs[0], "bnb")
	res, err := suite.handler(suite.ctx, msg)
	suite.NoError(err)
	suite.Require().NotNil(res)
}
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
