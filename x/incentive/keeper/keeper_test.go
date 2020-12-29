package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	keeper := tApp.GetIncentiveKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) TestGetSetDeleteClaim() {
	c := types.NewClaim(suite.addrs[0], c("ukava", 1000000), "bnb", types.NewRewardIndex("ukava", sdk.ZeroDec()))
	_, found := suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb")
	suite.False(found)
	suite.NotPanics(func() {
		suite.keeper.SetClaim(suite.ctx, c)
	})
	testC, found := suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb")
	suite.True(found)
	suite.Equal(c, testC)
	suite.NotPanics(func() {
		suite.keeper.DeleteClaim(suite.ctx, suite.addrs[0], "bnb")
	})
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb")
	suite.False(found)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
