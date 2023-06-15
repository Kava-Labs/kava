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

// ChainDetails wraps information about the properties & endpoints of a chain.
type ChainDetails struct {
	RpcUrl    string
	GrpcUrl   string
	EvmRpcUrl string

	ChainId      string
	StakingDenom string
}

func (c ChainDetails) EvmClient() (*ethclient.Client, error) {
	return ethclient.Dial(c.EvmRpcUrl)
}

func (c ChainDetails) GrpcConn() (*grpc.ClientConn, error) {
	return util.NewGrpcConnection(c.GrpcUrl)
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
// someday they may be accepted as configurable parameters.
var (
	kvtoolKavaChain = ChainDetails{
		RpcUrl:    "http://localhost:26657",
		GrpcUrl:   "http://localhost:9090",
		EvmRpcUrl: "http://localhost:8545",

		ChainId:      "kavalocalnet_8888-1",
		StakingDenom: "ukava",
	}
	kvtoolIbcChain = ChainDetails{
		RpcUrl:    "http://localhost:26658",
		GrpcUrl:   "http://localhost:9092",
		EvmRpcUrl: "http://localhost:8547",

		ChainId:      "kavalocalnet_8889-2",
		StakingDenom: "uatom",
	}
)
