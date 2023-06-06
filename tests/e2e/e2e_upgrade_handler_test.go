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
	// beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	// afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)
}

// Thorough testing of the upgrade handler for v0.24 depends on:
// - chain starting from v0.23 template
//   - no allowed cosmos denoms
// - EIP712 not allowing cosmos coin conversion messages

// NOTE:
// e2e_convert_cosmos_coins_test.go test confirm the following:
// - AllowedCosmosDenom gets initialized
// - Convert messages can be signed via EIP712
