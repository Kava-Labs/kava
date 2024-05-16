package e2e_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/contracts/contracts/noop_caller"
	"github.com/kava-labs/kava/precompile/contracts/noop"
)

// TestNoopPrecompileInitialization tests that enabled precompiles are initialized (nonce and code are set),
// and correspondingly disabled precompiles are not initialized.
// Noop precompile is enabled on noop.ContractAddress address, but disabled on noop.ContractAddress2.
func (suite *IntegrationTestSuite) TestNoopPrecompileInitialization() {
	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	suite.Run("registered and enabled", func() {
		nonce, err := ethClient.NonceAt(context.Background(), noop.ContractAddress, nil)
		suite.Require().NoError(err)
		suite.Require().Equal(uint64(1), nonce)

		code, err := ethClient.CodeAt(context.Background(), noop.ContractAddress, nil)
		suite.Require().NoError(err)
		suite.Require().Equal([]byte{1}, code)
	})

	suite.Run("registered but disabled", func() {
		nonce, err := ethClient.NonceAt(context.Background(), noop.ContractAddress2, nil)
		suite.Require().NoError(err)
		suite.Require().Equal(uint64(0), nonce)

		code, err := ethClient.CodeAt(context.Background(), noop.ContractAddress2, nil)
		suite.Require().NoError(err)
		suite.Require().Empty([]byte{}, code)
	})
}

// TestNoopPrecompileIndirectCall tests EOA -> NoopCaller -> Precompile scenarios, meaning: EOA calls NoopCaller,
// NoopCaller calls precompile.
// In test we consider 2 main scenarios:
// - calling enabled precompile
// - calling disabled precompile
// Calling enabled precompile successfully calls precompile (which does nothing) and returns without an error.
// Calling disabled precompile is the same as calling random non-existent address (which does nothing).
// Errors:
// - calling disabled precompile doesn't return an error on evm-level, but it may still return a client-side error (depending on how it was called).
// - Regular type-safe call returns an error because solidity compiler adds additional check that contract code isn't empty.
// - Static call opcode doesn't return an error.
func (suite *IntegrationTestSuite) TestNoopPrecompileIndirectCall() {
	suite.Run("enabled precompile", func() {
		_, enabledNoopCaller := suite.deployNoopCaller("noop-precompile-indirect-call-enabled-precompile", noop.ContractAddress)

		suite.Run("regular type-safe call", func() {
			err := enabledNoopCaller.Noop(nil)
			suite.Require().NoError(err)
		})

		suite.Run("evm static call opcode", func() {
			_, err := enabledNoopCaller.NoopStaticCall(nil)
			suite.Require().NoError(err)
		})
	})

	suite.Run("disabled precompile", func() {
		_, disabledNoopCaller := suite.deployNoopCaller("noop-precompile-indirect-call-disabled-precompile", noop.ContractAddress2)

		suite.Run("regular type-safe call", func() {
			err := disabledNoopCaller.Noop(nil)
			suite.Require().ErrorContains(err, "execution reverted")
		})

		suite.Run("evm static call opcode", func() {
			// static call doesn't return an error, because it bypasses contract code isn't empty check
			_, err := disabledNoopCaller.NoopStaticCall(nil)
			suite.Require().NoError(err)
		})
	})
}

// TestNoopPrecompileDirectCall tests EOA -> Precompile scenarios, meaning: EOA directly calls precompile.
func (suite *IntegrationTestSuite) TestNoopPrecompileDirectCall() {
	suite.Run("enabled precompile", func() {
		enabledNoopCaller := suite.newNoopCaller(noop.ContractAddress)

		suite.Run("regular type-safe call", func() {
			err := enabledNoopCaller.Noop(nil)
			suite.Require().NoError(err)
		})
	})

	suite.Run("disabled precompile", func() {
		disabledNoopCaller := suite.newNoopCaller(noop.ContractAddress2)

		suite.Run("regular type-safe call", func() {
			err := disabledNoopCaller.Noop(nil)
			suite.Require().ErrorContains(err, "no contract code at given address")
		})
	})
}

// call directly via JSON-RPC: expect CALL Context
// regular type-safe call via NoopCaller: expect CALL Context

// call opcode
// staticcall
// delegatecall
// callcode
//
// read-only flag
func (suite *IntegrationTestSuite) TestNoopPrecompileCallContext() {
	whale := suite.Kava.NewFundedAccount("noop-caller-events-whale-account", sdk.NewCoins(ukava(1e6))) // 1 KAVA

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	noopCaller := suite.newNoopCaller(noop.ContractAddress)

	_, err = noopCaller.EmitEvent(whale.EvmAuth)
	suite.Require().NoError(err)
	suite.waitForNewBlocks(1)

	// TODO(yevhenii):
	// - specify block range?
	// - specify NoopPrecompile contract address
	logs, err := ethClient.FilterLogs(context.Background(), ethereum.FilterQuery{
		//BlockHash *common.Hash     // used by eth_getLogs, return logs only from block with this hash
		//FromBlock *big.Int         // beginning of the queried range, nil means genesis block
		//ToBlock   *big.Int         // end of the range, nil means latest block
		//Addresses []common.Address // restricts matches to events created by specific contracts

		// The Topic list restricts matches to particular event topics. Each event has a list
		// of topics. Topics matches a prefix of that list. An empty element slice matches any
		// topic. Non-empty elements represent an alternative that matches any of the
		// contained topics.
		//
		// Examples:
		// {} or nil          matches any topic list
		// {{A}}              matches topic A in first position
		// {{}, {B}}          matches any topic in first position AND B in second position
		// {{A}, {B}}         matches topic A in first position AND B in second position
		// {{A, B}, {C, D}}   matches topic (A OR B) in first position AND (C OR D) in second position
		//Topics [][]common.Hash
	})
	suite.Require().NoError(err)
	fmt.Printf("logs: %v\n", logs)

	logsInJSON, err := json.Marshal(logs)
	suite.Require().NoError(err)
	fmt.Printf("logsInJSON: %s\n", logsInJSON)

	suite.Require().Len(logs, 1)
}

func (suite *IntegrationTestSuite) TestNoopCallerEvents() {
	whale := suite.Kava.NewFundedAccount("noop-caller-events-whale-account", sdk.NewCoins(ukava(1e6))) // 1 KAVA

	noopCallerAddr, noopCaller := suite.deployNoopCaller("noop-caller-events", noop.ContractAddress)

	_, err := noopCaller.EmitEvent(whale.EvmAuth)
	suite.Require().NoError(err)
	suite.waitForNewBlocks(1)

	eventSig := []byte("Event(string,string)")
	eventSigHash := crypto.Keccak256Hash(eventSig)

	logs, err := suite.Kava.EvmClient.FilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: []common.Address{noopCallerAddr},
		Topics: [][]common.Hash{
			{eventSigHash},
		},
	})
	suite.Require().NoError(err)
	suite.Require().Len(logs, 1)

	contractAbi, err := abi.JSON(strings.NewReader(noop_caller.NoopCallerMetaData.ABI))
	suite.Require().NoError(err)

	//logsInJSON, err := json.Marshal(logs)
	//suite.Require().NoError(err)
	//fmt.Printf("logsInJSON: %s\n", logsInJSON)

	event := struct {
		Param string `abi:"param"`
	}{}
	err = contractAbi.UnpackIntoInterface(&event, "Event", logs[0].Data)
	suite.Require().NoError(err)
	suite.Require().Equal("test-param", event.Param)
}

func (suite *IntegrationTestSuite) deployNoopCaller(name string, target common.Address) (common.Address, *noop_caller.NoopCaller) {
	helperAcc := suite.Kava.NewFundedAccount(name, sdk.NewCoins(ukava(1e6))) // 1 KAVA

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	noopCallerAddr, _, noopCaller, err := noop_caller.DeployNoopCaller(helperAcc.EvmAuth, ethClient, target)
	suite.Require().NoError(err)

	// wait until contract is deployed
	suite.Eventually(func() bool {
		code, err := ethClient.CodeAt(context.Background(), noopCallerAddr, nil)
		suite.Require().NoError(err)

		return len(code) != 0
	}, 20*time.Second, 1*time.Second)

	return noopCallerAddr, noopCaller
}

func (suite *IntegrationTestSuite) newNoopCaller(target common.Address) *noop_caller.NoopCaller {
	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	noopCaller, err := noop_caller.NewNoopCaller(target, ethClient)
	suite.Require().NoError(err)

	return noopCaller
}

func (suite *IntegrationTestSuite) waitForNewBlocks(n uint64) {
	beginHeight, err := suite.Kava.EvmClient.BlockNumber(context.Background())
	suite.Require().NoError(err)

	suite.Eventually(func() bool {
		curHeight, err := suite.Kava.EvmClient.BlockNumber(context.Background())
		suite.Require().NoError(err)

		return curHeight >= beginHeight+n
	}, 10*time.Second, 1*time.Second)
}
