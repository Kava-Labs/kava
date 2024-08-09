package registry

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/ethereum/go-ethereum/precompile/modules"

	"github.com/kava-labs/kava/precompile/contracts/noop"
)

const (
	// NoopContractAddress the primary noop contract address for testing
	NoopContractAddress = "0x9000000000000000000000000000000000000001"
	// NoopContractAddress2 the secondary noop contract address for testing
	NoopContractAddress2 = "0x9000000000000000000000000000000000000002"
)

// init registers stateful precompile contracts with the global precompile registry
// defined in kava-labs/go-ethereum/precompile/modules
func init() {
	register(NoopContractAddress, noop.NewContract)
	register(NoopContractAddress2, noop.NewContract)
}

// register accepts a 0x address string and a stateful precompile contract constructor, instantiates the
// precompile contract via the constructor, and registers it with the precompile module registry.
//
// This panics if the contract can not be created or the module can not be registered
func register(address string, newContract func() (contract.StatefulPrecompiledContract, error)) {
	contract, err := newContract()
	if err != nil {
		panic(fmt.Errorf("error creating contract for address %s: %w", address, err))
	}

	module := modules.Module{
		Address:  common.HexToAddress(address),
		Contract: contract,
	}

	err = modules.RegisterModule(module)
	if err != nil {
		panic(fmt.Errorf("error registering contract module for address %s: %w", address, err))
	}
}
