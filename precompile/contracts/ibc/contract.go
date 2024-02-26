package ibc

import (
	_ "embed"
	"errors"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ethermint_statedb "github.com/evmos/ethermint/x/evm/statedb"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/ethereum/go-ethereum/vmerrs"
)

const (
	transferGasCost uint64 = contract.WriteGasCostPerSlot
)

// Singleton StatefulPrecompiledContract.
var (
	// IBCRawABI contains the raw ABI of IBC contract.
	//go:embed IBC.abi
	IBCRawABI string

	IBCABI = contract.MustParseABI(IBCRawABI)

	IBCPrecompile = createIBCPrecompile()
)

var ErrUnsupportedStateDB = errors.New("unsupported statedb")

func transfer(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, transferGasCost); err != nil {
		return nil, 0, err
	}
	if readOnly {
		return nil, remainingGas, vmerrs.ErrWriteProtection
	}

	// Set the nonce of the precompile's address (as is done when a contract is created) to ensure
	// that it is marked as non-empty and will not be cleaned up when the statedb is finalized.
	accessibleState.GetStateDB().SetNonce(ContractAddress, 1)
	// Set the code of the precompile's address to a non-zero length byte slice to ensure that the precompile
	// can be called from within Solidity contracts. Solidity adds a check before invoking a contract to ensure
	// that it does not attempt to invoke a non-existent contract.
	accessibleState.GetStateDB().SetCode(ContractAddress, []byte{0x1})

	transferInput, err := UnpackTransferInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	stateDB, ok := accessibleState.GetStateDB().(*ethermint_statedb.StateDB)
	if !ok {
		return nil, remainingGas, ErrUnsupportedStateDB
	}
	resp, err := stateDB.IBCTransfer(stateDB.Context(), &ibctransfertypes.MsgTransfer{
		SourcePort:    transferInput.sourcePort,
		SourceChannel: transferInput.sourceChannel,
		Token: sdk.Coin{
			Denom:  transferInput.denom,
			Amount: sdkmath.NewIntFromBigInt(&transferInput.amount),
		},
		Sender:   transferInput.sender,
		Receiver: transferInput.receiver,
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: transferInput.revisionNumber,
			RevisionHeight: transferInput.revisionHeight,
		},
		TimeoutTimestamp: transferInput.timeoutTimestamp,
		Memo:             "",
	})
	if err != nil {
		return nil, remainingGas, fmt.Errorf("IBCTransfer failed: %v\n", err)
	}
	_ = resp

	packedOutput := make([]byte, 0)
	return packedOutput, remainingGas, nil
}

func createIBCPrecompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("ibcTransfer(string,string,string,uint256,string,string,uint64,uint64,uint64)"),
		transfer,
	))

	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}

type TransferInput struct {
	sourcePort       string
	sourceChannel    string
	denom            string
	amount           big.Int
	sender           string
	receiver         string
	revisionNumber   uint64
	revisionHeight   uint64
	timeoutTimestamp uint64
}

// UnpackTransferInput attempts to unpack [input] into the *TransferInput type argument
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackTransferInput(input []byte) (*TransferInput, error) {
	// The strict mode in decoding is disabled after DUpgrade. You can re-enable by changing the last argument to true.
	res, err := IBCABI.UnpackInput("ibcTransfer", input, false)
	if err != nil {
		return nil, err
	}

	var transferInput TransferInput
	transferInput.sourcePort = *abi.ConvertType(res[0], new(string)).(*string)
	transferInput.sourceChannel = *abi.ConvertType(res[1], new(string)).(*string)
	transferInput.denom = *abi.ConvertType(res[2], new(string)).(*string)

	transferInput.amount = *abi.ConvertType(res[3], new(big.Int)).(*big.Int)
	transferInput.sender = *abi.ConvertType(res[4], new(string)).(*string)
	transferInput.receiver = *abi.ConvertType(res[5], new(string)).(*string)

	transferInput.revisionNumber = *abi.ConvertType(res[6], new(uint64)).(*uint64)
	transferInput.revisionHeight = *abi.ConvertType(res[7], new(uint64)).(*uint64)
	transferInput.timeoutTimestamp = *abi.ConvertType(res[8], new(uint64)).(*uint64)
	return &transferInput, nil
}
