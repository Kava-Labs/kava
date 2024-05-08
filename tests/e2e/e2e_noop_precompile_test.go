package e2e_test

import (
	"context"
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
		enabledNoopCaller := suite.deployNoopCaller("noop-precompile-indirect-call-enabled-precompile", noop.ContractAddress)

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
		disabledNoopCaller := suite.deployNoopCaller("noop-precompile-indirect-call-disabled-precompile", noop.ContractAddress2)

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

func (suite *IntegrationTestSuite) deployNoopCaller(name string, target common.Address) *noop_caller.NoopCaller {
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

	return noopCaller
}

func (suite *IntegrationTestSuite) newNoopCaller(target common.Address) *noop_caller.NoopCaller {
	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	noopCaller, err := noop_caller.NewNoopCaller(target, ethClient)
	suite.Require().NoError(err)

	return noopCaller
}
