package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type KvtoolRunnerConfig struct {
	KavaConfigTemplate string

	ImageTag   string
	IncludeIBC bool

	EnableAutomatedUpgrade  bool
	KavaUpgradeName         string
	KavaUpgradeHeight       int64
	KavaUpgradeBaseImageTag string

	SkipShutdown bool
}

// KvtoolRunner implements a NodeRunner that spins up local chains with kvtool.
// It has support for the following:
// - running a Kava node
// - optionally, running an IBC node with a channel opened to the Kava node
// - optionally, start the Kava node on one version and upgrade to another
type KvtoolRunner struct {
	config KvtoolRunnerConfig
}

var _ NodeRunner = &KvtoolRunner{}

// NewKvtoolRunner creates a new KvtoolRunner.
func NewKvtoolRunner(config KvtoolRunnerConfig) *KvtoolRunner {
	return &KvtoolRunner{
		config: config,
	}
}

// StartChains implements NodeRunner.
// For KvtoolRunner, it sets up, runs, and connects to a local chain via kvtool.
func (k *KvtoolRunner) StartChains() Chains {
	// install kvtool if not already installed
	installKvtoolCmd := exec.Command("./scripts/install-kvtool.sh")
	installKvtoolCmd.Stdout = os.Stdout
	installKvtoolCmd.Stderr = os.Stderr
	if err := installKvtoolCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to install kvtool: %s", err.Error()))
	}

	// start local test network with kvtool
	log.Println("starting kava node")
	kvtoolArgs := []string{"testnet", "bootstrap", "--kava.configTemplate", k.config.KavaConfigTemplate}
	// include an ibc chain if desired
	if k.config.IncludeIBC {
		kvtoolArgs = append(kvtoolArgs, "--ibc")
	}
	// handle automated upgrade functionality, if defined
	if k.config.EnableAutomatedUpgrade {
		kvtoolArgs = append(kvtoolArgs,
			"--upgrade-name", k.config.KavaUpgradeName,
			"--upgrade-height", fmt.Sprint(k.config.KavaUpgradeHeight),
			"--upgrade-base-image-tag", k.config.KavaUpgradeBaseImageTag,
		)
	}
	// start the chain
	startKavaCmd := exec.Command("kvtool", kvtoolArgs...)
	startKavaCmd.Env = os.Environ()
	startKavaCmd.Env = append(startKavaCmd.Env, fmt.Sprintf("KAVA_TAG=%s", k.config.ImageTag))
	startKavaCmd.Stdout = os.Stdout
	startKavaCmd.Stderr = os.Stderr
	log.Println(startKavaCmd.String())
	if err := startKavaCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to start kava: %s", err.Error()))
	}

	// wait for chain to be live.
	// if an upgrade is defined, this waits for the upgrade to be completed.
	if err := waitForChainStart(kvtoolKavaChain); err != nil {
		k.Shutdown()
		panic(err)
	}
	log.Println("kava is started!")

	chains := NewChains()
	chains.Register("kava", &kvtoolKavaChain)
	if k.config.IncludeIBC {
		chains.Register("ibc", &kvtoolIbcChain)
	}
	return chains
}

// Shutdown implements NodeRunner.
// For KvtoolRunner, it shuts down the local kvtool network.
// To prevent shutting down the chain (eg. to preserve logs or examine post-test state)
// use the `SkipShutdown` option on the config.
func (k *KvtoolRunner) Shutdown() {
	if k.config.SkipShutdown {
		log.Printf("would shut down but SkipShutdown is true")
		return
	}
	log.Println("shutting down kava node")
	shutdownKavaCmd := exec.Command("kvtool", "testnet", "down")
	shutdownKavaCmd.Stdout = os.Stdout
	shutdownKavaCmd.Stderr = os.Stderr
	if err := shutdownKavaCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to shutdown kvtool: %s", err.Error()))
	}
}
