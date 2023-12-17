package util

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kava-labs/kava/app/params"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	tmmempool "github.com/cometbft/cometbft/mempool"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	ErrSdkBroadcastTimeout = errors.New("timed out waiting for tx to be committed to block")
	ErrUnsuccessfulTx      = errors.New("tx committed but returned nonzero code")
)

type KavaMsgRequest struct {
	Msgs      []sdk.Msg
	GasLimit  uint64
	FeeAmount sdk.Coins
	Memo      string
	// Arbitrary data to be referenced in the corresponding KavaMsgResponse, unused
	// in signing. This is mostly useful to match KavaMsgResponses with KavaMsgRequests.
	Data interface{}
}

type KavaMsgResponse struct {
	Request KavaMsgRequest
	Tx      authsigning.Tx
	TxBytes []byte
	Result  sdk.TxResponse
	Err     error
}

// internal result for inner loop logic
type broadcastTxResult int

const (
	txOK broadcastTxResult = iota
	txFailed
	txRetry
	txResetSequence
)

// KavaSigner broadcasts msgs to a single kava node
type KavaSigner struct {
	chainID         string
	encodingConfig  params.EncodingConfig
	authClient      authtypes.QueryClient
	txClient        txtypes.ServiceClient
	privKey         cryptotypes.PrivKey
	inflightTxLimit uint64
}

func NewKavaSigner(
	chainID string,
	encodingConfig params.EncodingConfig,
	authClient authtypes.QueryClient,
	txClient txtypes.ServiceClient,
	privKey cryptotypes.PrivKey,
	inflightTxLimit uint64) *KavaSigner {

	return &KavaSigner{
		chainID:         chainID,
		encodingConfig:  encodingConfig,
		authClient:      authClient,
		txClient:        txClient,
		privKey:         privKey,
		inflightTxLimit: inflightTxLimit,
	}
}

func (s *KavaSigner) pollAccountState() <-chan authtypes.AccountI {
	accountState := make(chan authtypes.AccountI)

	go func() {
		for {
			request := authtypes.QueryAccountRequest{
				Address: GetAccAddress(s.privKey).String(),
			}
			response, err := s.authClient.Account(context.Background(), &request)

			if err == nil {
				var account authtypes.AccountI

				err = s.encodingConfig.InterfaceRegistry.UnpackAny(response.Account, &account)
				if err == nil {
					accountState <- account
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return accountState
}

func (s *KavaSigner) Run(requests <-chan KavaMsgRequest) (<-chan KavaMsgResponse, error) {
	// poll account state in it's own goroutine
	// and send status updates to the signing goroutine
	//
	// TODO: instead of polling, we can wait for block
	// websocket events with a fallback to polling
	accountState := s.pollAccountState()

	responses := make(chan KavaMsgResponse)
	go func() {
		// wait until account is loaded to start signing
		account := <-accountState
		// store current request waiting to be broadcasted
		var currentRequest *KavaMsgRequest
		// keep track of all successfully broadcasted txs
		// index is sequence % inflightTxLimit
		inflight := make([]*KavaMsgResponse, s.inflightTxLimit)
		// used for confirming sent txs only
		prevDeliverTxSeq := account.GetSequence()
		// tx sequence of already signed messages
		checkTxSeq := account.GetSequence()
		// tx sequence of broadcast queue, is reset upon
		// unauthorized errors to recheck/refill mempool
		broadcastTxSeq := account.GetSequence()

		for {
			// the inflight limit includes the current request
			//
			// account.GetSequence() represents the first tx in the mempool (at the last known state)
			// or the next sequence to sign with if checkTxSeq == account.GetSequence() (zero msgs in flight)
			//
			// checkTxSeq always represents the next available mempool sequence to sign with
			//
			// if currentRequest is nil, then it will be used for the next request received
			// if currentRequest is not nil, then checkTxSeq will be used to sign that request
			//
			// therefore, assuming no errors, broadcastTxSeq will be checkTxSeq-1 or checkTxSeq, dependent on
			// if the currentRequest has been been successfully broadcast
			//
			// if an unauthorized error occurs, a tx in the mempool was dropped (or mempool flushed, node restart, etc)
			// and broadcastTxSeq is reset to account.GetSequence() in order to refil the mempool and ensure
			// checkTxSeq is valid
			//
			// if an authorized error occurs due to another process signing messages on behalf of the same
			// address, then broadcastTxSeq will continually be reset until that sequence is delivered to a block
			//
			// this results in the message we signed with the same sequence being skipped as well as
			// draining our inflight messages to 0.
			//
			// On deployments, a similar event will occur. we will continually broadcast until
			// all of the previous transactions are processed and out of the mempool.
			//
			// it's possible to increase the checkTx (up to the inflight limit) until met with a successful broadcast,
			// to fill the mempool faster, but this feature is umimplemented and would be best enabled only once
			// on startup.  An authorized error during normal operation would be difficult or impossible to tell apart
			// from a dropped mempool tx (without further improving mempool queries).  Options such as persisting inflight
			// state out of process may be better.
			inflightLimitReached := checkTxSeq-account.GetSequence() >= s.inflightTxLimit

			// if we are still processing a request or the inflight limit is reached
			// then block until the next account update without accepting new requests
			if currentRequest != nil || inflightLimitReached {
				account = <-accountState
			} else {
				// block on state update or new requests
				select {
				case account = <-accountState:
				case request := <-requests:
					currentRequest = &request
				}
			}

			// send delivered (included in block) responses to caller
			if account.GetSequence() > prevDeliverTxSeq {
				for i := prevDeliverTxSeq; i < account.GetSequence(); i++ {
					response := inflight[i%s.inflightTxLimit]
					// sequences may be skipped due to errors
					if response != nil {
						responses <- *response
					}
					// clear to prevent duplicate confirmations on errors
					inflight[i%s.inflightTxLimit] = nil
				}
				prevDeliverTxSeq = account.GetSequence()
			}

			// recover from errors due to untracked messages in mempool
			// this will happen on deploys, or if another process
			// signs a tx using the same address
			if checkTxSeq < account.GetSequence() {
				checkTxSeq = account.GetSequence()
			}

			// if currentRequest then lastRequestTxSeq == checkTxSeq
			// if not currentRequest then lastRequestTxSeq == checkTxSeq - 1
			lastRequestTxSeq := checkTxSeq
			if currentRequest == nil && lastRequestTxSeq > 0 {
				lastRequestTxSeq--
			}
			// reset broadcast seq if iterated over last request seq
			// we always want to broadcast the current or last request
			// to heartbeat the mempool
			if broadcastTxSeq > lastRequestTxSeq {
				broadcastTxSeq = lastRequestTxSeq
			}

			// loop serves three purposes
			//   - recover from dropped txs (broadcastTxSeq < lastRequestTxSeq)
			//   - send new requests (currentRequest is set)
			//   - send mempool heartbeat (currentRequest is nil)
		BROADCAST_LOOP:
			for broadcastTxSeq <= lastRequestTxSeq {

				// we have a new request that has not been successfully broadcasted
				// and are at the last broadcastTxSeq (broadcastTxSeq == checkTxSeq in this case)
				sendingCurrentRequest := broadcastTxSeq == lastRequestTxSeq && currentRequest != nil

				// check if we have a previous response to check/retry/send for the broadcastTxSeq
				response := inflight[broadcastTxSeq%s.inflightTxLimit]

				// no response -- either checkTxSeq was skipped (untracked mempool tx), or
				// we are signing a new transactions (currentRequest is not nil)
				if response == nil {
					// nothing to do if no response to retry and not sending a current request
					if !sendingCurrentRequest {
						// move onto next broadcastTxSeq or exit loop
						broadcastTxSeq++
						continue
					}

					txBuilder := s.encodingConfig.TxConfig.NewTxBuilder()
					txBuilder.SetMsgs(currentRequest.Msgs...)
					txBuilder.SetGasLimit(currentRequest.GasLimit)
					txBuilder.SetFeeAmount(currentRequest.FeeAmount)

					signerData := authsigning.SignerData{
						ChainID:       s.chainID,
						AccountNumber: account.GetAccountNumber(),
						Sequence:      broadcastTxSeq,
					}

					tx, txBytes, err := Sign(s.encodingConfig.TxConfig, s.privKey, txBuilder, signerData)

					response = &KavaMsgResponse{
						Request: *currentRequest,
						Tx:      tx,
						TxBytes: txBytes,
						Err:     err,
					}

					// could not sign and encode the currentRequest
					if response.Err != nil {
						// clear invalid request, since this is non-recoverable
						currentRequest = nil

						// response immediately with error
						responses <- *response

						// exit loop
						broadcastTxSeq++
						continue
					}
				}

				// broadcast tx and get result
				//
				// there are four main types of results
				//
				// OK (tx in mempool, store response - add to inflight txs)
				// Retry (tx not in mempool, but retry - do not change inflight status)
				// Failed (tx not in mempool, not recoverable - clear inflight status, reply to channel)
				// Unauthorized (tx not in mempool - sequence not valid)
				broadcastRequest := txtypes.BroadcastTxRequest{
					TxBytes: response.TxBytes,
					Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
				}
				broadcastResponse, err := s.txClient.BroadcastTx(context.Background(), &broadcastRequest)

				// set to determine action at the end of loop
				// default is OK
				txResult := txOK

				// determine action to take when err (and no response)
				if err != nil {
					if tmmempool.IsPreCheckError(err) {
						// ErrPreCheck - not recoverable
						response.Err = err
						txResult = txFailed
					} else {
						// could not contact node (POST failed, dns errors, etc)
						// exit loop, wait for another account state update
						// TODO: are there cases here that we will never recover from?
						// should we implement retry limit?
						response.Err = err
						txResult = txRetry
					}
				} else {
					// store rpc result in response
					response.Result = *broadcastResponse.TxResponse

					// determine action to take based on rpc result
					switch response.Result.Code {
					// 0: success, in mempool
					case errorsmod.SuccessABCICode:
						txResult = txOK
					// 4: unauthorized
					case sdkerrors.ErrUnauthorized.ABCICode():
						txResult = txResetSequence
					// 19: success, tx already in mempool
					case sdkerrors.ErrTxInMempoolCache.ABCICode():
						txResult = txOK
					// 20: mempool full
					case sdkerrors.ErrMempoolIsFull.ABCICode():
						txResult = txRetry
					// 32: wrong sequence
					case sdkerrors.ErrWrongSequence.ABCICode():
						txResult = txResetSequence
					default:
						response.Err = fmt.Errorf("message failed to broadcast, unrecoverable error code %d", response.Result.Code)
						txResult = txFailed
					}
				}

				switch txResult {
				case txOK:
					// clear any errors from previous attempts
					response.Err = nil

					// store for delivery later
					inflight[broadcastTxSeq%s.inflightTxLimit] = response

					// if this is the current/last request, then clear
					// the request and increment the checkTxSeq
					if sendingCurrentRequest {
						currentRequest = nil
						checkTxSeq++
					}

					// go to next request
					broadcastTxSeq++
				case txFailed:
					// do not store the request as inflight (it's not in the mempool)
					inflight[broadcastTxSeq%s.inflightTxLimit] = nil

					// clear current request if it failed
					if sendingCurrentRequest {
						currentRequest = nil
					}

					// immediatley response to channel
					responses <- *response
					// go to next request
					broadcastTxSeq++
				case txRetry:
					break BROADCAST_LOOP
				case txResetSequence:
					broadcastTxSeq = account.GetSequence()
					break BROADCAST_LOOP
				}
			}
		}
	}()

	return responses, nil
}

// Address returns the address of the Signer
func (s *KavaSigner) Address() sdk.AccAddress {
	return GetAccAddress(s.privKey)
}

// Sign signs a populated TxBuilder and returns a signed Tx and raw transaction bytes
func Sign(
	txConfig sdkclient.TxConfig,
	privKey cryptotypes.PrivKey,
	txBuilder sdkclient.TxBuilder,
	signerData authsigning.SignerData,
) (authsigning.Tx, []byte, error) {
	signatureData := signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sigV2 := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     &signatureData,
		Sequence: signerData.Sequence,
	}
	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return txBuilder.GetTx(), nil, err
	}

	signBytes, err := txConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx())
	if err != nil {
		return txBuilder.GetTx(), nil, err
	}
	signature, err := privKey.Sign(signBytes)
	if err != nil {
		return txBuilder.GetTx(), nil, err
	}

	sigV2.Data = &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: signature,
	}
	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return txBuilder.GetTx(), nil, err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return txBuilder.GetTx(), nil, err
	}

	return txBuilder.GetTx(), txBytes, nil
}

func GetAccAddress(privKey cryptotypes.PrivKey) sdk.AccAddress {
	return privKey.PubKey().Address().Bytes()
}

// WaitForSdkTxCommit polls the chain until the tx hash is found or times out.
// Returns an error immediately if tx hash is empty
func WaitForSdkTxCommit(txClient txtypes.ServiceClient, txHash string, timeout time.Duration) (*sdk.TxResponse, error) {
	if txHash == "" {
		return nil, fmt.Errorf("tx hash is empty")
	}
	var err error
	var txRes *sdk.TxResponse
	var res *txtypes.GetTxResponse
	outOfTime := time.After(timeout)
	for {
		select {
		case <-outOfTime:
			err = ErrSdkBroadcastTimeout
		default:
			res, err = txClient.GetTx(context.Background(), &txtypes.GetTxRequest{Hash: txHash})
			if err != nil {
				status, ok := grpcstatus.FromError(err)
				if ok && status.Code() == codes.NotFound {
					// tx still not committed to a block. retry!
					time.Sleep(100 * time.Millisecond)
					continue
				}
				break
			}
			txRes = res.TxResponse
			if err == nil && txRes.Code != uint32(codes.OK) {
				err = errorsmod.Wrapf(ErrUnsuccessfulTx, "code = %d; %s", txRes.Code, txRes.RawLog)
			}
		}
		break
	}
	return txRes, err
}
