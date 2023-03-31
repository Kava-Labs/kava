package e2e_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/tests/e2e/contracts/greeter"
	"github.com/kava-labs/kava/tests/util"
)

func (suite *IntegrationTestSuite) TestEthCallToGreeterContract() {
	// this test manipulates state of the Greeter contract which means other tests shouldn't use it.

	// setup funded account to interact with contract
	user := suite.Kava.NewFundedAccount("greeter-contract-user", sdk.NewCoins(ukava(10e6)))

	greeterAddr := suite.Kava.ContractAddrs["greeter"]
	contract, err := greeter.NewGreeter(greeterAddr, suite.Kava.EvmClient)
	suite.NoError(err)

	beforeGreeting, err := contract.Greet(nil)
	suite.NoError(err)

	updatedGreeting := "look at me, using the evm"
	tx, err := contract.SetGreeting(user.EvmAuth, updatedGreeting)
	suite.NoError(err)

	_, err = util.WaitForEvmTxReceipt(suite.Kava.EvmClient, tx.Hash(), 10*time.Second)
	suite.NoError(err)

	afterGreeting, err := contract.Greet(nil)
	suite.NoError(err)

	suite.Equal("what's up!", beforeGreeting)
	suite.Equal(updatedGreeting, afterGreeting)
}
