package noop

import (
	"fmt"

	"github.com/ethereum/go-ethereum/precompile/contract"
)

// NewContract returns a new noop stateful precompiled contract.
//
//	This contract is used for testing purposes only and should not be used on public chains.
//	The functions of this contract (once implemented), will be used to exercise and test the various aspects of
//	the EVM such as gas usage, argument parsing, events, etc. The specific operations tested under this contract are
//	still to be determined.
func NewContract() (contract.StatefulPrecompiledContract, error) {
	precompile, err := contract.NewStatefulPrecompileContract([]*contract.StatefulPrecompileFunction{})

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate noop precompile: %w", err)
	}

	return precompile, nil
}
