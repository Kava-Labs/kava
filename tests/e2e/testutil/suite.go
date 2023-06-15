package testutil

import (
	"fmt"
	"math/big"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
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

	UpgradeHeight        int64
	DeployedErc20Address common.Address
}

func (suite *E2eTestSuite) SetupSuite() {
	var err error
	fmt.Println("setting up test suite.")
	app.SetSDKConfig()

	suiteConfig := ParseSuiteConfig()
	suite.config = suiteConfig
	suite.DeployedErc20Address = common.HexToAddress(suiteConfig.KavaErc20Address)

	// setup the correct NodeRunner for the given config
	if suiteConfig.Kvtool != nil {
		suite.runner = suite.SetupKvtoolNodeRunner()
	} else if suiteConfig.LiveNetwork != nil {
		suite.runner = suite.SetupLiveNetworkNodeRunner()
	} else {
		panic("expected either kvtool or live network configs to be defined")
	}

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

	suite.InitKavaEvmData()
}

func (suite *E2eTestSuite) TearDownSuite() {
	fmt.Println("tearing down test suite.")

	// TODO: track asset denoms & then return all funds to initial funding account.

	// close all account request channels
	suite.Kava.Shutdown()
	if suite.Ibc != nil {
		suite.Ibc.Shutdown()
	}
	// gracefully shutdown docker container(s)
	suite.runner.Shutdown()
}

func (suite *E2eTestSuite) SetupKvtoolNodeRunner() *runner.KvtoolRunner {
	// upgrade tests are only supported on kvtool networks
	suite.UpgradeHeight = suite.config.Kvtool.KavaUpgradeHeight

	runnerConfig := runner.KvtoolRunnerConfig{
		KavaConfigTemplate: suite.config.Kvtool.KavaConfigTemplate,

		IncludeIBC: suite.config.IncludeIbcTests,
		ImageTag:   "local",

		EnableAutomatedUpgrade:  suite.config.Kvtool.IncludeAutomatedUpgrade,
		KavaUpgradeName:         suite.config.Kvtool.KavaUpgradeName,
		KavaUpgradeHeight:       suite.config.Kvtool.KavaUpgradeHeight,
		KavaUpgradeBaseImageTag: suite.config.Kvtool.KavaUpgradeBaseImageTag,

		SkipShutdown: suite.config.SkipShutdown,
	}

	return runner.NewKvtoolRunner(runnerConfig)
}

func (suite *E2eTestSuite) SetupLiveNetworkNodeRunner() *runner.LiveNodeRunner {
	// live network setup doesn't presently support ibc
	if suite.config.IncludeIbcTests {
		panic("ibc tests not supported for live network configuration")
	}

	runnerConfig := runner.LiveNodeRunnerConfig{
		KavaRpcUrl:    suite.config.LiveNetwork.KavaRpcUrl,
		KavaGrpcUrl:   suite.config.LiveNetwork.KavaGrpcUrl,
		KavaEvmRpcUrl: suite.config.LiveNetwork.KavaEvmRpcUrl,
	}

	return runner.NewLiveNodeRunner(runnerConfig)
}

func (suite *E2eTestSuite) SkipIfIbcDisabled() {
	if !suite.config.IncludeIbcTests {
		suite.T().SkipNow()
	}
}

func (suite *E2eTestSuite) SkipIfUpgradeDisabled() {
	if suite.config.Kvtool != nil && suite.config.Kvtool.IncludeAutomatedUpgrade {
		suite.T().SkipNow()
	}
}

// KavaHomePath returns the OS-specific filepath for the kava home directory
// Assumes network is running with kvtool installed from the sub-repository in tests/e2e/kvtool
func (suite *E2eTestSuite) KavaHomePath() string {
	return filepath.Join("kvtool", "full_configs", "generated", "kava", "initstate", ".kava")
}

// BigIntsEqual is a helper method for comparing the equality of two big ints
func (suite *E2eTestSuite) BigIntsEqual(expected *big.Int, actual *big.Int, msg string) {
	suite.Truef(expected.Cmp(actual) == 0, "%s (expected: %s, actual: %s)", msg, expected.String(), actual.String())
}
