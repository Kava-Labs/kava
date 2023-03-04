package testutil

import (
	"context"
	"fmt"
	"os"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/app"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/tests/e2e/runner"
)

const (
	ChainId           = "kavalocalnet_8888-1"
	FundedAccountName = "whale"
	StakingDenom      = "ukava"
	// use coin type 60 so we are compatible with accounts from `kava add keys --eth <name>`
	// these accounts use the ethsecp256k1 signing algorithm that allows the signing client
	// to manage both sdk & evm txs.
	Bip44CoinType = 60
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

	Chains runner.Chains

	accounts map[string]*SigningAccount
}

func (suite *E2eTestSuite) SetupSuite() {
	var err error
	fmt.Println("setting up test suite.")
	app.SetSDKConfig()
	suite.encodingConfig = app.MakeEncodingConfig()

	// this mnemonic is expected to be a funded account that can seed the funds for all
	// new accounts created during tests. it will be available under Accounts["whale"]
	fundedAccountMnemonic := os.Getenv("E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC")
	if fundedAccountMnemonic == "" {
		suite.Fail("no E2E_KAVA_FUNDED_ACCOUNT_MNEMONIC provided")
	}

	config := runner.Config{
		IncludeIBC: true, // TODO: env var that toggles IBC tests.
		ImageTag:   "local",
	}
	suite.runner = runner.NewKavaNode(config)
	suite.Chains = suite.runner.StartChains()

	kavachain := suite.Chains.MustGetChain("kava")

	// setup an unauthenticated evm client
	suite.EvmClient, err = kavachain.EvmClient()
	if err != nil {
		suite.runner.Shutdown()
		suite.Fail("failed to connect to evm: %s", err)
	}

	// create grpc connection
	suite.grpcConn, err = kavachain.GrpcConn()
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
	whale := suite.AddNewSigningAccount(
		FundedAccountName,
		hd.CreateHDPath(Bip44CoinType, 0, 0),
		ChainId,
		fundedAccountMnemonic,
	)

	// check that funded account is actually funded.
	fmt.Printf("account used for funding (%s) address: %s\n", FundedAccountName, whale.SdkAddress)
	whaleFunds := suite.QuerySdkForBalances(whale.SdkAddress)
	if whaleFunds.IsZero() {
		suite.FailNow("no available funds.", "funded account mnemonic is for account with no funds")
	}
}

func (suite *E2eTestSuite) TearDownSuite() {
	fmt.Println("tearing down test suite.")
	// close all account request channels
	for _, a := range suite.accounts {
		close(a.sdkReqChan)
	}
	// gracefully shutdown docker container(s)
	suite.runner.Shutdown()
}

func (suite *E2eTestSuite) QuerySdkForBalances(addr sdk.AccAddress) sdk.Coins {
	res, err := suite.Bank.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: addr.String(),
	})
	suite.NoError(err)
	return res.Balances
}
