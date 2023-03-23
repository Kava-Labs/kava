package e2e_test

import (
	"fmt"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
)

// TestUpgradeHandler can be used to run tests post-upgrade. If an upgrade is enabled, all tests
// are run against the upgraded chain. However, this file is a good place to consolidate all
// acceptance tests for a given set of upgrade handlers.
func (suite IntegrationTestSuite) TestUpgradeHandler() {
	suite.SkipIfUpgradeDisabled()
	fmt.Println("An upgrade has run!")
	suite.True(true)

	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// check community pool balance before & after
	beforeCommPoolBalance, err := suite.Kava.Community.Balance(beforeUpgradeCtx, &communitytypes.QueryBalanceRequest{})
	suite.NoError(err)
	afterCommPoolBalance, err := suite.Kava.Community.Balance(afterUpgradeCtx, &communitytypes.QueryBalanceRequest{})
	suite.NoError(err)

	// expect no balance before upgrade
	suite.True(beforeCommPoolBalance.Coins.Empty())
	// expect handler-moved tokens to be in macc now!
	suite.Equal(app.MovedCommunityPoolFunds, afterCommPoolBalance.Coins)

	// check params before & after
	beforeEarnParams, err := suite.Kava.Earn.Params(beforeUpgradeCtx, &earntypes.QueryParamsRequest{})
	suite.NoError(err)
	afterEarnParams, err := suite.Kava.Earn.Params(afterUpgradeCtx, &earntypes.QueryParamsRequest{})
	suite.NoError(err)

	// expect more than one vault
	suite.Greater(len(beforeEarnParams.Params.AllowedVaults), 1)
	// expect only one vault post-upgrade
	suite.Equal(len(afterEarnParams.Params.AllowedVaults), 1)
}
