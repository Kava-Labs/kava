package testutil

import "github.com/kava-labs/kava/tests/e2e/contracts/greeter"

// InitKavaEvmData is run after the chain is running, but before the tests are run.
// It is used to initialize some EVM state, such as deploying contracts.
func (suite *E2eTestSuite) InitKavaEvmData() {
	whale := suite.Kava.GetAccount(FundedAccountName)

	// deploy an example contract
	greeterAddr, _, _, err := greeter.DeployGreeter(
		whale.evmSigner.Auth,
		whale.evmSigner.EvmClient,
		"what's up!",
	)
	suite.NoError(err, "failed to deploy a contract to the EVM")
	suite.Kava.ContractAddrs["greeter"] = greeterAddr
}
