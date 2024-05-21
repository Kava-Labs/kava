package noop

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/precompile/contract"
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

// createNoopPrecompile returns a StatefulPrecompiledContract with getters and setters for the precompile.
func createNoopPrecompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("noop()"),
		noop,
	))

	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}
