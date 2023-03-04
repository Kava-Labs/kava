package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/app"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/kava-labs/kava/tests/util"
)

// Chain wraps query clients & accounts for a network
type Chain struct {
	encodingConfig kavaparams.EncodingConfig

	EvmClient *ethclient.Client
	Auth      authtypes.QueryClient
	Bank      banktypes.QueryClient
	Tm        tmservice.ServiceClient
	Tx        txtypes.ServiceClient
}

func NewChain(ports *runner.ChainPorts) (*Chain, error) {
	chain := &Chain{}
	chain.encodingConfig = app.MakeEncodingConfig()

	grpcUrl := fmt.Sprintf("http://localhost:%s", ports.GrpcPort)
	grpcConn, err := util.NewGrpcConnection(grpcUrl)
	if err != nil {
		return chain, err
	}

	evmRpcUrl := fmt.Sprintf("http://localhost:%s", ports.EvmPort)
	chain.EvmClient, err = ethclient.Dial(evmRpcUrl)
	if err != nil {
		return chain, err
	}

	chain.Auth = authtypes.NewQueryClient(grpcConn)
	chain.Bank = banktypes.NewQueryClient(grpcConn)
	chain.Tm = tmservice.NewServiceClient(grpcConn)
	chain.Tx = txtypes.NewServiceClient(grpcConn)

	return chain, nil
}
