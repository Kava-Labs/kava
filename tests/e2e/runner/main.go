package runner

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// NodeRunner is responsible for starting and managing docker containers to run a node.
type NodeRunner interface {
	StartChains() Chains
	Shutdown()
}

func waitForChainStart(chainDetails ChainDetails) error {
	// exponential backoff on trying to ping the node, timeout after 30 seconds
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = 30 * time.Second
	if err := backoff.Retry(func() error { return pingKava(chainDetails.RpcUrl) }, b); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}

	b.Reset()
	// the evm takes a bit longer to start up. wait for it to start as well.
	if err := backoff.Retry(func() error { return pingEvm(chainDetails.EvmRpcUrl) }, b); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}
	return nil
}

func pingKava(rpcUrl string) error {
	log.Println("pinging kava chain...")
	statusUrl := fmt.Sprintf("%s/status", rpcUrl)
	res, err := http.Get(statusUrl)
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

func pingEvm(evmRpcUrl string) error {
	log.Println("pinging evm...")
	res, err := http.Get(evmRpcUrl)
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
