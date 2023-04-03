package e2e_test

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
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

func (suite *IntegrationTestSuite) TestEthCallToErc20() {
	randoReceiver := util.SdkToEvmAddress(app.RandomAddress())
	amount := big.NewInt(1e6)

	// make unauthenticated eth_call query to check balance
	beforeBalance := suite.GetErc20Balance(randoReceiver)

	// make authenticate eth_call to transfer tokens
	res := suite.FundKavaErc20Balance(randoReceiver, amount)
	suite.NoError(res.Err)

	// make another unauthenticated eth_call query to check new balance
	afterBalance := suite.GetErc20Balance(randoReceiver)

	suite.BigIntsEqual(big.NewInt(0), beforeBalance, "expected before balance to be zero")
	suite.BigIntsEqual(amount, afterBalance, "unexpected post-transfer balance")
}
