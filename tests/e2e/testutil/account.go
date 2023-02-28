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
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	emtypes "github.com/tharsis/ethermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
)

var BroadcastTimeoutErr = errors.New("timed out waiting for tx to be committed to block")

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

// GetAccount returns the account with the given name or fails
func (suite *E2eTestSuite) GetAccount(name string) *SigningAccount {
	acc, found := suite.accounts[name]
	if !found {
		suite.Failf("account does not exist", "failed to find account with name %s", name)
	}
	return acc
}

// AddNewSigningAccount sets up a new account with a signer.
func (suite *E2eTestSuite) AddNewSigningAccount(name string, hdPath *hd.BIP44Params, chainId, mnemonic string) *SigningAccount {
	if _, found := suite.accounts[name]; found {
		suite.Failf("can't create signing account", "account with name %s already exists", name)
	}

	// Kava signing account for SDK side
	privKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath.String())
	suite.NoErrorf(err, "failed to derive private key from mnemonic for %s: %s", name, err)
	privKey := &ethsecp256k1.PrivKey{Key: privKeyBytes}

	kavaSigner := util.NewKavaSigner(
		chainId,
		suite.encodingConfig,
		suite.Auth,
		suite.Tx,
		privKey,
		100,
	)

	sdkReqChan := make(chan util.KavaMsgRequest)
	sdkResChan, err := kavaSigner.Run(sdkReqChan)
	suite.NoErrorf(err, "failed to start signer for account %s: %s", name, err)

	// Kava signing account for EVM side
	evmChainId, err := emtypes.ParseChainID(chainId)
	suite.NoErrorf(err, "unable to parse ethermint-compatible chain id from %s", chainId)
	ecdsaPrivKey, err := crypto.HexToECDSA(hex.EncodeToString(privKeyBytes))
	suite.NoError(err, "failed to generate ECDSA private key from bytes")

	evmSigner, err := util.NewEvmSigner(
		suite.EvmClient,
		ecdsaPrivKey,
		evmChainId,
	)
	suite.NoErrorf(err, "failed to create evm signer")

	evmReqChan := make(chan util.EvmTxRequest)
	evmResChan := evmSigner.Run(evmReqChan)

	logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", name), log.LstdFlags)

	suite.accounts[name] = &SigningAccount{
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

	return suite.accounts[name]
}

// SignAndBroadcastKavaTx sends a request to the signer and awaits its response.
func (a *SigningAccount) SignAndBroadcastKavaTx(req util.KavaMsgRequest) util.KavaMsgResponse {
	a.l.Printf("broadcasting sdk tx %+v\n", req.Data)
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
			response.Err = BroadcastTimeoutErr
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

func (suite *E2eTestSuite) NewFundedAccount(name string, funds sdk.Coins) *SigningAccount {
	entropy, err := bip39.NewEntropy(128)
	suite.NoErrorf(err, "failed to generate entropy for account %s: %s", name, err)
	mnemonic, err := bip39.NewMnemonic(entropy)
	suite.NoErrorf(err, "failed to create new mnemonic for account %s: %s", name, err)

	acc := suite.AddNewSigningAccount(
		name,
		hd.CreateHDPath(app.Bip44CoinType, 0, 0),
		ChainId,
		mnemonic,
	)

	whale := suite.GetAccount(FundedAccountName)
	whale.l.Printf("attempting to fund created account (%s=%s)\n", name, acc.SdkAddress.String())
	res := whale.SignAndBroadcastKavaTx(
		util.KavaMsgRequest{
			Msgs: []sdk.Msg{
				banktypes.NewMsgSend(whale.SdkAddress, acc.SdkAddress, funds),
			},
			GasLimit:  2e5,
			FeeAmount: sdk.NewCoins(sdk.NewCoin(StakingDenom, sdk.NewInt(75000))),
			Data:      fmt.Sprintf("initial funding of account %s", name),
		},
	)

	suite.NoErrorf(res.Err, "failed to fund new account %s: %s", name, res.Err)

	whale.l.Printf("successfully funded [%s]\n", name)

	return acc
}
