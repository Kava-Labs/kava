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
	ImageTag string

	KavaRpcPort  string
	KavaGrpcPort string
	KavaRestPort string
	KavaEvmPort  string
}

// NodeRunner is responsible for starting and managing docker containers to run a node.
type NodeRunner interface {
	StartChains()
	Shutdown()
}

// SingleKavaNodeRunner manages and runs a single Kava node.
type SingleKavaNodeRunner struct {
	config Config
}

var _ NodeRunner = &SingleKavaNodeRunner{}

func NewSingleKavaNode(config Config) *SingleKavaNodeRunner {
	return &SingleKavaNodeRunner{
		config: config,
	}
}

func (k *SingleKavaNodeRunner) StartChains() {
	installKvtoolCmd := exec.Command("./scripts/install-kvtool.sh")
	installKvtoolCmd.Stdout = os.Stdout
	installKvtoolCmd.Stderr = os.Stderr
	if err := installKvtoolCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to install kvtool: %s", err.Error()))
	}
	log.Println("starting kava node")
	startKavaCmd := exec.Command("kvtool", "testnet", "bootstrap")
	startKavaCmd.Stdout = os.Stdout
	startKavaCmd.Stderr = os.Stderr
	if err := startKavaCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to start kava: %s", err.Error()))
	}

	err := k.waitForChainStart()
	if err != nil {
		k.Shutdown()
		panic(err)
	}
	log.Println("kava is started!")
}

func (k *SingleKavaNodeRunner) Shutdown() {
	log.Println("shutting down kava node")
	shutdownKavaCmd := exec.Command("kvtool", "testnet", "down")
	shutdownKavaCmd.Stdout = os.Stdout
	shutdownKavaCmd.Stderr = os.Stderr
	if err := shutdownKavaCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to shutdown kvtool: %s", err.Error()))
	}
}

func (k *SingleKavaNodeRunner) waitForChainStart() error {
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

func (k *SingleKavaNodeRunner) pingKava() error {
	log.Println("pinging kava chain...")
	url := fmt.Sprintf("http://localhost:%s/status", k.config.KavaRpcPort)
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

func (k *SingleKavaNodeRunner) pingEvm() error {
	log.Println("pinging evm...")
	url := fmt.Sprintf("http://localhost:%s", k.config.KavaEvmPort)
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
