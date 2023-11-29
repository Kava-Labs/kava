package runner

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/client/grpc"
)

// LiveNodeRunnerConfig implements NodeRunner.
// It connects to a running network via the RPC, GRPC, and EVM urls.
type LiveNodeRunnerConfig struct {
	KavaRpcUrl    string
	KavaGrpcUrl   string
	KavaEvmRpcUrl string

	UpgradeHeight int64
}

// LiveNodeRunner implements NodeRunner for an already-running chain.
// If a LiveNodeRunner is used, end-to-end tests are run against a live chain.
type LiveNodeRunner struct {
	config LiveNodeRunnerConfig
}

var _ NodeRunner = LiveNodeRunner{}

// NewLiveNodeRunner creates a new LiveNodeRunner.
func NewLiveNodeRunner(config LiveNodeRunnerConfig) *LiveNodeRunner {
	return &LiveNodeRunner{config}
}

// StartChains implements NodeRunner.
// It initializes connections to the chain based on parameters.
// It attempts to ping the necessary endpoints and panics if they cannot be reached.
func (r LiveNodeRunner) StartChains() Chains {
	fmt.Println("establishing connection to live kava network")
	chains := NewChains()

	kavaChain := ChainDetails{
		RpcUrl:    r.config.KavaRpcUrl,
		GrpcUrl:   r.config.KavaGrpcUrl,
		EvmRpcUrl: r.config.KavaEvmRpcUrl,
	}

	if err := waitForChainStart(kavaChain); err != nil {
		panic(fmt.Sprintf("failed to ping chain: %s", err))
	}

	// determine chain id
	client, err := grpc.NewClient(kavaChain.GrpcUrl)
	if err != nil {
		panic(fmt.Sprintf("failed to create kava grpc client: %s", err))
	}

	nodeInfo, err := client.Query.Tm.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	if err != nil {
		panic(fmt.Sprintf("failed to fetch kava node info: %s", err))
	}
	kavaChain.ChainId = nodeInfo.DefaultNodeInfo.Network

	// determine staking denom
	stakingParams, err := client.Query.Staking.Params(context.Background(), &stakingtypes.QueryParamsRequest{})
	if err != nil {
		panic(fmt.Sprintf("failed to fetch kava staking params: %s", err))
	}
	kavaChain.StakingDenom = stakingParams.Params.BondDenom

	chains.Register("kava", &kavaChain)

	fmt.Printf("successfully connected to live network %+v\n", kavaChain)

	return chains
}

// Shutdown implements NodeRunner.
// As the chains are externally operated, this is a no-op.
func (LiveNodeRunner) Shutdown() {
	fmt.Println("shutting down e2e test connections.")
}
