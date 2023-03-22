package runner

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type Config struct {
	KavaConfigTemplate string

	ImageTag   string
	IncludeIBC bool

	EnableAutomatedUpgrade  bool
	KavaUpgradeName         string
	KavaUpgradeHeight       int64
	KavaUpgradeBaseImageTag string

	SkipShutdown bool
}

// NodeRunner is responsible for starting and managing docker containers to run a node.
type NodeRunner interface {
	StartChains() Chains
	Shutdown()
}

// KavaNodeRunner manages and runs a single Kava node.
type KavaNodeRunner struct {
	config    Config
	kavaChain *ChainDetails
}

var _ NodeRunner = &KavaNodeRunner{}

func NewKavaNode(config Config) *KavaNodeRunner {
	return &KavaNodeRunner{
		config: config,
	}
}

func (k *KavaNodeRunner) StartChains() Chains {
	installKvtoolCmd := exec.Command("./scripts/install-kvtool.sh")
	installKvtoolCmd.Stdout = os.Stdout
	installKvtoolCmd.Stderr = os.Stderr
	if err := installKvtoolCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to install kvtool: %s", err.Error()))
	}

	log.Println("starting kava node")
	kvtoolArgs := []string{"testnet", "bootstrap", "--kava.configTemplate", k.config.KavaConfigTemplate}
	if k.config.IncludeIBC {
		kvtoolArgs = append(kvtoolArgs, "--ibc")
	}
	if k.config.EnableAutomatedUpgrade {
		kvtoolArgs = append(kvtoolArgs,
			"--upgrade-name", k.config.KavaUpgradeName,
			"--upgrade-height", fmt.Sprint(k.config.KavaUpgradeHeight),
			"--upgrade-base-image-tag", k.config.KavaUpgradeBaseImageTag,
		)
	}
	startKavaCmd := exec.Command("kvtool", kvtoolArgs...)
	startKavaCmd.Env = os.Environ()
	startKavaCmd.Env = append(startKavaCmd.Env, fmt.Sprintf("KAVA_TAG=%s", k.config.ImageTag))
	startKavaCmd.Stdout = os.Stdout
	startKavaCmd.Stderr = os.Stderr
	log.Println(startKavaCmd.String())
	if err := startKavaCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to start kava: %s", err.Error()))
	}

	k.kavaChain = &kavaChain

	err := k.waitForChainStart()
	if err != nil {
		k.Shutdown()
		panic(err)
	}
	log.Println("kava is started!")

	chains := NewChains()
	chains.Register("kava", k.kavaChain)
	if k.config.IncludeIBC {
		chains.Register("ibc", &ibcChain)
	}
	return chains
}

func (k *KavaNodeRunner) Shutdown() {
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

func (k *KavaNodeRunner) waitForChainStart() error {
	// exponential backoff on trying to ping the node, timeout after 30 seconds
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = 30 * time.Second
	if err := backoff.Retry(k.pingKava, b); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}
	b.Reset()
	// the evm takes a bit longer to start up. wait for it to start as well.
	if err := backoff.Retry(k.pingEvm, b); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}
	return nil
}

func (k *KavaNodeRunner) pingKava() error {
	log.Println("pinging kava chain...")
	url := fmt.Sprintf("http://localhost:%s/status", k.kavaChain.RpcPort)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("ping to status failed: %d", res.StatusCode)
	}
	log.Println("successfully started Kava!")
	return nil
}

func (k *KavaNodeRunner) pingEvm() error {
	log.Println("pinging evm...")
	url := fmt.Sprintf("http://localhost:%s", k.kavaChain.EvmPort)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// when running, it should respond 405 to a GET request
	if res.StatusCode != 405 {
		return fmt.Errorf("ping to evm failed: %d", res.StatusCode)
	}
	log.Println("successfully pinged EVM!")
	return nil
}
