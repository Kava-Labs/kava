package testutil

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	emtests "github.com/evmos/ethermint/tests"
	emtypes "github.com/evmos/ethermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
)

// SigningAccount wraps details about an account and its private keys.
// It exposes functionality for signing and broadcasting transactions.
type SigningAccount struct {
	name     string
	mnemonic string

	evmPrivKey cryptotypes.PrivKey
	evmSigner  *util.EvmSigner
	evmReqChan chan<- util.EvmTxRequest
	evmResChan <-chan util.EvmTxResponse

	kavaSigner *util.KavaSigner
	sdkReqChan chan<- util.KavaMsgRequest
	sdkResChan <-chan util.KavaMsgResponse

	EvmAuth *bind.TransactOpts

	EvmAddress common.Address
	SdkAddress sdk.AccAddress

	gasDenom string

	l *log.Logger
}

// GetAccount returns the account with the given name or fails.
func (chain *Chain) GetAccount(name string) *SigningAccount {
	acc, found := chain.accounts[name]
	if !found {
		chain.t.Fatalf("failed to find account with name %s", name)
	}
	return acc
}

// AddNewSigningAccount sets up a new account with a signer for SDK and EVM transactions.
func (chain *Chain) AddNewSigningAccount(name string, hdPath *hd.BIP44Params, chainId, mnemonic string) *SigningAccount {
	if _, found := chain.accounts[name]; found {
		chain.t.Fatalf("account with name %s already exists", name)
	}

	// Kava signing account for SDK side
	privKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath.String())
	require.NoErrorf(chain.t, err, "failed to derive private key from mnemonic for %s: %s", name, err)
	privKey := &ethsecp256k1.PrivKey{Key: privKeyBytes}

	return chain.AddNewSigningAccountFromPrivKey(
		name,
		privKey,
		mnemonic,
		chainId,
	)
}

// AddNewSigningAccountFromPrivKey sets up a new account with a signer for SDK and EVM transactions,
// using the given private key.
func (chain *Chain) AddNewSigningAccountFromPrivKey(
	name string,
	privKey cryptotypes.PrivKey,
	mnemonic string, // optional
	chainId string,
) *SigningAccount {
	if _, found := chain.accounts[name]; found {
		chain.t.Fatalf("account with name %s already exists", name)
	}

	// Kava signing account for SDK side
	kavaSigner := util.NewKavaSigner(
		chainId,
		chain.EncodingConfig,
		chain.Auth,
		chain.Tx,
		privKey,
		100,
	)

	sdkReqChan := make(chan util.KavaMsgRequest)
	sdkResChan, err := kavaSigner.Run(sdkReqChan)
	require.NoErrorf(chain.t, err, "failed to start signer for account %s: %s", name, err)

	// Kava signing account for EVM side
	evmChainId, err := emtypes.ParseChainID(chainId)
	require.NoErrorf(chain.t, err, "unable to parse ethermint-compatible chain id from %s", chainId)
	ecdsaPrivKey, err := crypto.HexToECDSA(hex.EncodeToString(privKey.Bytes()))
	require.NoError(chain.t, err, "failed to generate ECDSA private key from bytes")

	evmSigner, err := util.NewEvmSigner(
		chain.EvmClient,
		ecdsaPrivKey,
		evmChainId,
	)
	require.NoErrorf(chain.t, err, "failed to create evm signer")

	evmReqChan := make(chan util.EvmTxRequest)
	evmResChan := evmSigner.Run(evmReqChan)

	logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", name), log.LstdFlags)

	chain.accounts[name] = &SigningAccount{
		name:     name,
		mnemonic: mnemonic,
		l:        logger,

		gasDenom: chain.StakingDenom,

		evmPrivKey: privKey,
		evmSigner:  evmSigner,
		evmReqChan: evmReqChan,
		evmResChan: evmResChan,

		kavaSigner: kavaSigner,
		sdkReqChan: sdkReqChan,
		sdkResChan: sdkResChan,

		EvmAuth: evmSigner.Auth,

		EvmAddress: evmSigner.Address(),
		SdkAddress: kavaSigner.Address(),
	}

	return chain.accounts[name]
}

// SignAndBroadcastKavaTx sends a request to the signer and awaits its response.
func (a *SigningAccount) SignAndBroadcastKavaTx(req util.KavaMsgRequest) util.KavaMsgResponse {
	a.l.Printf("broadcasting sdk tx. has data = %+v\n", req.Data)
	// send the request to signer
	a.sdkReqChan <- req

	// TODO: timeout awaiting the response.
	// block and await response
	// response is not returned until the msg is committed to a block
	res := <-a.sdkResChan

	// error will be set if response is not Code 0 (success) or Code 19 (already in mempool)
	if res.Err != nil {
		a.l.Printf("response code: %d error: %s\n", res.Result.Code, res.Result.RawLog)
	} else {
		a.l.Printf("response code: %d, hash %s\n", res.Result.Code, res.Result.TxHash)
	}

	return res
}

// EvmTxResponse is util.EvmTxResponse that also includes the Receipt, if available
type EvmTxResponse struct {
	util.EvmTxResponse
	Receipt *ethtypes.Receipt
}

// SignAndBroadcastEvmTx sends a request to the signer and awaits its response.
func (a *SigningAccount) SignAndBroadcastEvmTx(req util.EvmTxRequest) EvmTxResponse {
	a.l.Printf("broadcasting evm tx %+v\n", req.Data)
	// send the request to signer
	a.evmReqChan <- req

	// block and await response
	// response occurs once tx is submitted to pending tx pool.
	// poll for the receipt to wait for it to be included in a block
	res := <-a.evmResChan
	response := EvmTxResponse{
		EvmTxResponse: res,
	}
	// if failed during signing or broadcast, there will never be a receipt.
	if res.Err != nil {
		return response
	}

	// if we don't have a tx receipt within a given timeout, fail the request
	a.l.Printf("awaiting evm tx receipt for tx %s\n", res.TxHash)
	response.Receipt, response.Err = util.WaitForEvmTxReceipt(a.evmSigner.EvmClient, res.TxHash, 10*time.Second)

	return response
}

// SignRawEvmData signs raw evm data with the SigningAccount's private key.
// It does not broadcast the signed data.
func (a *SigningAccount) SignRawEvmData(msg []byte) ([]byte, types.PubKey, error) {
	keyringSigner := emtests.NewSigner(a.evmPrivKey)
	return keyringSigner.SignByAddress(a.SdkAddress, msg)
}

// NewFundedAccount creates a SigningAccount for a random account & funds the account from the whale.
func (chain *Chain) NewFundedAccount(name string, funds sdk.Coins) *SigningAccount {
	entropy, err := bip39.NewEntropy(128)
	require.NoErrorf(chain.t, err, "failed to generate entropy for account %s: %s", name, err)
	mnemonic, err := bip39.NewMnemonic(entropy)
	require.NoErrorf(chain.t, err, "failed to create new mnemonic for account %s: %s", name, err)

	acc := chain.AddNewSigningAccount(
		name,
		hd.CreateHDPath(app.Bip44CoinType, 0, 0),
		chain.ChainID,
		mnemonic,
	)

	// don't attempt to fund when no funds are desired
	if funds.IsZero() {
		return acc
	}

	whale := chain.GetAccount(FundedAccountName)

	// check that the whale has the necessary balance to fund account
	bal := chain.QuerySdkForBalances(whale.SdkAddress)
	require.Truef(chain.t,
		bal.IsAllGT(funds),
		"funded account lacks funds for account %s\nneeds: %s\nhas: %s", name, funds, bal,
	)

	whale.l.Printf("attempting to fund created account (%s=%s)\n", name, acc.SdkAddress.String())
	res := whale.BankSend(acc.SdkAddress, funds)

	require.NoErrorf(chain.t, res.Err, "failed to fund new account %s: %s", name, res.Err)

	whale.l.Printf("successfully funded [%s]\n", name)

	return acc
}

// GetNonce fetches the next nonce / sequence number for the account.
func (a *SigningAccount) NextNonce() (uint64, error) {
	return a.evmSigner.EvmClient.PendingNonceAt(context.Background(), a.EvmAddress)
}

// BankSend is a helper method for sending funds via x/bank's MsgSend
func (a *SigningAccount) BankSend(to sdk.AccAddress, amount sdk.Coins) util.KavaMsgResponse {
	return a.SignAndBroadcastKavaTx(
		util.KavaMsgRequest{
			Msgs:      []sdk.Msg{banktypes.NewMsgSend(a.SdkAddress, to, amount)},
			GasLimit:  2e5,                                                        // 200,000 gas
			FeeAmount: sdk.NewCoins(sdk.NewCoin(a.gasDenom, sdkmath.NewInt(200))), // assume min gas price of .001ukava
			Data:      fmt.Sprintf("sending %s to %s", amount, to),
		},
	)
}

// TransferErc20 is a helper method for sending an erc20 token
func (a *SigningAccount) TransferErc20(contract, to common.Address, amount *big.Int) (EvmTxResponse, error) {
	data := util.BuildErc20TransferCallData(to, amount)
	nonce, err := a.NextNonce()
	if err != nil {
		return EvmTxResponse{}, err
	}

	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(nonce, contract, big.NewInt(0), 1e5, big.NewInt(1e10), data),
		Data: fmt.Sprintf("fund %s with ERC20 balance (%s)", to.Hex(), amount.String()),
	}

	res := a.SignAndBroadcastEvmTx(req)
	return res, res.Err
}
