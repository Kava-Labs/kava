// Derived from https://github.com/tharsis/evmos/blob/0bfaf0db7be47153bc651e663176ba8deca960b5/x/erc20/keeper/evm.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keeper

import (
	"encoding/json"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tharsis/ethermint/server/config"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// CallEVM performs a smart contract method call using given args
func (k Keeper) CallEVM(
	ctx sdk.Context,
	abi abi.ABI,
	from common.Address,
	contract types.InternalEVMAddress,
	method string,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return nil, sdkerrors.Wrap(
			types.ErrABIPack,
			sdkerrors.Wrap(err, "failed to create transaction data").Error(),
		)
	}

	resp, err := k.CallEVMWithData(ctx, from, &contract, data)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "contract call failed: method '%s', contract '%s'", method, contract)
	}
	return resp, nil
}

// CallEVMWithData performs a smart contract method call using contract data
// Derived from tharsis/evmos
// https://github.com/tharsis/evmos/blob/ee54f496551df937915ff6f74a94732a35abc505/x/erc20/keeper/evm.go
func (k Keeper) CallEVMWithData(
	ctx sdk.Context,
	from common.Address,
	contract *types.InternalEVMAddress,
	data []byte,
) (*evmtypes.MsgEthereumTxResponse, error) {
	nonce, err := k.accountKeeper.GetSequence(ctx, from.Bytes())
	if err != nil {
		return nil, err
	}

	// To param needs to be nil to correctly apply txs to create contracts
	// Default common.Address value is 0x0000000000000000000000000000000000000000, not nil
	// which Ethermint handles differently -- erc20_test will fail
	// https://github.com/tharsis/ethermint/blob/caa1c5a6c6b7ed8ba4aaf6e0b0848f6be5ba6668/x/evm/keeper/state_transition.go#L357
	var to *common.Address
	if contract != nil {
		to = &contract.Address
	}

	transactionArgs := evmtypes.TransactionArgs{
		From: &from,
		To:   to,
		Data: (*hexutil.Bytes)(&data),
	}

	args, err := json.Marshal(transactionArgs)
	if err != nil {
		return nil, err
	}

	ethGasContext := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// EstimateGas applies the transaction against current block state to get
	// optimal gas value. Since this is done right before the ApplyMessage
	// below, it should essentially do the same thing but without affecting
	// state. While this is an *estimate* in regular use, this should be an
	// accurate exact amount in this case, as both the chain state and tx used
	// to estimate and apply are the exact same (ie. no txs between estimate and
	// apply, tx order is the same, etc.)
	gasRes, err := k.evmKeeper.EstimateGas(sdk.WrapSDKContext(ethGasContext), &evmtypes.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	if err != nil {
		return nil, sdkerrors.Wrap(evmtypes.ErrVMExecution, err.Error())
	}

	msg := ethtypes.NewMessage(
		from,
		to,
		nonce,
		big.NewInt(0), // amount
		gasRes.Gas,    // gasLimit
		big.NewInt(0), // gasFeeCap
		big.NewInt(0), // gasTipCap
		big.NewInt(0), // gasPrice
		data,
		ethtypes.AccessList{}, // AccessList
		true,                  // checkNonce
	)

	res, err := k.evmKeeper.ApplyMessage(ethGasContext, msg, evmtypes.NewNoOpTracer(), true)
	if err != nil {
		return nil, err
	}

	ctx.GasMeter().ConsumeGas(res.GasUsed, "evm gas consumed")

	if res.Failed() {
		return nil, sdkerrors.Wrap(evmtypes.ErrVMExecution, res.VmError)
	}

	return res, nil
}

// monitorApprovalEvent returns an error if the given transactions logs include
// an unexpected `Approval` event
func (k Keeper) monitorApprovalEvent(res *evmtypes.MsgEthereumTxResponse) error {
	if res == nil || len(res.Logs) == 0 {
		return nil
	}

	logApprovalSig := []byte("Approval(address,address,uint256)")
	logApprovalSigHash := crypto.Keccak256Hash(logApprovalSig)

	for _, log := range res.Logs {
		if log.Topics[0] == logApprovalSigHash.Hex() {
			return sdkerrors.Wrapf(
				types.ErrUnexpectedContractEvent, "unexpected contract Approval event",
			)
		}
	}

	return nil
}
