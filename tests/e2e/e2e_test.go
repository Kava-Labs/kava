package e2e_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/stretchr/testify/suite"
)

type SingleNodeE2eSuite struct {
	suite.Suite
	runner runner.NodeRunner
}

func (suite *SingleNodeE2eSuite) SetupSuite() {
	configDir, err := filepath.Abs("./generated/kava-1/config")
	if err != nil {
		panic(fmt.Sprintf("failed to get config dir: %s", err))
	}
	config := runner.Config{
		ConfigDir: configDir,

		KavaRpcPort:  "26657",
		KavaRestPort: "1317",
		EvmRpcPort:   "8545",

		ImageTag: "local",
	}
	suite.runner = runner.NewSingleKavaNode(config)
	suite.runner.StartChains()
}

func (suite *SingleNodeE2eSuite) TearDownSuite() {
	suite.runner.Shutdown()
}

func TestSingleNodeE2eSuite(t *testing.T) {
	suite.Run(t, new(SingleNodeE2eSuite))
}
