package testutil

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmhd "github.com/evmos/ethermint/crypto/hd"
	tmclient "github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/app"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/client/grpc"
	"github.com/kava-labs/kava/tests/e2e/runner"
	"github.com/kava-labs/kava/tests/util"
)

// Chain wraps query clients & accounts for a network
type Chain struct {
	accounts map[string]*SigningAccount
	t        *testing.T

	StakingDenom string
	ChainID      string
	Keyring      keyring.Keyring

	EvmClient     *ethclient.Client
	ContractAddrs map[string]common.Address
	erc20s        map[common.Address]struct{}

	EncodingConfig kavaparams.EncodingConfig

	TmSignClient tmclient.SignClient

	Grpc *grpc.KavaGrpcClient
}

// NewChain creates the query clients & signing account management for a chain run on a set of ports.
// A signing client for the fundedAccountMnemonic is initialized. This account is referred to in the
// code as "whale" and it is used to supply funds to all new accounts.
func NewChain(t *testing.T, details *runner.ChainDetails, fundedAccountMnemonic string) (*Chain, error) {
	chain := &Chain{
		t:             t,
		StakingDenom:  details.StakingDenom,
		ChainID:       details.ChainId,
		ContractAddrs: make(map[string]common.Address),
		erc20s:        make(map[common.Address]struct{}),
	}
	chain.EncodingConfig = app.MakeEncodingConfig()

	// setup keyring
	kr, err := keyring.New(
		sdk.KeyringServiceName(),
		keyring.BackendTest,
		util.KavaHomePath(),
		nil,
		chain.EncodingConfig.Marshaler,
		evmhd.EthSecp256k1Option(),
	)
	if err != nil {
		return nil, err
	}
	chain.Keyring = kr

	client, err := grpc.NewClient(details.GrpcUrl)
	if err != nil {
		chain.t.Fatalf("failed to create kava grpc client: %s", err)
	}
	chain.Grpc = client

	chain.EvmClient, err = details.EvmClient()
	if err != nil {
		return chain, err
	}

	chain.TmSignClient, err = details.RpcConn()
	if err != nil {
		return chain, err
	}

	// initialize accounts map
	chain.accounts = make(map[string]*SigningAccount)
	// setup the signing account for the initially funded account (used to fund all other accounts)
	whale := chain.AddNewSigningAccount(
		FundedAccountName,
		hd.CreateHDPath(Bip44CoinType, 0, 0),
		chain.ChainID,
		fundedAccountMnemonic,
	)

	// check that funded account is actually funded.
	fmt.Printf("[%s] account used for funding (%s) address: %s\n", chain.ChainID, FundedAccountName, whale.SdkAddress)
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

// ReturnAllFunds loops through all SigningAccounts and sends all their funds back to the
// initially funded account.
func (chain *Chain) ReturnAllFunds() {
	whale := chain.GetAccount(FundedAccountName)
	fmt.Println(chain.erc20s)
	for _, a := range chain.accounts {
		if a.SdkAddress.String() != whale.SdkAddress.String() {
			// NOTE: assumes all cosmos coin conversion funds have been converted back to sdk.

			// return all erc20 balance
			for erc20Addr := range chain.erc20s {
				erc20Bal := chain.GetErc20Balance(erc20Addr, a.EvmAddress)
				// if account has no balance, do nothing
				if erc20Bal.Cmp(big.NewInt(0)) == 0 {
					continue
				}
				_, err := a.TransferErc20(erc20Addr, whale.EvmAddress, erc20Bal)
				if err != nil {
					a.l.Printf("FAILED TO RETURN ERC20 FUNDS (contract: %s, balance: %d): %s\n",
						erc20Addr, erc20Bal, err,
					)
				}
			}

			// get sdk balance of account
			balance := chain.QuerySdkForBalances(a.SdkAddress)
			// assumes 200,000 gas w/ min fee of .001
			gas := sdk.NewInt64Coin(chain.StakingDenom, 200)

			// ensure they have enough gas to return funds
			if balance.AmountOf(chain.StakingDenom).LT(gas.Amount) {
				a.l.Printf("ACCOUNT LACKS GAS MONEY TO RETURN FUNDS: %s\n", balance)
				continue
			}

			// send it all back (minus gas) to the whale!
			res := a.BankSend(whale.SdkAddress, balance.Sub(gas))
			if res.Err != nil {
				a.l.Printf("failed to return funds: %s\n", res.Err)
			}
		}
	}
}

// RegisterErc20 is a method to record the address of erc20s on this chain.
// The full balances of each registered erc20 will be returned to the funded
// account when ReturnAllFunds is called.
func (chain *Chain) RegisterErc20(address common.Address) {
	chain.erc20s[address] = struct{}{}
}

// QuerySdkForBalances gets the balance of a particular address on this Chain.
func (chain *Chain) QuerySdkForBalances(addr sdk.AccAddress) sdk.Coins {
	res, err := chain.Grpc.Query.Bank.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
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

// GetErc20Balance fetches the ERC20 balance of `contract` for `address`.
func (chain *Chain) GetErc20Balance(contract, address common.Address) *big.Int {
	resData, err := chain.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contract,
		Data: util.BuildErc20BalanceOfCallData(address),
	}, nil)
	require.NoError(chain.t, err)

	return new(big.Int).SetBytes(resData)
}

func (chain *Chain) GetBeginBlockEventsFromQuery(
	ctx context.Context,
	query string,
) (sdk.StringEvents, int64, error) {
	// 1) Block search to find auction_start event and corresponding height
	// https://rpc.kava.io/block_search?query=%22auction_start.auction_id=16837%22

	blocks, err := chain.QueryBlock(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	if len(blocks) == 0 {
		return nil, 0, fmt.Errorf("no blocks found")
	}

	// 2) Block results to query events from height
	// https://rpc.kava.io/block_results?height=3146803
	events, err := chain.GetBeginBlockEvents(ctx, blocks[0].Block.Height)
	return events, blocks[0].Block.Height, err
}

func (chain *Chain) QueryBlock(ctx context.Context, query string) ([]*coretypes.ResultBlock, error) {
	page := 1
	perPage := 100

	res, err := chain.TmSignClient.BlockSearch(
		ctx,
		query,
		&page,
		&perPage,
		"desc",
	)

	if err != nil {
		return nil, fmt.Errorf("failed BlockSearch: %w", err)
	}

	return res.Blocks, nil
}

func (chain *Chain) GetBeginBlockEvents(ctx context.Context, height int64) (sdk.StringEvents, error) {
	res, err := chain.TmSignClient.BlockResults(
		ctx,
		&height,
	)

	if err != nil {
		return nil, fmt.Errorf("failed BlockResults: %w", err)
	}

	// Do not use sdk.StringifyEvents as it flattens events which makes it
	// more difficult to parse.
	strEvents := make(sdk.StringEvents, 0, len(res.BeginBlockEvents))
	for _, e := range res.BeginBlockEvents {
		strEvents = append(strEvents, sdk.StringifyEvent(e))
	}

	return strEvents, nil
}
