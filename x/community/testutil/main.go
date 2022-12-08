package testutil

import (
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
)

// Test suite used for all community tests
type Suite struct {
	suite.Suite
	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper
}

// The default state used by each test
func (suite *Suite) SetupTest() {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	tApp.InitializeFromGenesisStates()

	suite.App = tApp
	suite.Ctx = ctx
	suite.Keeper = tApp.GetCommunityKeeper()
}

// CreateFundedAccount creates a random account and mints `coins` to it.
func (suite *Suite) CreateFundedAccount(coins sdk.Coins) sdk.AccAddress {
	addr := app.RandomAddress()
	err := suite.App.FundAccount(suite.Ctx, addr, coins)
	suite.Require().NoError(err)
	return addr
}
