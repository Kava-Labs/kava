package runner

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	rpchttpclient "github.com/tendermint/tendermint/rpc/client/http"
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

// EvmClient dials the underlying EVM RPC url and returns an ethclient.
func (c ChainDetails) EvmClient() (*ethclient.Client, error) {
	return ethclient.Dial(c.EvmRpcUrl)
}

// GrpcConn creates a new connection to the underlying Grpc url.
func (c ChainDetails) GrpcConn() (*grpc.ClientConn, error) {
	return util.NewGrpcConnection(c.GrpcUrl)
}

// RpcConn creates a new connection to the underlying Rpc url.
func (c ChainDetails) RpcConn() (*rpchttpclient.HTTP, error) {
	return rpchttpclient.New(c.RpcUrl, "/websocket")
}

// Chains wraps a map of name -> details about how to connect to a chain.
// It prevents registering multiple chains with the same name & encapsulates
// panicking if attempting to access a chain that does not exist.
type Chains struct {
	byName map[string]*ChainDetails
}

// NewChains creates an empty Chains map.
func NewChains() Chains {
	return Chains{byName: make(map[string]*ChainDetails, 0)}
}

// MustGetChain returns the chain of a given name,
// or panics if a chain with that name has not been registered.
func (c Chains) MustGetChain(name string) *ChainDetails {
	chain, found := c.byName[name]
	if !found {
		panic(fmt.Sprintf("no chain with name %s found", name))
	}
	return chain
}

// Register adds a chain to the map.
// It returns an error if a ChainDetails with that name has already been registered.
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
