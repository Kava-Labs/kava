package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type keeperTestSuite struct {
	testutil.Suite
}

func (suite *keeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *keeperTestSuite) TestGetAllAccounts_NoBalance() {
	accounts := suite.Suite.Keeper.GetAllAccounts(suite.Ctx)
	suite.Require().Equal(0, len(accounts))
}

func (suite *keeperTestSuite) TestGetAllAccounts_ReturnAccounts() {
	suite.Suite.Keeper.SetBalance(suite.Ctx, suite.Suite.Addrs[0], sdk.NewInt(100))
	suite.Suite.Keeper.SetBalance(suite.Ctx, suite.Suite.Addrs[1], sdk.NewInt(200))
	accounts := suite.Suite.Keeper.GetAllAccounts(suite.Ctx)
	expected := []types.Account{
		{Address: suite.Suite.Addrs[0], Balance: sdk.NewInt(100)},
		{Address: suite.Suite.Addrs[1], Balance: sdk.NewInt(200)},
	}
	suite.Require().Equal(expected, accounts)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}
