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
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
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
	suite.Require().False(found)
	suite.Require().NotPanics(func() {
		suite.keeper.SetClaim(suite.ctx, c)
	})
	testC, found := suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb")
	suite.Require().True(found)
	suite.Require().Equal(c, testC)
	suite.Require().NotPanics(func() {
		suite.keeper.DeleteClaim(suite.ctx, suite.addrs[0], "bnb")
	})
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb")
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestIterateClaims() {
	for i := 0; i < len(suite.addrs); i++ {
		c := types.NewClaim(suite.addrs[i], c("ukava", 100000), "bnb-a", types.NewRewardIndex("ukava", sdk.ZeroDec()))
		suite.Require().NotPanics(func() {
			suite.keeper.SetClaim(suite.ctx, c)
		})
	}
	claims := types.Claims{}
	suite.keeper.IterateClaims(suite.ctx, func(c types.Claim) bool {
		claims = append(claims, c)
		return false
	})
	suite.Require().Equal(len(suite.addrs), len(claims))

	claims = suite.keeper.GetAllClaims(suite.ctx)
	suite.Require().Equal(len(suite.addrs), len(claims))
}

func (suite *KeeperTestSuite) TestOwnerIterateClaims() {
	testCollaterals := []string{"bnb-a", "xrp-a"}
	for i := 0; i < len(suite.addrs); i++ {
		for _, collateral := range testCollaterals {
			c := types.NewClaim(suite.addrs[i], c("ukava", 100000), collateral, types.NewRewardIndex("ukava", sdk.ZeroDec()))
			suite.Require().NotPanics(func() {
				suite.keeper.SetClaim(suite.ctx, c)
			})
		}
	}
	claims := types.Claims{}
	suite.keeper.IterateClaimsByOwner(suite.ctx, suite.addrs[0], func(c types.Claim) bool {
		claims = append(claims, c)
		return false
	})
	suite.Require().Equal(len(testCollaterals), len(claims))

	claims = suite.keeper.GetAllClaimsByOwner(suite.ctx, suite.addrs[0])
	suite.Require().Equal(len(testCollaterals), len(claims))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
