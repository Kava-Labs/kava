package e2e_test

import (
	"context"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
)

// func (suite *IntegrationTestSuite) TestEthCallToGreeterContract() {
// 	// this test manipulates state of the Greeter contract which means other tests shouldn't use it.

// 	// setup funded account to interact with contract
// 	user := suite.Kava.NewFundedAccount("greeter-contract-user", sdk.NewCoins(ukava(10e6)))

// 	greeterAddr := suite.Kava.ContractAddrs["greeter"]
// 	contract, err := greeter.NewGreeter(greeterAddr, suite.Kava.EvmClient)
// 	suite.NoError(err)

// 	beforeGreeting, err := contract.Greet(nil)
// 	suite.NoError(err)

// 	updatedGreeting := "look at me, using the evm"
// 	tx, err := contract.SetGreeting(user.EvmAuth, updatedGreeting)
// 	suite.NoError(err)

// 	_, err = util.WaitForEvmTxReceipt(suite.Kava.EvmClient, tx.Hash(), 10*time.Second)
// 	suite.NoError(err)

// 	afterGreeting, err := contract.Greet(nil)
// 	suite.NoError(err)

// 	suite.Equal("what's up!", beforeGreeting)
// 	suite.Equal(updatedGreeting, afterGreeting)
// }

// func (suite *IntegrationTestSuite) TestEthCallToErc20() {
// 	randoReceiver := util.SdkToEvmAddress(app.RandomAddress())
// 	amount := big.NewInt(1e6)

// 	// make unauthenticated eth_call query to check balance
// 	beforeBalance := suite.GetErc20Balance(randoReceiver)

// 	// make authenticate eth_call to transfer tokens
// 	res := suite.FundKavaErc20Balance(randoReceiver, amount)
// 	suite.NoError(res.Err)

// 	// make another unauthenticated eth_call query to check new balance
// 	afterBalance := suite.GetErc20Balance(randoReceiver)

// 	suite.BigIntsEqual(big.NewInt(0), beforeBalance, "expected before balance to be zero")
// 	suite.BigIntsEqual(amount, afterBalance, "unexpected post-transfer balance")
// }

func (suite *IntegrationTestSuite) TestEip712BasicMessageAuthorization() {
	// create new funded account
	depositor := suite.Kava.NewFundedAccount("eip712-msgSend", sdk.NewCoins(ukava(100e6)))

	// setup message for sending 1KAVA to random receiver
	receiver := app.RandomAddress()
	msgs := []sdk.Msg{
		banktypes.NewMsgSend(depositor.SdkAddress, receiver, sdk.NewCoins(ukava(1e6))),
	}

	// create tx
	tx := suite.NewEip712TxBuilder(
		depositor,
		suite.Kava,
		1e6,
		sdk.NewCoins(ukava(1e4)),
		msgs,
		"this is a memo",
	).GetTx()

	txBytes, err := suite.Kava.EncodingConfig.TxConfig.TxEncoder()(tx)
	suite.NoError(err)

	res, err := suite.Kava.Tx.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	})

	suite.NoError(err)
	suite.Equal(sdkerrors.SuccessABCICode, res.TxResponse.Code)

	fmt.Println("txhash! ", res.TxResponse.TxHash)
	fmt.Println("height: ", res.TxResponse.Height)
	fmt.Println("logs: ", res.TxResponse.RawLog)

	_, err = util.WaitForSdkTxCommit(suite.Kava.Tx, res.TxResponse.TxHash, 6*time.Second)
	suite.NoError(err)

	// check that the message was processed & the kava is transferred.
	balRes, err := suite.Kava.Bank.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: receiver.String(),
		Denom:   "ukava",
	})
	suite.NoError(err)
	suite.Equal(sdk.NewInt(1e6), balRes.Balance.Amount)
}
