package ibc

import (
	_ "embed"
	"errors"

	"github.com/ethereum/go-ethereum/precompile/contract"
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

func createIBCPrecompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("transferKava(string,string,string,uint64,uint64,uint64,string)"),
		transferKava,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("transferCosmosDenom(string,string,string,uint256,string,uint64,uint64,uint64,string)"),
		transferCosmosDenom,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("transferERC20(string,string,uint256,string,uint64,uint64,uint64,string,string)"),
		transferERC20,
	))

	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}
