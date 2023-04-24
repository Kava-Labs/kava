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
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/kava-labs/kava/tests/util"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
)

// Chain wraps query clients & accounts for a network
type Chain struct {
	accounts map[string]*SigningAccount
	t        *testing.T

	StakingDenom string
	ChainId      string

	EvmClient     *ethclient.Client
	ContractAddrs map[string]common.Address

	EncodingConfig kavaparams.EncodingConfig

	Auth      authtypes.QueryClient
	Authz     authz.QueryClient
	Bank      banktypes.QueryClient
	Committee committeetypes.QueryClient
	Community communitytypes.QueryClient
	Earn      earntypes.QueryClient
	Evm       evmtypes.QueryClient
	Gov       govv1types.QueryClient
	Cdp       cdptypes.QueryClient
	Tm        tmservice.ServiceClient
	Tx        txtypes.ServiceClient
}

// NewChain creates the query clients & signing account management for a chain run on a set of ports.
// A signing client for the fundedAccountMnemonic is initialized. This account is referred to in the
// code as "whale" and it is used to supply funds to all new accounts.
func NewChain(t *testing.T, details *runner.ChainDetails, fundedAccountMnemonic string) (*Chain, error) {
	chain := &Chain{
		t:             t,
		StakingDenom:  details.StakingDenom,
		ChainId:       details.ChainId,
		ContractAddrs: make(map[string]common.Address),
	}
	chain.EncodingConfig = app.MakeEncodingConfig()

	grpcUrl := fmt.Sprintf("http://localhost:%s", details.GrpcPort)
	grpcConn, err := util.NewGrpcConnection(grpcUrl)
	if err != nil {
		return chain, err
	}

	evmRpcUrl := fmt.Sprintf("http://localhost:%s", details.EvmPort)
	chain.EvmClient, err = ethclient.Dial(evmRpcUrl)
	if err != nil {
		return chain, err
	}

	chain.Auth = authtypes.NewQueryClient(grpcConn)
	chain.Authz = authz.NewQueryClient(grpcConn)
	chain.Bank = banktypes.NewQueryClient(grpcConn)
	chain.Committee = committeetypes.NewQueryClient(grpcConn)
	chain.Community = communitytypes.NewQueryClient(grpcConn)
	chain.Earn = earntypes.NewQueryClient(grpcConn)
	chain.Evm = evmtypes.NewQueryClient(grpcConn)
	chain.Gov = govv1types.NewQueryClient(grpcConn)
	chain.Cdp = cdptypes.NewQueryClient(grpcConn)
	chain.Tm = tmservice.NewServiceClient(grpcConn)
	chain.Tx = txtypes.NewServiceClient(grpcConn)

	// initialize accounts map
	chain.accounts = make(map[string]*SigningAccount)
	// setup the signing account for the initially funded account (used to fund all other accounts)
	whale := chain.AddNewSigningAccount(
		FundedAccountName,
		hd.CreateHDPath(Bip44CoinType, 0, 0),
		chain.ChainId,
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

// Shutdown performs closes all the account request channels for this chain.
func (chain *Chain) Shutdown() {
	// close all account request channels
	for _, a := range chain.accounts {
		close(a.sdkReqChan)
	}
}

// QuerySdkForBalances gets the balance of a particular address on this Chain.
func (chain *Chain) QuerySdkForBalances(addr sdk.AccAddress) sdk.Coins {
	res, err := chain.Bank.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: addr.String(),
	})
	require.NoError(chain.t, err)
	return res.Balances
}

func (chain *Chain) QuerySdkForModuleAccount(moduleName string) authtypes.AccountI {
	res, err := chain.Auth.ModuleAccountByName(
		context.Background(),
		&authtypes.QueryModuleAccountByNameRequest{Name: moduleName},
	)
	require.NoError(chain.t, err)
	var account authtypes.AccountI
	err = chain.EncodingConfig.InterfaceRegistry.UnpackAny(res.Account, &account)
	require.NoError(chain.t, err)
	return account
}
