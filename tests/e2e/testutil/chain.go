package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/kava-labs/kava/tests/util"
)

// Chain wraps query clients & accounts for a network
type Chain struct {
	encodingConfig kavaparams.EncodingConfig
	accounts       map[string]*SigningAccount
	t              *testing.T

	EvmClient *ethclient.Client
	Auth      authtypes.QueryClient
	Bank      banktypes.QueryClient
	Tm        tmservice.ServiceClient
	Tx        txtypes.ServiceClient
}

func NewChain(t *testing.T, ports *runner.ChainPorts, fundedAccountMnemonic string) (*Chain, error) {
	chain := &Chain{t: t}
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

	// initialize accounts map
	chain.accounts = make(map[string]*SigningAccount)
	// setup the signing account for the initially funded account (used to fund all other accounts)
	whale := chain.AddNewSigningAccount(
		FundedAccountName,
		hd.CreateHDPath(Bip44CoinType, 0, 0),
		ChainId,
		fundedAccountMnemonic,
	)

	// check that funded account is actually funded.
	fmt.Printf("account used for funding (%s) address: %s\n", FundedAccountName, whale.SdkAddress)
	whaleFunds := chain.QuerySdkForBalances(whale.SdkAddress)
	if whaleFunds.IsZero() {
		chain.t.Fatal("funded account mnemonic is for account with no funds")
	}

	return chain, nil
}

func (chain *Chain) Shutdown() {
	// close all account request channels
	for _, a := range chain.accounts {
		close(a.sdkReqChan)
	}
}

func (chain *Chain) QuerySdkForBalances(addr sdk.AccAddress) sdk.Coins {
	res, err := chain.Bank.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: addr.String(),
	})
	require.NoError(chain.t, err)
	return res.Balances
}
