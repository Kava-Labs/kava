package e2e_test

import (
	"fmt"
)

// TestUpgradeHandler can be used to run tests post-upgrade. If an upgrade is enabled, all tests
// are run against the upgraded chain. However, this file is a good place to consolidate all
// acceptance tests for a given set of upgrade handlers.
func (suite *IntegrationTestSuite) TestUpgradeHandler() {
	suite.SkipIfUpgradeDisabled()
	fmt.Println("An upgrade has run!")
	suite.True(true)

	// Uncomment & use these contexts to compare chain state before & after the upgrade occurs.
	// beforeUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight - 1)
	// afterUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight)
}
