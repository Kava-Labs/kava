package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/hd"
	"github.com/evmos/ethermint/server/config"
	etherminttypes "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

// CanSignEthTx returns an error if the signing key algorithm is not eth_secp256k1.
func CanSignEthTx(ctx client.Context) error {
	keyInfo, err := ctx.Keyring.KeyByAddress(ctx.FromAddress)
	if err != nil {
		return err
	}

	pubKey, err := keyInfo.GetPubKey()
	if err != nil {
		return err
	}

	if pubKey.Type() != string(hd.EthSecp256k1Type) {
		return fmt.Errorf(
			"invalid from address pubkey type, expected %s but got %s",
			hd.EthSecp256k1Type,
			pubKey.Type(),
		)
	}

	return nil
}

// PackContractCallData creates a smart contract method call data with the
// provided method and args.
func PackContractCallData(abi abi.ABI, method string, args ...interface{}) ([]byte, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction data: %w", err)
	}

	return data, nil
}

// CreateEthCallContractTx creates and signs a Eth transaction wrapped in a
// cosmos Tx.
func CreateEthCallContractTx(
	ctx client.Context,
	contractAddr *common.Address,
	data []byte,
) (signing.Tx, error) {
	evmQueryClient := evmtypes.NewQueryClient(ctx)
	feemarketQueryClient := feemarkettypes.NewQueryClient(ctx)

	chainID, err := etherminttypes.ParseChainID(ctx.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain ID: %w", err)
	}

	evmParamsRes, err := evmQueryClient.Params(context.Background(), &evmtypes.QueryParamsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch evm params: %w", err)
	}

	// Estimate Gas
	from := common.BytesToAddress(ctx.FromAddress.Bytes())
	transactionArgs := evmtypes.TransactionArgs{
		From: &from,
		To:   contractAddr,
		Data: (*hexutil.Bytes)(&data),
	}

	args, err := json.Marshal(transactionArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction args for gas estimate: %w", err)
	}

	res, err := evmQueryClient.EstimateGas(context.Background(), &evmtypes.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas from EVM: %w", err)
	}

	// Fetch base fee
	basefeeRes, err := feemarketQueryClient.BaseFee(
		context.Background(),
		&feemarkettypes.QueryBaseFeeRequest{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch basefee from feemarket: %w", err)
	}

	// Fetch account nonce, ignore error to use use 0 nonce if first tx
	_, accSeq, _ := ctx.AccountRetriever.GetAccountNumberSequence(ctx, ctx.FromAddress)

	// Create MsgEthereumTx
	ethTx := evmtypes.NewTx(
		chainID,
		accSeq,                      // nonce
		contractAddr,                // to
		nil,                         // amount
		res.Gas,                     // gasLimit
		nil,                         // gasPrice
		basefeeRes.BaseFee.BigInt(), // gasFeeCap
		big.NewInt(1),               // gasTipCap
		data,                        // input
		&ethtypes.AccessList{},
	)

	// Must set from address before signing
	ethTx.From = from.String()

	// Sign Ethereum TX (not the cosmos Msg)
	signer := ethtypes.LatestSignerForChainID(chainID)

	// Must sign with a `/ethermint.crypto.v1.ethsecp256k1.PubKey` and not
	// `/cosmos.crypto.secp256k1.PubKey` or this will panic with the following:
	// panic: wrong size for signature: got 64, want 65
	if err := ethTx.Sign(signer, ctx.Keyring); err != nil {
		return nil, err
	}

	return ethTx.BuildTx(ctx.TxConfig.NewTxBuilder(), evmParamsRes.Params.EvmDenom)
}
