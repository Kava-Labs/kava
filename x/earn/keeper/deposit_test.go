package keeper_test

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	os.Exit(m.Run())
}

type keeperTestSuite struct {
	testutil.Suite
}

func (suite *keeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}

func (suite *keeperTestSuite) TestDeposit() {
	suite.CreateVault("busd", types.STRATEGY_TYPE_STABLECOIN_STAKERS)

	acc := suite.CreateAccount(sdk.NewCoins(sdk.NewInt64Coin("busd", 1000)))

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), sdk.NewInt64Coin("busd", 100))
	suite.Require().NoError(err)
}
