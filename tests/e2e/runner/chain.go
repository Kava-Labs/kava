package runner

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"google.golang.org/grpc"

	"github.com/kava-labs/kava/tests/util"
)

var (
	ErrChainAlreadyExists = errors.New("chain already exists")
)

// ChainDetails wraps information about the ports exposed to the host that endpoints could be access on.
type ChainDetails struct {
	EvmRpcUrl string

	RpcPort  string
	GrpcPort string
	RestPort string

	ChainId      string
	StakingDenom string
}

func (c ChainDetails) EvmClient() (*ethclient.Client, error) {
	return ethclient.Dial(c.EvmRpcUrl)
}

func (c ChainDetails) GrpcConn() (*grpc.ClientConn, error) {
	grpcUrl := fmt.Sprintf("http://localhost:%s", c.GrpcPort)
	return util.NewGrpcConnection(grpcUrl)
}

type Chains struct {
	byName map[string]*ChainDetails
}

func NewChains() Chains {
	return Chains{byName: make(map[string]*ChainDetails, 0)}
}

func (c Chains) MustGetChain(name string) *ChainDetails {
	chain, found := c.byName[name]
	if !found {
		panic(fmt.Sprintf("no chain with name %s found", name))
	}
	return chain
}

func (c *Chains) Register(name string, chain *ChainDetails) error {
	if _, found := c.byName[name]; found {
		return ErrChainAlreadyExists
	}
	c.byName[name] = chain
	return nil
}

// the Chain details are all hardcoded because they are currently fixed by kvtool.
// some day they may be configurable, at which point `runner` can determine the ports
// and generate these details dynamically
var (
	kavaChain = ChainDetails{
		EvmRpcUrl: "http://localhost:8545",

		RpcPort:  "26657",
		RestPort: "1317",
		GrpcPort: "9090",

		ChainId:      "kavalocalnet_8888-1",
		StakingDenom: "ukava",
	}
	ibcChain = ChainDetails{
		EvmRpcUrl: "http://localhost:8547",

		RpcPort:  "26658",
		RestPort: "1318",
		GrpcPort: "9092",

		ChainId:      "kavalocalnet_8889-2",
		StakingDenom: "uatom",
	}
)
