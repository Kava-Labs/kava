package noop

import (
	"fmt"

	"github.com/ethereum/go-ethereum/precompile/contract"
)

// NewContract returns a new noop stateful precompiled contract.
func NewContract() (contract.StatefulPrecompiledContract, error) {
	precompile, err := contract.NewStatefulPrecompileContract([]*contract.StatefulPrecompileFunction{})

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate noop precompile: %w", err)
	}

	return precompile, nil
}
