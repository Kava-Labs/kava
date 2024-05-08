package noop

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/precompile/modules"
)

var (
	// ContractAddress is the defined address of the precompile contract.
	// This should be unique across all precompile contracts.
	// See precompile/registry/registry.go for registered precompile contracts and more information.
	ContractAddress = common.HexToAddress("0x9000000000000000000000000000000000000001")

	// ContractAddress2 is a second address on which Noop precompile is registered, it can be useful for testing purposes.
	ContractAddress2 = common.HexToAddress("0x9000000000000000000000000000000000000002")
)

// Module is the precompile module. It is used to register the precompile contract.
var Module = modules.Module{
	Address:  ContractAddress,
	Contract: NoopPrecompile,
}

var Module2 = modules.Module{
	Address:  ContractAddress2,
	Contract: NoopPrecompile,
}

func init() {
	// Register the precompile module.
	// Each precompile contract registers itself through [RegisterModule] function.
	if err := modules.RegisterModule(Module); err != nil {
		panic(err)
	}

	if err := modules.RegisterModule(Module2); err != nil {
		panic(err)
	}
}
