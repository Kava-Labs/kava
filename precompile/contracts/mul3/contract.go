package mul3

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/precompile/contract"
)

// Singleton StatefulPrecompiledContract.
var (
	Mul3Precompile = createMul3Precompile()
)

var (
	productKey = common.BytesToHash([]byte("productKey"))
)

func StoreProduct(stateDB vm.StateDB, product *big.Int) {
	valuePadded := common.LeftPadBytes(product.Bytes(), common.HashLength)
	valueHash := common.BytesToHash(valuePadded)

	stateDB.SetState(ContractAddress, productKey, valueHash)
}

func GetProduct(stateDB vm.StateDB) (*big.Int, error) {
	value := stateDB.GetState(ContractAddress, productKey)
	if len(value.Bytes()) == 0 {
		return big.NewInt(0), nil
	}

	var product big.Int
	product.SetBytes(value.Bytes())

	return &product, nil
}

func calcMul3(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, err error) {
	// Set the nonce of the precompile's address (as is done when a contract is created) to ensure
	// that it is marked as non-empty and will not be cleaned up when the statedb is finalized.
	accessibleState.GetStateDB().SetNonce(ContractAddress, 1)
	// Set the code of the precompile's address to a non-zero length byte slice to ensure that the precompile
	// can be called from within Solidity contracts. Solidity adds a check before invoking a contract to ensure
	// that it does not attempt to invoke a non-existent contract.
	accessibleState.GetStateDB().SetCode(ContractAddress, []byte{0x1})

	if len(input) != 96 {
		return nil, fmt.Errorf("unexpected input length, want: 96, got: %v", len(input))
	}

	var a, b, c, rez big.Int
	a.SetBytes(input[:32])
	b.SetBytes(input[32:64])
	c.SetBytes(input[64:96])
	rez.Mul(&a, &b)
	rez.Mul(&rez, &c)

	StoreProduct(accessibleState.GetStateDB(), &rez)

	packedOutput := make([]byte, 0)
	return packedOutput, nil
}

func getMul3(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, err error) {
	product, err := GetProduct(accessibleState.GetStateDB())
	if err != nil {
		return nil, err
	}

	packedOutput := common.LeftPadBytes(product.Bytes(), 32)
	return packedOutput, nil
}

// createMul3Precompile returns a StatefulPrecompiledContract with getters and setters for the precompile.
func createMul3Precompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.CalculateFunctionSelector("calcMul3(uint256,uint256,uint256)"),
		calcMul3,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.CalculateFunctionSelector("getMul3()"),
		getMul3,
	))

	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}
