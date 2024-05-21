// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package noop_caller

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

// NoopCallerMetaData contains all meta data concerning the NoopCaller contract.
var NoopCallerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"noop\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"noop_static_call\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60806040523461002f576100196100146100fa565b610196565b610021610034565b6104bc6101a482396104bc90f35b61003a565b60405190565b600080fd5b601f801991011690565b634e487b7160e01b600052604160045260246000fd5b906100699061003f565b810190811060018060401b0382111761008157604052565b610049565b90610099610092610034565b928361005f565b565b600080fd5b60018060a01b031690565b6100b4906100a0565b90565b6100c0816100ab565b036100c757565b600080fd5b905051906100d9826100b7565b565b906020828203126100f5576100f2916000016100cc565b90565b61009b565b6101186106608038038061010d81610086565b9283398101906100db565b90565b60001b90565b9061013260018060a01b039161011b565b9181191691161790565b90565b61015361014e610158926100a0565b61013c565b6100a0565b90565b6101649061013f565b90565b6101709061015b565b90565b90565b9061018b61018661019292610167565b610173565b8254610121565b9055565b6101a1906000610176565b56fe60806040526004361015610013575b610157565b61001e60003561003d565b80635dfc2e4a146100385763a79ad1a50361000e57610122565b610069565b60e01c90565b60405190565b600080fd5b600080fd5b600091031261005e57565b61004e565b60000190565b3461009757610079366004610053565b610081610253565b610089610043565b8061009381610063565b0390f35b610049565b5190565b60209181520190565b60005b8381106100bd575050906000910152565b8060209183015181850152016100ac565b601f801991011690565b6100f7610100602093610105936100ee8161009c565b938480936100a0565b958691016100a9565b6100ce565b0190565b61011f91602082019160008184039101526100d8565b90565b3461015257610132366004610053565b61014e61013d61041c565b610145610043565b91829182610109565b0390f35b610049565b600080fd5b60001c90565b60018060a01b031690565b61017961017e9161015c565b610162565b90565b61018b905461016d565b90565b60018060a01b031690565b90565b6101b06101ab6101b59261018e565b610199565b61018e565b90565b6101c19061019c565b90565b6101cd906101b8565b90565b6101d99061019c565b90565b6101e5906101d0565b90565b600080fd5b634e487b7160e01b600052604160045260246000fd5b9061020d906100ce565b810190811067ffffffffffffffff82111761022757604052565b6101ed565b60e01b90565b600091031261023d57565b61004e565b61024a610043565b3d6000823e3d90fd5b61026d6102686102636000610181565b6101c4565b6101dc565b635dfc2e4a90803b156102e55761029191600091610289610043565b93849261022c565b825281806102a160048201610063565b03915afa80156102e0576102b3575b50565b6102d39060003d81116102d9575b6102cb8183610203565b810190610232565b386102b0565b503d6102c1565b610242565b6101e8565b606090565b906103026102fb610043565b9283610203565b565b67ffffffffffffffff81116103225761031e6020916100ce565b0190565b6101ed565b9061033961033483610304565b6102ef565b918252565b3d60001461035b5761034f3d610327565b903d6000602084013e5b565b6103636102ea565b90610359565b60209181520190565b60207f6c65640000000000000000000000000000000000000000000000000000000000917f63616c6c20746f20707265636f6d70696c656420636f6e74726163742066616960008201520152565b6103cd6023604092610369565b6103d681610372565b0190565b6103f090602081019060008183039101526103c0565b90565b156103fa57565b610402610043565b62461bcd60e51b815280610418600482016103da565b0390fd5b6104246102ea565b50600080600461045f632efe172560e11b610450610440610043565b9384926020840190815201610063565b60208201810382520382610203565b61046882610181565b90602081019051915afa61048361047d61033e565b916103f3565b9056fea26469706673582212201949af946f01415abf551272444f438e8c4bf54581b4d4006ae0452c7b08fdf864736f6c63430008190033",
}

// NoopCallerABI is the input ABI used to generate the binding from.
// Deprecated: Use NoopCallerMetaData.ABI instead.
var NoopCallerABI = NoopCallerMetaData.ABI

// NoopCallerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use NoopCallerMetaData.Bin instead.
var NoopCallerBin = NoopCallerMetaData.Bin

// DeployNoopCaller deploys a new Ethereum contract, binding an instance of NoopCaller to it.
func DeployNoopCaller(auth *bind.TransactOpts, backend bind.ContractBackend, _target common.Address) (common.Address, *types.Transaction, *NoopCaller, error) {
	parsed, err := NoopCallerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(NoopCallerBin), backend, _target)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &NoopCaller{NoopCallerCaller: NoopCallerCaller{contract: contract}, NoopCallerTransactor: NoopCallerTransactor{contract: contract}, NoopCallerFilterer: NoopCallerFilterer{contract: contract}}, nil
}

// NoopCaller is an auto generated Go binding around an Ethereum contract.
type NoopCaller struct {
	NoopCallerCaller     // Read-only binding to the contract
	NoopCallerTransactor // Write-only binding to the contract
	NoopCallerFilterer   // Log filterer for contract events
}

// NoopCallerCaller is an auto generated read-only Go binding around an Ethereum contract.
type NoopCallerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NoopCallerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type NoopCallerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NoopCallerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type NoopCallerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NoopCallerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type NoopCallerSession struct {
	Contract     *NoopCaller       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// NoopCallerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type NoopCallerCallerSession struct {
	Contract *NoopCallerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// NoopCallerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type NoopCallerTransactorSession struct {
	Contract     *NoopCallerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// NoopCallerRaw is an auto generated low-level Go binding around an Ethereum contract.
type NoopCallerRaw struct {
	Contract *NoopCaller // Generic contract binding to access the raw methods on
}

// NoopCallerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type NoopCallerCallerRaw struct {
	Contract *NoopCallerCaller // Generic read-only contract binding to access the raw methods on
}

// NoopCallerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type NoopCallerTransactorRaw struct {
	Contract *NoopCallerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewNoopCaller creates a new instance of NoopCaller, bound to a specific deployed contract.
func NewNoopCaller(address common.Address, backend bind.ContractBackend) (*NoopCaller, error) {
	contract, err := bindNoopCaller(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &NoopCaller{NoopCallerCaller: NoopCallerCaller{contract: contract}, NoopCallerTransactor: NoopCallerTransactor{contract: contract}, NoopCallerFilterer: NoopCallerFilterer{contract: contract}}, nil
}

// NewNoopCallerCaller creates a new read-only instance of NoopCaller, bound to a specific deployed contract.
func NewNoopCallerCaller(address common.Address, caller bind.ContractCaller) (*NoopCallerCaller, error) {
	contract, err := bindNoopCaller(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &NoopCallerCaller{contract: contract}, nil
}

// NewNoopCallerTransactor creates a new write-only instance of NoopCaller, bound to a specific deployed contract.
func NewNoopCallerTransactor(address common.Address, transactor bind.ContractTransactor) (*NoopCallerTransactor, error) {
	contract, err := bindNoopCaller(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &NoopCallerTransactor{contract: contract}, nil
}

// NewNoopCallerFilterer creates a new log filterer instance of NoopCaller, bound to a specific deployed contract.
func NewNoopCallerFilterer(address common.Address, filterer bind.ContractFilterer) (*NoopCallerFilterer, error) {
	contract, err := bindNoopCaller(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &NoopCallerFilterer{contract: contract}, nil
}

// bindNoopCaller binds a generic wrapper to an already deployed contract.
func bindNoopCaller(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := NoopCallerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NoopCaller *NoopCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NoopCaller.Contract.NoopCallerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NoopCaller *NoopCallerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NoopCaller.Contract.NoopCallerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NoopCaller *NoopCallerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NoopCaller.Contract.NoopCallerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_NoopCaller *NoopCallerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _NoopCaller.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_NoopCaller *NoopCallerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NoopCaller.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_NoopCaller *NoopCallerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _NoopCaller.Contract.contract.Transact(opts, method, params...)
}

// Noop is a free data retrieval call binding the contract method 0x5dfc2e4a.
//
// Solidity: function noop() view returns()
func (_NoopCaller *NoopCallerCaller) Noop(opts *bind.CallOpts) error {
	var out []interface{}
	err := _NoopCaller.contract.Call(opts, &out, "noop")

	if err != nil {
		return err
	}

	return err

}

// Noop is a free data retrieval call binding the contract method 0x5dfc2e4a.
//
// Solidity: function noop() view returns()
func (_NoopCaller *NoopCallerSession) Noop() error {
	return _NoopCaller.Contract.Noop(&_NoopCaller.CallOpts)
}

// Noop is a free data retrieval call binding the contract method 0x5dfc2e4a.
//
// Solidity: function noop() view returns()
func (_NoopCaller *NoopCallerCallerSession) Noop() error {
	return _NoopCaller.Contract.Noop(&_NoopCaller.CallOpts)
}

// NoopStaticCall is a free data retrieval call binding the contract method 0xa79ad1a5.
//
// Solidity: function noop_static_call() view returns(bytes)
func (_NoopCaller *NoopCallerCaller) NoopStaticCall(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _NoopCaller.contract.Call(opts, &out, "noop_static_call")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// NoopStaticCall is a free data retrieval call binding the contract method 0xa79ad1a5.
//
// Solidity: function noop_static_call() view returns(bytes)
func (_NoopCaller *NoopCallerSession) NoopStaticCall() ([]byte, error) {
	return _NoopCaller.Contract.NoopStaticCall(&_NoopCaller.CallOpts)
}

// NoopStaticCall is a free data retrieval call binding the contract method 0xa79ad1a5.
//
// Solidity: function noop_static_call() view returns(bytes)
func (_NoopCaller *NoopCallerCallerSession) NoopStaticCall() ([]byte, error) {
	return _NoopCaller.Contract.NoopStaticCall(&_NoopCaller.CallOpts)
}
