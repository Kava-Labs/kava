package runner

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type Config struct {
	ConfigDir string
	ImageTag  string

	KavaRpcPort  string
	KavaRestPort string
	EvmRpcPort   string
}

type NodeRunner interface {
	StartChains()
	Shutdown()
}

type SingleKavaNodeSuite struct {
	config Config

	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func NewSingleKavaNode(config Config) *SingleKavaNodeSuite {
	return &SingleKavaNodeSuite{
		config: config,
	}
}

func (k *SingleKavaNodeSuite) StartChains() {
	fmt.Println("starting kava node")
	k.setupDockerPool()
	err := k.waitForChainStart()
	if err != nil {
		k.Shutdown()
		panic(err)
	}
}

func (k *SingleKavaNodeSuite) Shutdown() {
	fmt.Println("shutting down kava node")
	k.pool.Purge(k.resource)
}

func (k *SingleKavaNodeSuite) setupDockerPool() {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("Failed to make docker pool: %s", err))
	}
	k.pool = pool

	err = pool.Client.Ping()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to docker: %s", err))
	}

	resource, err := k.pool.RunWithOptions(&dockertest.RunOptions{
		Name:       "kavanode",
		Repository: "kava/kava",
		Tag:        k.config.ImageTag,
		Cmd: []string{
			"sh",
			"-c",
			"/root/.kava/config/init-data-directory.sh && kava start --rpc.laddr=tcp://0.0.0.0:26657",
		},
		Mounts: []string{
			fmt.Sprintf("%s:/root/.kava/config", k.config.ConfigDir),
		},
		ExposedPorts: []string{
			"26657", // port inside container for Kava RPC
			"1317",  // port inside container for Kava REST API
			"8545",  // port inside container for EVM JSON-RPC
		},
		// expose the internal ports on the configured ports
		PortBindings: map[docker.Port][]docker.PortBinding{
			"26657": {{HostIP: "", HostPort: k.config.KavaRpcPort}},
			"1327":  {{HostIP: "", HostPort: k.config.KavaRestPort}},
			"8545":  {{HostIP: "", HostPort: k.config.EvmRpcPort}},
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to start kava node: %s", err))
	}
	k.resource = resource
}

func (k *SingleKavaNodeSuite) waitForChainStart() error {
	// exponential backoff on trying to ping the node, timeout after 30 seconds
	k.pool.MaxWait = 30 * time.Second
	if err := k.pool.Retry(k.pingKava); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}
	// the evm takes a bit longer to start up. wait for it to start as well.
	if err := k.pool.Retry(k.pingEvm); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}
	return nil
}

func (k *SingleKavaNodeSuite) pingKava() error {
	fmt.Println("pinging kava chain...")
	url := fmt.Sprintf("http://localhost:%s/status", k.config.KavaRpcPort)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("ping to status failed: %d", res.StatusCode)
	}
	fmt.Println("successfully started Kava!")
	return nil
}

func (k *SingleKavaNodeSuite) pingEvm() error {
	fmt.Println("pinging evm...")
	url := fmt.Sprintf("http://localhost:%s", k.config.EvmRpcPort)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// when running, it should respond 405 to a GET request
	if res.StatusCode != 405 {
		return fmt.Errorf("ping to evm failed: %d", res.StatusCode)
	}
	fmt.Println("successfully pinged EVM!")
	return nil
}
