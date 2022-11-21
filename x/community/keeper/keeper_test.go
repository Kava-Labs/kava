package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite
	App        app.TestApp
	Ctx        sdk.Context
	Keeper     keeper.Keeper
	BankKeeper bankkeeper.Keeper
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	tApp.InitializeFromGenesisStates()

	suite.App = tApp
	suite.Ctx = ctx
	suite.Keeper = tApp.GetCommunityKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// CreateFundedAccount creates a random account and mints `coins` to it.
func (suite *KeeperTestSuite) CreateFundedAccount(coins sdk.Coins) sdk.AccAddress {
	addr := app.RandomAddress()
	err := suite.App.FundAccount(suite.Ctx, addr, coins)
	suite.Require().NoError(err)
	return addr
}

func (suite *KeeperTestSuite) TestFundCommunityPool() {
	suite.SetupTest()
	maccAddr := suite.App.GetAccountKeeper().GetModuleAddress(types.ModuleAccountName)

	funds := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10000)),
		sdk.NewCoin("usdx", sdk.NewInt(100)),
	)
	sender := suite.CreateFundedAccount(funds)

	err := suite.Keeper.FundCommunityPool(suite.Ctx, sender, funds)
	suite.Require().NoError(err)

	// check that community pool received balance
	suite.App.CheckBalance(suite.T(), suite.Ctx, maccAddr, funds)
	// check that sender had balance deducted
	suite.App.CheckBalance(suite.T(), suite.Ctx, sender, sdk.NewCoins())
}
