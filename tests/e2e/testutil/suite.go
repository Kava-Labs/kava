package testutil

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/e2e/runner"
)

const (
	ChainId           = "kavalocalnet_8888-1"
	FundedAccountName = "whale"
	StakingDenom      = "ukava"
	// use coin type 60 so we are compatible with accounts from `kava add keys --eth <name>`
	// these accounts use the ethsecp256k1 signing algorithm that allows the signing client
	// to manage both sdk & evm txs.
	Bip44CoinType = 60
)

type E2eTestSuite struct {
	suite.Suite

	runner runner.NodeRunner

	Kava *Chain
	Ibc  *Chain
}

func (suite *E2eTestSuite) SetupSuite() {
	// TODO: env var that toggles IBC tests.
	includeIbc := true

	var err error
	fmt.Println("setting up test suite.")
	app.SetSDKConfig()

	// this mnemonic is expected to be a funded account that can seed the funds for all
	// new accounts created during tests. it will be available under Accounts["whale"]
	fundedAccountMnemonic := os.Getenv("E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC")
	if fundedAccountMnemonic == "" {
		suite.Fail("no E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC provided")
	}

	config := runner.Config{
		IncludeIBC: includeIbc,
		ImageTag:   "local",
	}
	suite.runner = runner.NewKavaNode(config)

	chains := suite.runner.StartChains()
	kavachain := chains.MustGetChain("kava")
	suite.Kava, err = NewChain(suite.T(), kavachain, fundedAccountMnemonic)
	if err != nil {
		suite.runner.Shutdown()
		suite.T().Fatalf("failed to create kava chain querier: %s", err)
	}

	if includeIbc {
		ibcchain := chains.MustGetChain("ibc")
		suite.Ibc, err = NewChain(suite.T(), ibcchain, fundedAccountMnemonic)
		if err != nil {
			suite.runner.Shutdown()
			suite.T().Fatalf("failed to create ibc chain querier: %s", err)
		}
	}
}

func (suite *E2eTestSuite) TearDownSuite() {
	fmt.Println("tearing down test suite.")
	// close all account request channels
	suite.Kava.Shutdown()
	if suite.Ibc != nil {
		suite.Ibc.Shutdown()
	}
	// gracefully shutdown docker container(s)
	suite.runner.Shutdown()
}
