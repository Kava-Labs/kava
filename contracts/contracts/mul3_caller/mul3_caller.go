// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package mul3_caller

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// Mul3CallerMetaData contains all meta data concerning the Mul3Caller contract.
var Mul3CallerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"c\",\"type\":\"uint256\"}],\"name\":\"calcMul3\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"c\",\"type\":\"uint256\"}],\"name\":\"calcMul3Call\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMul3\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"result\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMul3StaticCall\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234601c57600e6020565b61071461002c823961071490f35b6026565b60405190565b600080fdfe60806040526004361015610013575b610268565b61001e60003561005d565b80632eb84fe214610058578063446a94a4146100535780635c8062c21461004e5763dbb45c000361000e57610234565b6101f8565b610164565b6100a9565b60e01c90565b60405190565b600080fd5b600080fd5b600091031261007e57565b61006e565b90565b61008f90610083565b9052565b91906100a790600060208501940190610086565b565b346100d9576100b9366004610073565b6100d56100c461035d565b6100cc610063565b91829182610093565b0390f35b610069565b5190565b60209181520190565b60005b8381106100ff575050906000910152565b8060209183015181850152016100ee565b601f801991011690565b61013961014260209361014793610130816100de565b938480936100e2565b958691016100eb565b610110565b0190565b610161916020820191600081840391015261011a565b90565b3461019457610174366004610073565b61019061017f610525565b610187610063565b9182918261014b565b0390f35b610069565b6101a281610083565b036101a957565b600080fd5b905035906101bb82610199565b565b90916060828403126101f3576101f06101d984600085016101ae565b936101e781602086016101ae565b936040016101ae565b90565b61006e565b346102295761022561021461020e3660046101bd565b916105c1565b61021c610063565b9182918261014b565b0390f35b610069565b60000190565b346102635761024d6102473660046101bd565b91610642565b610255610063565b8061025f8161022e565b0390f35b610069565b600080fd5b600090565b600360981b90565b60018060a01b031690565b90565b61029c6102976102a19261027a565b610285565b61027a565b90565b6102ad90610288565b90565b6102b9906102a4565b90565b6102c590610288565b90565b6102d1906102bc565b90565b600080fd5b634e487b7160e01b600052604160045260246000fd5b906102f990610110565b810190811067ffffffffffffffff82111761031357604052565b6102d9565b60e01b90565b9050519061032b82610199565b565b90602082820312610347576103449160000161031e565b90565b61006e565b610354610063565b3d6000823e3d90fd5b61036561026d565b50610399602061038361037e610379610272565b6102b0565b6102c8565b632eb84fe290610391610063565b938492610318565b825281806103a96004820161022e565b03915afa9081156103ee576000916103c0575b5090565b6103e1915060203d81116103e7575b6103d981836102ef565b81019061032d565b386103bc565b503d6103cf565b61034c565b606090565b9061040b610404610063565b92836102ef565b565b67ffffffffffffffff811161042b57610427602091610110565b0190565b6102d9565b9061044261043d8361040d565b6103f8565b918252565b3d600014610464576104583d610430565b903d6000602084013e5b565b61046c6103f3565b90610462565b60209181520190565b60207f6c65640000000000000000000000000000000000000000000000000000000000917f63616c6c20746f20707265636f6d70696c656420636f6e74726163742066616960008201520152565b6104d66023604092610472565b6104df8161047b565b0190565b6104f990602081019060008183039101526104c9565b90565b1561050357565b61050b610063565b62461bcd60e51b815280610521600482016104e3565b0390fd5b61052d6103f3565b50600080600461056863175c27f160e11b610559610549610063565b938492602084019081520161022e565b602082018103825203826102ef565b610570610272565b90602081019051915afa61058b610585610447565b916104fc565b90565b6040906105b86105bf94969593966105ae60608401986000850190610086565b6020830190610086565b0190610086565b565b61060b6004916105fc6000959486956105d86103f3565b506236ed1760ea1b939190916105ec610063565b968795602087019081520161058e565b602082018103825203826102ef565b610613610272565b9082602082019151925af161062f610629610447565b916104fc565b90565b600091031261063d57565b61006e565b9061065b610656610651610272565b6102b0565b6102c8565b63dbb45c0092919392813b156106d957600061068a91610695829661067e610063565b98899788968795610318565b85526004850161058e565b03925af180156106d4576106a7575b50565b6106c79060003d81116106cd575b6106bf81836102ef565b810190610632565b386106a4565b503d6106b5565b61034c565b6102d456fea26469706673582212205d0960a8100fec0947a16b6682f7f9157ffbfacc2d3352f3e0867bd04523c9ae64736f6c63430008190033",
}

// Mul3CallerABI is the input ABI used to generate the binding from.
// Deprecated: Use Mul3CallerMetaData.ABI instead.
var Mul3CallerABI = Mul3CallerMetaData.ABI

// Mul3CallerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use Mul3CallerMetaData.Bin instead.
var Mul3CallerBin = Mul3CallerMetaData.Bin

// DeployMul3Caller deploys a new Ethereum contract, binding an instance of Mul3Caller to it.
func DeployMul3Caller(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Mul3Caller, error) {
	parsed, err := Mul3CallerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(Mul3CallerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Mul3Caller{Mul3CallerCaller: Mul3CallerCaller{contract: contract}, Mul3CallerTransactor: Mul3CallerTransactor{contract: contract}, Mul3CallerFilterer: Mul3CallerFilterer{contract: contract}}, nil
}

// Mul3Caller is an auto generated Go binding around an Ethereum contract.
type Mul3Caller struct {
	Mul3CallerCaller     // Read-only binding to the contract
	Mul3CallerTransactor // Write-only binding to the contract
	Mul3CallerFilterer   // Log filterer for contract events
}

// Mul3CallerCaller is an auto generated read-only Go binding around an Ethereum contract.
type Mul3CallerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Mul3CallerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type Mul3CallerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Mul3CallerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Mul3CallerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Mul3CallerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Mul3CallerSession struct {
	Contract     *Mul3Caller       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Mul3CallerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Mul3CallerCallerSession struct {
	Contract *Mul3CallerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// Mul3CallerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Mul3CallerTransactorSession struct {
	Contract     *Mul3CallerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// Mul3CallerRaw is an auto generated low-level Go binding around an Ethereum contract.
type Mul3CallerRaw struct {
	Contract *Mul3Caller // Generic contract binding to access the raw methods on
}

// Mul3CallerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Mul3CallerCallerRaw struct {
	Contract *Mul3CallerCaller // Generic read-only contract binding to access the raw methods on
}

// Mul3CallerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Mul3CallerTransactorRaw struct {
	Contract *Mul3CallerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMul3Caller creates a new instance of Mul3Caller, bound to a specific deployed contract.
func NewMul3Caller(address common.Address, backend bind.ContractBackend) (*Mul3Caller, error) {
	contract, err := bindMul3Caller(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Mul3Caller{Mul3CallerCaller: Mul3CallerCaller{contract: contract}, Mul3CallerTransactor: Mul3CallerTransactor{contract: contract}, Mul3CallerFilterer: Mul3CallerFilterer{contract: contract}}, nil
}

// NewMul3CallerCaller creates a new read-only instance of Mul3Caller, bound to a specific deployed contract.
func NewMul3CallerCaller(address common.Address, caller bind.ContractCaller) (*Mul3CallerCaller, error) {
	contract, err := bindMul3Caller(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Mul3CallerCaller{contract: contract}, nil
}

// NewMul3CallerTransactor creates a new write-only instance of Mul3Caller, bound to a specific deployed contract.
func NewMul3CallerTransactor(address common.Address, transactor bind.ContractTransactor) (*Mul3CallerTransactor, error) {
	contract, err := bindMul3Caller(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Mul3CallerTransactor{contract: contract}, nil
}

// NewMul3CallerFilterer creates a new log filterer instance of Mul3Caller, bound to a specific deployed contract.
func NewMul3CallerFilterer(address common.Address, filterer bind.ContractFilterer) (*Mul3CallerFilterer, error) {
	contract, err := bindMul3Caller(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Mul3CallerFilterer{contract: contract}, nil
}

// bindMul3Caller binds a generic wrapper to an already deployed contract.
func bindMul3Caller(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := Mul3CallerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Mul3Caller *Mul3CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Mul3Caller.Contract.Mul3CallerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Mul3Caller *Mul3CallerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Mul3Caller.Contract.Mul3CallerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Mul3Caller *Mul3CallerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Mul3Caller.Contract.Mul3CallerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Mul3Caller *Mul3CallerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Mul3Caller.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Mul3Caller *Mul3CallerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Mul3Caller.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Mul3Caller *Mul3CallerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Mul3Caller.Contract.contract.Transact(opts, method, params...)
}

// GetMul3 is a free data retrieval call binding the contract method 0x2eb84fe2.
//
// Solidity: function getMul3() view returns(uint256 result)
func (_Mul3Caller *Mul3CallerCaller) GetMul3(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Mul3Caller.contract.Call(opts, &out, "getMul3")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMul3 is a free data retrieval call binding the contract method 0x2eb84fe2.
//
// Solidity: function getMul3() view returns(uint256 result)
func (_Mul3Caller *Mul3CallerSession) GetMul3() (*big.Int, error) {
	return _Mul3Caller.Contract.GetMul3(&_Mul3Caller.CallOpts)
}

// GetMul3 is a free data retrieval call binding the contract method 0x2eb84fe2.
//
// Solidity: function getMul3() view returns(uint256 result)
func (_Mul3Caller *Mul3CallerCallerSession) GetMul3() (*big.Int, error) {
	return _Mul3Caller.Contract.GetMul3(&_Mul3Caller.CallOpts)
}

// GetMul3StaticCall is a free data retrieval call binding the contract method 0x446a94a4.
//
// Solidity: function getMul3StaticCall() view returns(bytes)
func (_Mul3Caller *Mul3CallerCaller) GetMul3StaticCall(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _Mul3Caller.contract.Call(opts, &out, "getMul3StaticCall")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetMul3StaticCall is a free data retrieval call binding the contract method 0x446a94a4.
//
// Solidity: function getMul3StaticCall() view returns(bytes)
func (_Mul3Caller *Mul3CallerSession) GetMul3StaticCall() ([]byte, error) {
	return _Mul3Caller.Contract.GetMul3StaticCall(&_Mul3Caller.CallOpts)
}

// GetMul3StaticCall is a free data retrieval call binding the contract method 0x446a94a4.
//
// Solidity: function getMul3StaticCall() view returns(bytes)
func (_Mul3Caller *Mul3CallerCallerSession) GetMul3StaticCall() ([]byte, error) {
	return _Mul3Caller.Contract.GetMul3StaticCall(&_Mul3Caller.CallOpts)
}

// CalcMul3 is a paid mutator transaction binding the contract method 0xdbb45c00.
//
// Solidity: function calcMul3(uint256 a, uint256 b, uint256 c) returns()
func (_Mul3Caller *Mul3CallerTransactor) CalcMul3(opts *bind.TransactOpts, a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.contract.Transact(opts, "calcMul3", a, b, c)
}

// CalcMul3 is a paid mutator transaction binding the contract method 0xdbb45c00.
//
// Solidity: function calcMul3(uint256 a, uint256 b, uint256 c) returns()
func (_Mul3Caller *Mul3CallerSession) CalcMul3(a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.Contract.CalcMul3(&_Mul3Caller.TransactOpts, a, b, c)
}

// CalcMul3 is a paid mutator transaction binding the contract method 0xdbb45c00.
//
// Solidity: function calcMul3(uint256 a, uint256 b, uint256 c) returns()
func (_Mul3Caller *Mul3CallerTransactorSession) CalcMul3(a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.Contract.CalcMul3(&_Mul3Caller.TransactOpts, a, b, c)
}

// CalcMul3Call is a paid mutator transaction binding the contract method 0x5c8062c2.
//
// Solidity: function calcMul3Call(uint256 a, uint256 b, uint256 c) returns(bytes)
func (_Mul3Caller *Mul3CallerTransactor) CalcMul3Call(opts *bind.TransactOpts, a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.contract.Transact(opts, "calcMul3Call", a, b, c)
}

// CalcMul3Call is a paid mutator transaction binding the contract method 0x5c8062c2.
//
// Solidity: function calcMul3Call(uint256 a, uint256 b, uint256 c) returns(bytes)
func (_Mul3Caller *Mul3CallerSession) CalcMul3Call(a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.Contract.CalcMul3Call(&_Mul3Caller.TransactOpts, a, b, c)
}

// CalcMul3Call is a paid mutator transaction binding the contract method 0x5c8062c2.
//
// Solidity: function calcMul3Call(uint256 a, uint256 b, uint256 c) returns(bytes)
func (_Mul3Caller *Mul3CallerTransactorSession) CalcMul3Call(a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.Contract.CalcMul3Call(&_Mul3Caller.TransactOpts, a, b, c)
}
