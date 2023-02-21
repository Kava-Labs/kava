package e2e_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kava-labs/kava/tests/e2e/runner"
)

func TestE2eSingleNode(t *testing.T) {
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

	chainRunner := runner.NewSingleKavaNode(config)
	chainRunner.StartChains()

	t.Log("running the chain!")

	chainRunner.Shutdown()
}
