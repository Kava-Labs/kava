package testutil

import (
	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// Test suite used for all community tests
type Suite struct {
	suite.Suite
	App    app.TestApp
	Ctx    sdk.Context
	Keeper keeper.Keeper

	MaccAddress sdk.AccAddress
}

// The default state used by each test
func (suite *Suite) SetupTest() {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	suite.App = tApp.InitializeFromGenesisStates()

	suite.Ctx = ctx
	suite.Keeper = tApp.GetCommunityKeeper()
	communityPoolAddress := tApp.GetAccountKeeper().GetModuleAddress(types.ModuleAccountName)
	// hello, greppers!
	suite.Equal("kava17d2wax0zhjrrecvaszuyxdf5wcu5a0p4qlx3t5", communityPoolAddress.String())
	suite.MaccAddress = communityPoolAddress
}

// CreateFundedAccount creates a random account and mints `coins` to it.
func (suite *Suite) CreateFundedAccount(coins sdk.Coins) sdk.AccAddress {
	addr := app.RandomAddress()
	err := suite.App.FundAccount(suite.Ctx, addr, coins)
	suite.Require().NoError(err)
	return addr
}
