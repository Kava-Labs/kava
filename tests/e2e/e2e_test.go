package e2e_test

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/ethereum/go-ethereum/ethclient"

	kavagrpc "github.com/kava-labs/go-tools/grpc"
	"github.com/kava-labs/kava/tests/e2e/runner"
)

type E2eTestSuite struct {
	suite.Suite

	runner   runner.NodeRunner
	grpcConn *grpc.ClientConn

	EvmClient *ethclient.Client
	Tm        tmservice.ServiceClient
}

func (suite *E2eTestSuite) SetupSuite() {
	configDir, err := filepath.Abs("./generated/kava-1/config")
	if err != nil {
		suite.Fail("failed to get config dir: %s", err)
	}
	config := runner.Config{
		ConfigDir: configDir,

		KavaRpcPort:  "26657",
		KavaRestPort: "1317",
		KavaGrpcPort: "9090",
		KavaEvmPort:  "8545",

		ImageTag: "local",
	}
	suite.runner = runner.NewSingleKavaNode(config)
	suite.runner.StartChains()

	evmRpcUrl := fmt.Sprintf("http://localhost:%s", config.KavaEvmPort)
	suite.EvmClient, err = ethclient.Dial(evmRpcUrl)
	if err != nil {
		suite.runner.Shutdown()
		suite.Fail("failed to connect to evm: %s", err)
	}

	grpcUrl := fmt.Sprintf("http://localhost:%s", config.KavaGrpcPort)
	suite.grpcConn, err = kavagrpc.NewGrpcConnection(grpcUrl)
	if err != nil {
		suite.runner.Shutdown()
		suite.Fail("failed to create grpc connection: %s", err)
	}

	suite.Tm = tmservice.NewServiceClient(suite.grpcConn)
}

func (suite *E2eTestSuite) TearDownSuite() {
	suite.runner.Shutdown()
}

func TestE2eTestSuite(t *testing.T) {
	suite.Run(t, new(E2eTestSuite))
}

func (suite *E2eTestSuite) TestChainID() {
	// TODO: make chain agnostic, don't hardcode expected chain ids

	evmNetworkId, err := suite.EvmClient.NetworkID(context.Background())
	suite.NoError(err)
	suite.Equal(big.NewInt(8888), evmNetworkId)

	nodeInfo, err := suite.Tm.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	suite.NoError(err)
	suite.Equal("kavalocalnet_8888-1", nodeInfo.DefaultNodeInfo.Network)
}
