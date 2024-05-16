package noop

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	accessibleState.GetStateDB().AddLog(&types.Log{
		//// Consensus fields:
		//// address of the contract that generated the event
		//Address common.Address
		//// list of topics provided by the contract.
		//Topics []common.Hash
		//// supplied by the contract, usually ABI-encoded
		//Data []byte
		//
		//// Derived fields. These fields are filled in by the node
		//// but not secured by consensus.
		//// block in which the transaction was included
		//BlockNumber uint64
		//// hash of the transaction
		//TxHash common.Hash
		//// index of the transaction in the block
		//TxIndex uint
		//// hash of the block in which the transaction was included
		//BlockHash common.Hash
		//// index of the log in the block
		//Index uint
		//
		//// The Removed field is true if this log was reverted due to a chain reorganisation.
		//// You must pay attention to this field if you receive logs through a filter query.
		//Removed bool
	})
	ethermintStateDB := accessibleState.GetStateDB().(*statedb.StateDB)

	err = ethermintStateDB.Commit()
	fmt.Printf("Commit Error: %v\n", err)
	if err != nil {
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

	accessibleState.GetStateDB().AddLog(&types.Log{
		//// Consensus fields:
		//// address of the contract that generated the event
		//Address common.Address
		//// list of topics provided by the contract.
		//Topics []common.Hash
		//// supplied by the contract, usually ABI-encoded
		//Data []byte
		//
		//// Derived fields. These fields are filled in by the node
		//// but not secured by consensus.
		//// block in which the transaction was included
		//BlockNumber uint64
		//// hash of the transaction
		//TxHash common.Hash
		//// index of the transaction in the block
		//TxIndex uint
		//// hash of the block in which the transaction was included
		//BlockHash common.Hash
		//// index of the log in the block
		//Index uint
	})
	ethermintStateDB := accessibleState.GetStateDB().(*statedb.StateDB)

	err = ethermintStateDB.Commit()
	fmt.Printf("Commit Error: %v\n", err)
	if err != nil {
		return nil, 0, err
	}

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
