package util

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

var (
	ErrEvmBroadcastTimeout = errors.New("timed out waiting for tx to be committed to block")
	// ErrEvmTxFailed is returned when a tx is committed to a block, but the receipt status is 0.
	// this means the tx failed. we don't have debug_traceTransaction RPC command so the best way
	// to determine the problem is to attempt to make the tx manually.
	ErrEvmTxFailed = errors.New("transaction was committed but failed. likely an execution revert by contract code")
)

type EvmTxRequest struct {
	Tx   *ethtypes.Transaction
	Data interface{}
}
type EvmTxResponse struct {
	Request EvmTxRequest
	TxHash  common.Hash
	Err     error
}

type ErrEvmFailedToSign struct{ Err error }

func (e ErrEvmFailedToSign) Error() string {
	return fmt.Sprintf("failed to sign tx: %s", e.Err)
}

type ErrEvmFailedToBroadcast struct{ Err error }

func (e ErrEvmFailedToBroadcast) Error() string {
	return fmt.Sprintf("failed to broadcast tx: %s", e.Err)
}

// EvmSigner manages signing and broadcasting requests to transfer Erc20 tokens
// Will work for calling all contracts that have func signature `transfer(address,uint256)`
type EvmSigner struct {
	signerAddress common.Address
	Auth          *bind.TransactOpts
	EvmClient     *ethclient.Client
}

func NewEvmSigner(
	evmClient *ethclient.Client,
	privKey *ecdsa.PrivateKey,
	chainId *big.Int,
) (*EvmSigner, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	if err != nil {
		return &EvmSigner{}, err
	}

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return &EvmSigner{}, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	return &EvmSigner{
		Auth:          auth,
		signerAddress: crypto.PubkeyToAddress(*publicKeyECDSA),
		EvmClient:     evmClient,
	}, nil
}

func NewEvmSignerFromMnemonic(evmClient *ethclient.Client, evmChainId *big.Int, mnemonic string) (*EvmSigner, error) {
	hdPath := hd.CreateHDPath(60, 0, 0)
	privKeyBytes, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath.String())
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to derive private key from mnemonic")
	}
	privKey := &ethsecp256k1.PrivKey{Key: privKeyBytes}
	ecdsaPrivKey, err := crypto.HexToECDSA(hex.EncodeToString(privKey.Bytes()))
	if err != nil {
		return nil, err
	}

	return NewEvmSigner(evmClient, ecdsaPrivKey, evmChainId)
}

func (s *EvmSigner) Run(requests <-chan EvmTxRequest) <-chan EvmTxResponse {
	responses := make(chan EvmTxResponse)

	// receive tx requests, sign & broadcast them.
	// Responses are sent once the tx is added to the pending tx pool.
	// To see result, use TransactionReceipt after tx has been included in a block.
	go func() {
		for {
			// wait for incoming request
			req := <-requests

			signedTx, err := s.Auth.Signer(s.signerAddress, req.Tx)
			if err != nil {
				err = ErrEvmFailedToSign{Err: err}
			} else {
				err = s.EvmClient.SendTransaction(context.Background(), signedTx)
				if err != nil {
					err = ErrEvmFailedToBroadcast{Err: err}
				}
			}

			responses <- EvmTxResponse{
				Request: req,
				TxHash:  signedTx.Hash(),
				Err:     err,
			}
		}
	}()

	return responses
}

func (s *EvmSigner) Address() common.Address {
	return s.signerAddress
}

// WaitForEvmTxReceipt polls for a tx receipt and errors on timeout.
// If the receipt comes back, but with status 0 (failed), an error is returned.
func WaitForEvmTxReceipt(client *ethclient.Client, txHash common.Hash, timeout time.Duration) (*ethtypes.Receipt, error) {
	var receipt *ethtypes.Receipt
	var err error
	outOfTime := time.After(timeout)
	for {
		select {
		case <-outOfTime:
			err = ErrEvmBroadcastTimeout
		default:
			receipt, err = client.TransactionReceipt(context.Background(), txHash)
			if errors.Is(err, ethereum.NotFound) {
				// tx still not committed to a block. retry!
				time.Sleep(100 * time.Millisecond)
				continue
			}
			// a response status of 0 means the tx was successfully committed but failed to execute
			if receipt.Status == 0 {
				err = ErrEvmTxFailed
			}
		}
		break
	}
	return receipt, err
}
