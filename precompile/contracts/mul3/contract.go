package mul3

import (
	_ "embed"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/precompile/contract"
	"github.com/ethereum/go-ethereum/vmerrs"
)

const (
	calcMul3GasCost uint64 = contract.WriteGasCostPerSlot
	getMul3GasCost  uint64 = contract.ReadGasCostPerSlot
)

// Singleton StatefulPrecompiledContract.
var (
	// Mul3RawABI contains the raw ABI of Mul3 contract.
	//go:embed IMul3.abi
	Mul3RawABI string

	Mul3ABI = contract.MustParseABI(Mul3RawABI)

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
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, calcMul3GasCost); err != nil {
		return nil, 0, err
	}
	if readOnly {
		return nil, remainingGas, vmerrs.ErrWriteProtection
	}

	if len(input) != 96 {
		return nil, remainingGas, fmt.Errorf("unexpected input length, want: 96, got: %v", len(input))
	}

	var a, b, c, rez big.Int
	a.SetBytes(input[:32])
	b.SetBytes(input[32:64])
	c.SetBytes(input[64:96])
	rez.Mul(&a, &b)
	rez.Mul(&rez, &c)

	StoreProduct(accessibleState.GetStateDB(), &rez)

	packedOutput := make([]byte, 0)
	return packedOutput, remainingGas, nil
}

func getMul3(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, getMul3GasCost); err != nil {
		return nil, 0, err
	}

	product, err := GetProduct(accessibleState.GetStateDB())
	if err != nil {
		return nil, remainingGas, err
	}

	packedOutput := common.LeftPadBytes(product.Bytes(), 32)
	return packedOutput, remainingGas, nil
}

func calcMul3WithError(
	accessibleState contract.AccessibleState,
	caller common.Address,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	readOnly bool,
) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = contract.DeductGas(suppliedGas, calcMul3GasCost); err != nil {
		return nil, 0, err
	}
	if readOnly {
		return nil, remainingGas, vmerrs.ErrWriteProtection
	}

	if len(input) != 96 {
		return nil, remainingGas, fmt.Errorf("unexpected input length, want: 96, got: %v", len(input))
	}

	var a, b, c, rez big.Int
	a.SetBytes(input[:32])
	b.SetBytes(input[32:64])
	c.SetBytes(input[64:96])
	rez.Mul(&a, &b)
	rez.Mul(&rez, &c)

	StoreProduct(accessibleState.GetStateDB(), &rez)

	return nil, remainingGas, fmt.Errorf("calculation error")
}

// createMul3Precompile returns a StatefulPrecompiledContract with getters and setters for the precompile.
func createMul3Precompile() contract.StatefulPrecompiledContract {
	var functions []*contract.StatefulPrecompileFunction

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("calcMul3(uint256,uint256,uint256)"),
		calcMul3,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("getMul3()"),
		getMul3,
	))

	functions = append(functions, contract.NewStatefulPrecompileFunction(
		contract.MustCalculateFunctionSelector("calcMul3WithError(uint256,uint256,uint256)"),
		calcMul3WithError,
	))

	// Construct the contract with no fallback function.
	statefulContract, err := contract.NewStatefulPrecompileContract(functions)
	if err != nil {
		panic(err)
	}
	return statefulContract
}

type CalcMul3Input struct {
	A *big.Int
	B *big.Int
	C *big.Int
}

// PackCalcMul3 packs [inputStruct] of type CalcMul3Input into the appropriate arguments for calcMul3.
func PackCalcMul3(inputStruct CalcMul3Input) ([]byte, error) {
	return Mul3ABI.Pack("calcMul3", inputStruct.A, inputStruct.B, inputStruct.C)
}

// PackGetMul3 packs the include selector (first 4 func signature bytes).
func PackGetMul3() ([]byte, error) {
	return Mul3ABI.Pack("getMul3")
}

// UnpackGetMul3Output attempts to unpack given [output] into the *big.Int type output
// assumes that [output] does not include selector (omits first 4 func signature bytes)
func UnpackGetMul3Output(output []byte) (*big.Int, error) {
	res, err := Mul3ABI.Unpack("getMul3", output)
	if err != nil {
		return new(big.Int), err
	}
	unpacked := *abi.ConvertType(res[0], new(*big.Int)).(**big.Int)
	return unpacked, nil
}
