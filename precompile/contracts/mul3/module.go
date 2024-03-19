package mul3

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/precompile/modules"
)

// ContractAddress is the defined address of the precompile contract.
// This should be unique across all precompile contracts.
// See precompile/registry/registry.go for registered precompile contracts and more information.
var ContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000001")

// Module is the precompile module. It is used to register the precompile contract.
var Module = modules.Module{
	Address:  ContractAddress,
	Contract: Mul3Precompile,
}

func init() {
	// Register the precompile module.
	// Each precompile contract registers itself through [RegisterModule] function.
	if err := modules.RegisterModule(Module); err != nil {
		panic(err)
	}
}
