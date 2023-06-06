package testutil

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/kava-labs/kava/app"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/kava-labs/kava/tests/util"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
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
	Bank      banktypes.QueryClient
	Committee committeetypes.QueryClient
	Community communitytypes.QueryClient
	Earn      earntypes.QueryClient
	Evm       evmtypes.QueryClient
	Evmutil   evmutiltypes.QueryClient
	Tm        tmservice.ServiceClient
	Tx        txtypes.ServiceClient
	Upgrade   upgradetypes.QueryClient
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
	chain.Bank = banktypes.NewQueryClient(grpcConn)
	chain.Committee = committeetypes.NewQueryClient(grpcConn)
	chain.Community = communitytypes.NewQueryClient(grpcConn)
	chain.Earn = earntypes.NewQueryClient(grpcConn)
	chain.Evm = evmtypes.NewQueryClient(grpcConn)
	chain.Evmutil = evmutiltypes.NewQueryClient(grpcConn)
	chain.Tm = tmservice.NewServiceClient(grpcConn)
	chain.Tx = txtypes.NewServiceClient(grpcConn)
	chain.Upgrade = upgradetypes.NewQueryClient(grpcConn)

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
	fmt.Printf("[%s] account used for funding (%s) address: %s\n", chain.ChainId, FundedAccountName, whale.SdkAddress)
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

// GetModuleBalances returns the balance of a requested module account
func (chain *Chain) GetModuleBalances(moduleName string) sdk.Coins {
	addr := authtypes.NewModuleAddress(moduleName)
	return chain.QuerySdkForBalances(addr)
}

func (chain *Chain) GetErc20Balance(contract, address common.Address) *big.Int {
	resData, err := chain.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contract,
		Data: util.BuildErc20BalanceOfCallData(address),
	}, nil)
	require.NoError(chain.t, err)

	return new(big.Int).SetBytes(resData)
}
