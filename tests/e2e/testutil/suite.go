package testutil

import (
	"fmt"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/e2e/runner"
)

const (
	FundedAccountName = "whale"
	// use coin type 60 so we are compatible with accounts from `kava add keys --eth <name>`
	// these accounts use the ethsecp256k1 signing algorithm that allows the signing client
	// to manage both sdk & evm txs.
	Bip44CoinType = 60

	IbcPort    = "transfer"
	IbcChannel = "channel-0"
)

type E2eTestSuite struct {
	suite.Suite

	config SuiteConfig
	runner runner.NodeRunner

	Kava *Chain
	Ibc  *Chain
}

func (suite *E2eTestSuite) SetupSuite() {
	var err error
	fmt.Println("setting up test suite.")
	app.SetSDKConfig()

	suiteConfig := ParseSuiteConfig()
	suite.config = suiteConfig

	runnerConfig := runner.Config{
		IncludeIBC: suiteConfig.IncludeIbcTests,
		ImageTag:   "local",

		EnableAutomatedUpgrade:  suiteConfig.IncludeAutomatedUpgrade,
		KavaUpgradeName:         suiteConfig.KavaUpgradeName,
		KavaUpgradeHeight:       suiteConfig.KavaUpgradeHeight,
		KavaUpgradeBaseImageTag: suiteConfig.KavaUpgradeBaseImageTag,
	}
	suite.runner = runner.NewKavaNode(runnerConfig)

	chains := suite.runner.StartChains()
	kavachain := chains.MustGetChain("kava")
	suite.Kava, err = NewChain(suite.T(), kavachain, suiteConfig.FundedAccountMnemonic)
	if err != nil {
		suite.runner.Shutdown()
		suite.T().Fatalf("failed to create kava chain querier: %s", err)
	}

	if suiteConfig.IncludeIbcTests {
		ibcchain := chains.MustGetChain("ibc")
		suite.Ibc, err = NewChain(suite.T(), ibcchain, suiteConfig.FundedAccountMnemonic)
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
	if !suite.config.SkipShutdown {
		suite.runner.Shutdown()
	}
}

func (suite *E2eTestSuite) SkipIfIbcDisabled() {
	if !suite.config.IncludeIbcTests {
		suite.T().SkipNow()
	}
}

func (suite *E2eTestSuite) SkipIfUpgradeDisabled() {
	if !suite.config.IncludeAutomatedUpgrade {
		suite.T().SkipNow()
	}
}
