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

// DeployedErc20 is a type that wraps the details of the pre-deployed erc20 used by the e2e test suite.
// The Address comes from SuiteConfig.KavaErc20Address
// The CosmosDenom is fetched from the EnabledConversionPairs param of x/evmutil.
// The tests expect the following:
// - the funded account has a nonzero balance of the erc20
// - the erc20 is enabled for conversion to sdk.Coin
// - the corresponding sdk.Coin is enabled as an earn vault denom
// These requirements are checked in InitKavaEvmData().
type DeployedErc20 struct {
	Address     common.Address
	CosmosDenom string
}

// E2eTestSuite is a testify test suite for running end-to-end integration tests on Kava.
type E2eTestSuite struct {
	suite.Suite

	config SuiteConfig
	runner runner.NodeRunner

	Kava *Chain
	Ibc  *Chain

	UpgradeHeight int64
	DeployedErc20 DeployedErc20
}

// SetupSuite is run before all tests. It initializes chain connections and sets up the
// account used for funding accounts in the tests.
func (suite *E2eTestSuite) SetupSuite() {
	var err error
	fmt.Println("setting up test suite.")
	app.SetSDKConfig()

	suiteConfig := ParseSuiteConfig()
	suite.config = suiteConfig
	suite.DeployedErc20 = DeployedErc20{
		Address: common.HexToAddress(suiteConfig.KavaErc20Address),
		// Denom is fetched in InitKavaEvmData()
	}

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

// TearDownSuite is run after all tests have run.
// In the event of a panic during the tests, it is run after testify recovers.
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

// SetupKvtoolNodeRunner is a helper method for building a KvtoolRunnerConfig from the suite config.
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

// SetupLiveNetworkNodeRunner is a helper method for building a LiveNodeRunner from the suite config.
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

// SkipIfIbcDisabled should be called at the start of tests that require IBC.
// It gracefully skips the current test if IBC tests are disabled.
func (suite *E2eTestSuite) SkipIfIbcDisabled() {
	if !suite.config.IncludeIbcTests {
		suite.T().SkipNow()
	}
}

// SkipIfUpgradeDisabled should be called at the start of tests that require automated upgrades.
// It gracefully skips the current test if upgrades are dissabled.
// Note: automated upgrade tests are currently only enabled for Kvtool suite runs.
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
