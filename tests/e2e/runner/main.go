package runner

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type Config struct {
	ConfigDir string
	ImageTag  string

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

	pool     *dockertest.Pool
	resource *dockertest.Resource
}

var _ NodeRunner = &SingleKavaNodeRunner{}

func NewSingleKavaNode(config Config) *SingleKavaNodeRunner {
	return &SingleKavaNodeRunner{
		config: config,
	}
}

func (k *SingleKavaNodeRunner) StartChains() {
	log.Println("starting kava node")
	k.setupDockerPool()
	err := k.waitForChainStart()
	if err != nil {
		k.Shutdown()
		panic(err)
	}
}

func (k *SingleKavaNodeRunner) Shutdown() {
	log.Println("shutting down kava node")
	k.pool.Purge(k.resource)
}

func (k *SingleKavaNodeRunner) setupDockerPool() {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Failed to make docker pool: %s", err)
	}
	k.pool = pool

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to docker: %s", err)
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
			"9090",  // port inside container for Kava GRPC
			"8545",  // port inside container for EVM JSON-RPC
		},
		// expose the internal ports on the configured ports
		PortBindings: map[docker.Port][]docker.PortBinding{
			"26657": {{HostIP: "", HostPort: k.config.KavaRpcPort}},
			"1327":  {{HostIP: "", HostPort: k.config.KavaRestPort}},
			"9090":  {{HostIP: "", HostPort: k.config.KavaGrpcPort}},
			"8545":  {{HostIP: "", HostPort: k.config.KavaEvmPort}},
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Failed to start kava node: %s", err)
	}
	k.resource = resource
	log.Println(resource.GetHostPort("26657"))
}

func (k *SingleKavaNodeRunner) waitForChainStart() error {
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
