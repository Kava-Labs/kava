package ibc

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/ethereum/go-ethereum/vmerrs"

	"github.com/kava-labs/kava/precompile/statedb"
	"github.com/kava-labs/kava/x/evmutil/types"
)

func transferERC20(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	packedInput []byte,
	suppliedGas uint64,
	readOnly bool,
	value *big.Int,
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

	input, err := UnpackTransferERC20Input(packedInput)
	if err != nil {
		return nil, remainingGas, err
	}

	stateDB, ok := accessibleState.GetStateDB().(statedb.StateDBWithKeepers)
	if !ok {
		return nil, remainingGas, ErrUnsupportedStateDB
	}

	// TODO(yevhenii): ERC20 address validation?

	conversionPair, err := stateDB.EvmutilKeeper().GetEnabledConversionPairFromERC20Address(
		sdk.UnwrapSDKContext(stateDB.Context()),
		types.NewInternalEVMAddress(common.HexToAddress(input.kavaERC20Address)),
	)
	if err != nil {
		return nil, remainingGas, err
	}

	_, err = stateDB.EvmutilKeeper().ConvertERC20ToCoin(
		stateDB.Context(),
		&types.MsgConvertERC20ToCoin{
			Initiator:        caller.String(),
			Receiver:         sdk.AccAddress(caller[:]).String(),
			KavaERC20Address: input.kavaERC20Address,
			Amount:           sdkmath.NewIntFromBigInt(&input.amount),
		},
	)
	if err != nil {
		return nil, remainingGas, err
	}

	_, err = stateDB.IBCTransferKeeper().Transfer(stateDB.Context(), &ibctransfertypes.MsgTransfer{
		SourcePort:    input.sourcePort,
		SourceChannel: input.sourceChannel,
		Token: sdk.Coin{
			Denom:  conversionPair.Denom,
			Amount: sdkmath.NewIntFromBigInt(&input.amount),
		},
		Sender:   sdk.AccAddress(caller[:]).String(),
		Receiver: input.receiver,
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: input.revisionNumber,
			RevisionHeight: input.revisionHeight,
		},
		TimeoutTimestamp: input.timeoutTimestamp,
		Memo:             input.memo,
	})
	if err != nil {
		return nil, remainingGas, fmt.Errorf("can't send IBC transfer: %v\n", err)
	}

	packedOutput := make([]byte, 0)
	return packedOutput, remainingGas, nil
}

type TransferERC20Input struct {
	sourcePort       string
	sourceChannel    string
	amount           big.Int
	receiver         string
	revisionNumber   uint64
	revisionHeight   uint64
	timeoutTimestamp uint64
	memo             string
	kavaERC20Address string
}

// UnpackTransferERC20Input attempts to unpack [input] into the *TransferERC20Input type argument
// assumes that [input] does not include selector (omits first 4 func signature bytes)
func UnpackTransferERC20Input(packedInput []byte) (*TransferERC20Input, error) {
	res, err := IBCABI.UnpackInput("transferERC20", packedInput)
	if err != nil {
		return nil, err
	}

	var input TransferERC20Input
	input.sourcePort = *abi.ConvertType(res[0], new(string)).(*string)
	input.sourceChannel = *abi.ConvertType(res[1], new(string)).(*string)

	input.amount = *abi.ConvertType(res[2], new(big.Int)).(*big.Int)
	input.receiver = *abi.ConvertType(res[3], new(string)).(*string)

	input.revisionNumber = *abi.ConvertType(res[4], new(uint64)).(*uint64)
	input.revisionHeight = *abi.ConvertType(res[5], new(uint64)).(*uint64)
	input.timeoutTimestamp = *abi.ConvertType(res[6], new(uint64)).(*uint64)
	input.memo = *abi.ConvertType(res[7], new(string)).(*string)
	input.kavaERC20Address = *abi.ConvertType(res[8], new(string)).(*string)
	return &input, nil
}
