package noop

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/precompile/contract"

	"github.com/evmos/ethermint/x/evm/statedb"
)

const (
	noopGasCost uint64 = contract.ReadGasCostPerSlot
)

// Singleton StatefulPrecompiledContract.
var NoopPrecompile = createNoopPrecompile()

func noop(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, noopGasCost); err != nil {
		return nil, 0, err
	}

	return []byte{}, remainingGas, nil
}

func emitEvent(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, noopGasCost); err != nil {
		return nil, 0, err
	}

	eventSig := []byte("Event(string,string)")
	eventSigHash := crypto.Keccak256Hash(eventSig)
	testIndexedParamHash := crypto.Keccak256Hash([]byte("test-indexed-param"))

	accessibleState.GetStateDB().AddLog(&types.Log{
		Address: addr,
		Topics: []common.Hash{
			eventSigHash,
			testIndexedParamHash,
		},
		Data: []byte("test-param"),
	})
	ethermintStateDB := accessibleState.GetStateDB().(*statedb.StateDB)

	err = ethermintStateDB.Commit()
	fmt.Printf("Commit Error: %v\n", err)
	if err != nil {
		return nil, 0, err
	}

	fmt.Printf("ContractAddress: %v\n", ContractAddress)
	fmt.Printf("caller         : %v\n", caller)
	fmt.Printf("addr           : %v\n", addr)

	return []byte{}, remainingGas, nil
}

// createNoopPrecompile returns a StatefulPrecompiledContract with getters and setters for the precompile.
func createNoopPrecompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("noop()"),
		noop,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("emitEvent()"),
		emitEvent,
	))

	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}
