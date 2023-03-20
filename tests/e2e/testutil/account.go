package testutil

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	emtypes "github.com/tharsis/ethermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
)

var ErrBroadcastTimeout = errors.New("timed out waiting for tx to be committed to block")

type SigningAccount struct {
	name     string
	mnemonic string

	evmSigner  *util.EvmSigner
	evmReqChan chan<- util.EvmTxRequest
	evmResChan <-chan util.EvmTxResponse

	kavaSigner *util.KavaSigner
	sdkReqChan chan<- util.KavaMsgRequest
	sdkResChan <-chan util.KavaMsgResponse

	EvmAddress common.Address
	SdkAddress sdk.AccAddress

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

	kavaSigner := util.NewKavaSigner(
		chainId,
		chain.encodingConfig,
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
	ecdsaPrivKey, err := crypto.HexToECDSA(hex.EncodeToString(privKeyBytes))
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

		evmSigner:  evmSigner,
		evmReqChan: evmReqChan,
		evmResChan: evmResChan,

		kavaSigner: kavaSigner,
		sdkReqChan: sdkReqChan,
		sdkResChan: sdkResChan,

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
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			response.Err = ErrBroadcastTimeout
		default:
			response.Receipt, response.Err = a.evmSigner.EvmClient.TransactionReceipt(context.Background(), res.TxHash)
			if errors.Is(response.Err, ethereum.NotFound) {
				// tx still not committed to a block. retry!
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		break
	}

	return response
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
		chain.ChainId,
		mnemonic,
	)

	// don't attempt to fund when no funds are desired
	if funds.IsZero() {
		return acc
	}

	whale := chain.GetAccount(FundedAccountName)
	whale.l.Printf("attempting to fund created account (%s=%s)\n", name, acc.SdkAddress.String())
	res := whale.SignAndBroadcastKavaTx(
		util.KavaMsgRequest{
			Msgs: []sdk.Msg{
				banktypes.NewMsgSend(whale.SdkAddress, acc.SdkAddress, funds),
			},
			GasLimit:  2e5,
			FeeAmount: sdk.NewCoins(sdk.NewCoin(chain.StakingDenom, sdk.NewInt(75000))),
			Data:      fmt.Sprintf("initial funding of account %s", name),
		},
	)

	require.NoErrorf(chain.t, res.Err, "failed to fund new account %s: %s", name, res.Err)

	whale.l.Printf("successfully funded [%s]\n", name)

	return acc
}
