package testutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/app"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/kava-labs/kava/tests/util"
)

const (
	ChainId           = "kavalocalnet_8888-1"
	FundedAccountName = "whale"
	StakingDenom      = "ukava"
)

type E2eTestSuite struct {
	suite.Suite

	runner         runner.NodeRunner
	grpcConn       *grpc.ClientConn
	encodingConfig kavaparams.EncodingConfig

	EvmClient *ethclient.Client
	Auth      authtypes.QueryClient
	Bank      banktypes.QueryClient
	Tm        tmservice.ServiceClient
	Tx        txtypes.ServiceClient

	accounts map[string]*SigningAccount
}

func (suite *E2eTestSuite) SetupSuite() {
	app.SetSDKConfig()
	suite.encodingConfig = app.MakeEncodingConfig()

	// this mnemonic is expected to be a funded account that can seed the funds for all
	// new accounts created during tests. it will be available under Accounts["whale"]
	fundedAccountMnemonic := os.Getenv("E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC")
	if fundedAccountMnemonic == "" {
		suite.Fail("no E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC provided")
	}

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

	// setup an unauthenticated evm client
	evmRpcUrl := fmt.Sprintf("http://localhost:%s", config.KavaEvmPort)
	suite.EvmClient, err = ethclient.Dial(evmRpcUrl)
	if err != nil {
		suite.runner.Shutdown()
		suite.Fail("failed to connect to evm: %s", err)
	}

	// create grpc connection
	grpcUrl := fmt.Sprintf("http://localhost:%s", config.KavaGrpcPort)
	suite.grpcConn, err = util.NewGrpcConnection(grpcUrl)
	if err != nil {
		suite.runner.Shutdown()
		suite.Fail("failed to create grpc connection: %s", err)
	}

	// setup unauthenticated query clients for kava / cosmos
	suite.Auth = authtypes.NewQueryClient(suite.grpcConn)
	suite.Bank = banktypes.NewQueryClient(suite.grpcConn)
	suite.Tm = tmservice.NewServiceClient(suite.grpcConn)
	suite.Tx = txtypes.NewServiceClient(suite.grpcConn)

	// initialize accounts map
	suite.accounts = make(map[string]*SigningAccount)
	// setup the signing account for the initially funded account (used to fund all other accounts)
	suite.AddNewSigningAccount(
		FundedAccountName,
		hd.CreateHDPath(app.Bip44CoinType, 0, 0),
		ChainId,
		fundedAccountMnemonic,
	)
}

func (suite *E2eTestSuite) TearDownSuite() {
	// close all account request channels
	for _, a := range suite.accounts {
		close(a.requests)
	}
	// gracefully shutdown docker container(s)
	suite.runner.Shutdown()
}
