package e2e_test

import (
	"context"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/contracts/contracts/mul3_caller"
	"github.com/kava-labs/kava/precompile/contracts/mul3"
)

func (suite *IntegrationTestSuite) TestMul3Precompile_EOAToContractToPrecompile_Success() {
	// setup helper account
	helperAcc := suite.Kava.NewFundedAccount("mul3-precompile-eoa->contract->precompile-success-helper-account", sdk.NewCoins(ukava(1e6))) // 1 KAVA

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	mul3CallerAddr, _, mul3Caller, err := mul3_caller.DeployMul3Caller(helperAcc.EvmAuth, ethClient)
	suite.Require().NoError(err)

	// wait until contract is deployed
	suite.Eventually(func() bool {
		code, err := ethClient.CodeAt(context.Background(), mul3CallerAddr, nil)
		suite.Require().NoError(err)

		return len(code) != 0
	}, 20*time.Second, 1*time.Second)

	suite.Run("regular type-safe call", func() {
		for _, tc := range []struct {
			desc string
			a    *big.Int
			b    *big.Int
			c    *big.Int
			rez  *big.Int
		}{
			{
				desc: "test case #1",
				a:    big.NewInt(2),
				b:    big.NewInt(3),
				c:    big.NewInt(4),
				rez:  big.NewInt(24),
			},
			{
				desc: "test case #2",
				a:    big.NewInt(3),
				b:    big.NewInt(5),
				c:    big.NewInt(7),
				rez:  big.NewInt(105),
			},
		} {
			_, err = mul3Caller.CalcMul3(helperAcc.EvmAuth, tc.a, tc.b, tc.c)
			suite.Require().NoError(err)

			suite.Eventually(func() bool {
				rez, err := mul3Caller.GetMul3(nil)
				suite.Require().NoError(err)

				equal := rez.Cmp(tc.rez) == 0
				return equal
			}, 10*time.Second, 1*time.Second)
		}
	})

	suite.Run("evm call & static call opcodes", func() {
		for _, tc := range []struct {
			desc string
			a    *big.Int
			b    *big.Int
			c    *big.Int
			rez  *big.Int
		}{
			{
				desc: "test case #1",
				a:    big.NewInt(2),
				b:    big.NewInt(3),
				c:    big.NewInt(4),
				rez:  big.NewInt(24),
			},
			{
				desc: "test case #2",
				a:    big.NewInt(3),
				b:    big.NewInt(5),
				c:    big.NewInt(7),
				rez:  big.NewInt(105),
			},
		} {
			_, err = mul3Caller.CalcMul3Call(helperAcc.EvmAuth, tc.a, tc.b, tc.c)
			suite.Require().NoError(err)

			suite.Eventually(func() bool {
				rezBytes, err := mul3Caller.GetMul3StaticCall(nil)
				suite.Require().NoError(err)

				var rez big.Int
				rez.SetBytes(rezBytes)

				equal := rez.Cmp(tc.rez) == 0
				return equal
			}, 10*time.Second, 1*time.Second)
		}
	})
}

func (suite *IntegrationTestSuite) TestMul3Precompile_EOAToPrecompile_Success() {
	// setup helper account
	helperAcc := suite.Kava.NewFundedAccount("mul3-precompile-eoa->precompile-success-helper-account", sdk.NewCoins(ukava(1e6))) // 1 KAVA

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	mul3Caller, err := mul3_caller.NewMul3Caller(mul3.ContractAddress, ethClient)
	suite.Require().NoError(err)

	suite.Run("regular type-safe call", func() {
		for _, tc := range []struct {
			desc string
			a    *big.Int
			b    *big.Int
			c    *big.Int
			rez  *big.Int
		}{
			{
				desc: "test case #1",
				a:    big.NewInt(2),
				b:    big.NewInt(3),
				c:    big.NewInt(4),
				rez:  big.NewInt(24),
			},
			{
				desc: "test case #2",
				a:    big.NewInt(3),
				b:    big.NewInt(5),
				c:    big.NewInt(7),
				rez:  big.NewInt(105),
			},
		} {
			_, err = mul3Caller.CalcMul3(helperAcc.EvmAuth, tc.a, tc.b, tc.c)
			suite.Require().NoError(err)

			suite.Eventually(func() bool {
				rez, err := mul3Caller.GetMul3(nil)
				suite.Require().NoError(err)

				equal := rez.Cmp(tc.rez) == 0
				return equal
			}, 10*time.Second, 1*time.Second)
		}
	})
}

func (suite *IntegrationTestSuite) TestMul3Precompile_EOAToPrecompile_Failure() {
	// setup helper account
	helperAcc := suite.Kava.NewFundedAccount("mul3-precompile-eoa->precompile-failure-helper-account", sdk.NewCoins(ukava(1e6))) // 1 KAVA

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	mul3Caller, err := mul3_caller.NewMul3Caller(mul3.ContractAddress, ethClient)
	suite.Require().NoError(err)

	// successful scenario, after execution we'll have mul=24 in the state
	{
		_, err = mul3Caller.CalcMul3(helperAcc.EvmAuth, big.NewInt(2), big.NewInt(3), big.NewInt(4))
		suite.Require().NoError(err)

		suite.Eventually(func() bool {
			rez, err := mul3Caller.GetMul3(nil)
			suite.Require().NoError(err)

			equal := rez.Cmp(big.NewInt(24)) == 0
			return equal
		}, 10*time.Second, 1*time.Second)
	}

	// revert scenario, after execution we'll still have mul=24 in the state
	{
		_, err = mul3Caller.CalcMul3WithError(helperAcc.EvmAuth, big.NewInt(3), big.NewInt(5), big.NewInt(7))
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "calculation error")

		// wait for few blocks, and make sure that failed transaction didn't change the state
		suite.waitForNewBlocks(ethClient, 3)

		rez, err := mul3Caller.GetMul3(nil)
		suite.Require().NoError(err)

		equal := rez.Cmp(big.NewInt(24)) == 0
		suite.Require().True(equal)
	}
}

func (suite *IntegrationTestSuite) waitForNewBlocks(ethClient *ethclient.Client, n uint64) {
	beginHeight, err := ethClient.BlockNumber(context.Background())
	suite.Require().NoError(err)

	suite.Eventually(func() bool {
		curHeight, err := ethClient.BlockNumber(context.Background())
		suite.Require().NoError(err)

		return curHeight >= beginHeight+n
	}, 10*time.Second, 1*time.Second)
}
