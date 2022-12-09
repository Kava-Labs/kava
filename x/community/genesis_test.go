package community_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/community"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
)

type genesisTestSuite struct {
	testutil.Suite
}

func (suite *genesisTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(genesisTestSuite))
}

func (suite *genesisTestSuite) TestInitGenesis() {
	suite.SetupTest()

	accountKeeper := suite.App.GetAccountKeeper()

	suite.NotPanics(func() {
		community.InitGenesis(suite.Ctx, suite.Keeper, accountKeeper)
	})

	communityPoolAddress := accountKeeper.GetModuleAddress(types.ModuleAccountName)
	// hello, greppers!
	suite.Equal("kava17d2wax0zhjrrecvaszuyxdf5wcu5a0p4qlx3t5", communityPoolAddress.String())

	// check for module account this way b/c GetModuleAccount creates if not existing.
	acc := accountKeeper.GetAccount(suite.Ctx, communityPoolAddress)
	suite.NotNil(acc)
	_, ok := acc.(authtypes.ModuleAccountI)
	suite.True(ok)
}
