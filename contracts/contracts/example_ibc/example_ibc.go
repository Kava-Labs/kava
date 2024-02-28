// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package example_ibc

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

// ExampleIbcMetaData contains all meta data concerning the ExampleIbc contract.
var ExampleIbcMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transferCosmosDenomCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"kavaERC20Address\",\"type\":\"string\"}],\"name\":\"transferERC20Call\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transferKavaCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
	Bin: "0x60806040523461001f57610011610024565b610b2d6100308239610b2d90f35b61002a565b60405190565b600080fdfe60806040526004361015610013575b6104d4565b61001e60003561004d565b80631a3c0bc114610048578063413f03411461004357636a47f2710361000e5761048e565b610337565b6101c2565b60e01c90565b60405190565b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b909182601f830112156100ac5781359167ffffffffffffffff83116100a75760200192600183028401116100a257565b61006d565b610068565b610063565b67ffffffffffffffff1690565b6100c7816100b1565b036100ce57565b600080fd5b905035906100e0826100be565b565b909160e0828403126101b757600082013567ffffffffffffffff81116101b2578361010e918401610072565b929093602082013567ffffffffffffffff81116101ad5781610131918401610072565b929093604082013567ffffffffffffffff81116101a85783610154918401610072565b92909361016481606084016100d3565b9261017282608085016100d3565b926101808360a083016100d3565b9260c082013567ffffffffffffffff81116101a35761019f9201610072565b9091565b61005e565b61005e565b61005e565b61005e565b610059565b60000190565b6101e26101d03660046100e2565b999890989791979692969593956107e9565b6101ea610053565b806101f4816101bc565b0390f35b600080fd5b90565b610209816101fd565b0361021057565b600080fd5b9050359061022282610200565b565b9190916101208184031261033257600081013567ffffffffffffffff811161032d5783610252918301610072565b929093602083013567ffffffffffffffff81116103285781610275918501610072565b9290936102858360408301610215565b92606082013567ffffffffffffffff811161032357816102a6918401610072565b9290936102b683608084016100d3565b926102c48160a085016100d3565b926102d28260c083016100d3565b9260e082013567ffffffffffffffff811161031e57836102f3918401610072565b92909361010082013567ffffffffffffffff8111610319576103159201610072565b9091565b61005e565b61005e565b61005e565b61005e565b61005e565b610059565b346103785761036261034a366004610224565b9c9b909b9a919a99929998939897949796959661092a565b61036a610053565b80610374816101bc565b0390f35b6101f8565b916101208383031261048957600083013567ffffffffffffffff811161048457826103a9918501610072565b929093602081013567ffffffffffffffff811161047f57826103cc918301610072565b929093604083013567ffffffffffffffff811161047a57826103ef918501610072565b9290936103ff8260608301610215565b92608082013567ffffffffffffffff81116104755783610420918401610072565b9290936104308160a084016100d3565b9261043e8260c085016100d3565b9261044c8360e083016100d3565b9261010082013567ffffffffffffffff81116104705761046c9201610072565b9091565b61005e565b61005e565b61005e565b61005e565b61005e565b610059565b346104cf576104b96104a136600461037d565b9c9b909b9a919a999299989398979497969596610a63565b6104c1610053565b806104cb816101bc565b0390f35b6101f8565b600080fd5b60209181520190565b90826000939282370152565b601f801991011690565b91906105128161050b81610517956104d9565b80956104e2565b6104ee565b0190565b610524906100b1565b9052565b96999792946105989661056b6105a59d9b9761058e9761055c8c610584986105799860e083019260008185039101526104f8565b918c60208185039101526104f8565b9189830360408b01526104f8565b98606087019061051b565b608085019061051b565b60a083019061051b565b60c08185039101526104f8565b90565b634e487b7160e01b600052604160045260246000fd5b906105c8906104ee565b810190811067ffffffffffffffff8211176105e257604052565b6105a8565b6002600360981b0190565b906106056105fe610053565b92836105be565b565b67ffffffffffffffff8111610625576106216020916104ee565b0190565b6105a8565b9061063c61063783610607565b6105f2565b918252565b606090565b3d600014610663576106573d61062a565b903d6000602084013e5b565b61066b610641565b90610661565b90565b905090565b60207f743a200000000000000000000000000000000000000000000000000000000000917f63616c6c20746f20707265636f6d70696c65206661696c65642c206f7574707560008201520152565b6106d360238092610674565b6106dc81610679565b0190565b5190565b60005b8381106106f8575050906000910152565b8060209183015181850152016106e7565b61072e6107259260209261071c816106e0565b94858093610674565b938491016106e4565b0190565b9061073f610745926106c7565b90610709565b90565b90610770610754610053565b8093610764602083019182610732565b908103825203836105be565b565b61079161079a60209361079f93610788816106e0565b938480936104d9565b958691016106e4565b6104ee565b0190565b6107b99160208201916000818403910152610772565b90565b156107c45750565b6107e5906107d0610053565b91829162461bcd60e51b8352600483016107a3565b0390fd5b989490979399959199969296630119704160e61b99989a979091929394959697610811610053565b9b8c9b60208d019081526004019a6108289b610528565b60208201810382520361083b90826105be565b6108436105e7565b3491602081019051905a93600094938594f161085d610646565b61086690610671565b61086f90610748565b610878916107bc565b565b610883906101fd565b9052565b9c9a96946109189a956109279f9d999a94978f6108df6108d461090b9b6109019a6108c66108ec996108f79b61012088019188830360008a01526104f8565b9185830360208701526104f8565b94604083019061087a565b60608185039101526104f8565b9760808d019061051b565b60a08b019061051b565b60c089019061051b565b86830360e08801526104f8565b926101008185039101526104f8565b90565b9b97939c9894909a96929995919c633dd6d95160e21b9c9b9d9a909192939495969798999a610957610053565b9e8f9e8f6020019081526004019d61096e9e610887565b60208201810382520361098190826105be565b6109896105e7565b6020820191515a9260008094938194f16109a1610646565b6109aa90610671565b6109b390610748565b6109bc916107bc565b565b999c9a95610a1b610a609f9d9992610a2692610a3e98610a0d8f9398610a529f989a6109ff610a489f9b610a339d61012089019189830360008b01526104f8565b9186830360208801526104f8565b9260408185039101526104f8565b9360608c019061087a565b89830360808b01526104f8565b9860a087019061051b565b60c085019061051b565b60e083019061051b565b6101008185039101526104f8565b90565b9b97939c9894909a96929995919c633a9a241160e11b9c9b9d9a909192939495969798999a610a90610053565b9e8f9e8f6020019081526004019d610aa79e6109be565b602082018103825203610aba90826105be565b610ac26105e7565b6020820191515a9260008094938194f1610ada610646565b610ae390610671565b610aec90610748565b610af5916107bc565b56fea2646970667358221220efc2b214bc9c8a6136517094d2a5f7f641d19f473eb099d2137ca2de8ad7cef864736f6c63430008170033",
}

// ExampleIbcABI is the input ABI used to generate the binding from.
// Deprecated: Use ExampleIbcMetaData.ABI instead.
var ExampleIbcABI = ExampleIbcMetaData.ABI

// ExampleIbcBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ExampleIbcMetaData.Bin instead.
var ExampleIbcBin = ExampleIbcMetaData.Bin

// DeployExampleIbc deploys a new Ethereum contract, binding an instance of ExampleIbc to it.
func DeployExampleIbc(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ExampleIbc, error) {
	parsed, err := ExampleIbcMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ExampleIbcBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ExampleIbc{ExampleIbcCaller: ExampleIbcCaller{contract: contract}, ExampleIbcTransactor: ExampleIbcTransactor{contract: contract}, ExampleIbcFilterer: ExampleIbcFilterer{contract: contract}}, nil
}

// ExampleIbc is an auto generated Go binding around an Ethereum contract.
type ExampleIbc struct {
	ExampleIbcCaller     // Read-only binding to the contract
	ExampleIbcTransactor // Write-only binding to the contract
	ExampleIbcFilterer   // Log filterer for contract events
}

// ExampleIbcCaller is an auto generated read-only Go binding around an Ethereum contract.
type ExampleIbcCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExampleIbcTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ExampleIbcTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExampleIbcFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ExampleIbcFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExampleIbcSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ExampleIbcSession struct {
	Contract     *ExampleIbc       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ExampleIbcCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ExampleIbcCallerSession struct {
	Contract *ExampleIbcCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ExampleIbcTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ExampleIbcTransactorSession struct {
	Contract     *ExampleIbcTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ExampleIbcRaw is an auto generated low-level Go binding around an Ethereum contract.
type ExampleIbcRaw struct {
	Contract *ExampleIbc // Generic contract binding to access the raw methods on
}

// ExampleIbcCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ExampleIbcCallerRaw struct {
	Contract *ExampleIbcCaller // Generic read-only contract binding to access the raw methods on
}

// ExampleIbcTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ExampleIbcTransactorRaw struct {
	Contract *ExampleIbcTransactor // Generic write-only contract binding to access the raw methods on
}

// NewExampleIbc creates a new instance of ExampleIbc, bound to a specific deployed contract.
func NewExampleIbc(address common.Address, backend bind.ContractBackend) (*ExampleIbc, error) {
	contract, err := bindExampleIbc(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ExampleIbc{ExampleIbcCaller: ExampleIbcCaller{contract: contract}, ExampleIbcTransactor: ExampleIbcTransactor{contract: contract}, ExampleIbcFilterer: ExampleIbcFilterer{contract: contract}}, nil
}

// NewExampleIbcCaller creates a new read-only instance of ExampleIbc, bound to a specific deployed contract.
func NewExampleIbcCaller(address common.Address, caller bind.ContractCaller) (*ExampleIbcCaller, error) {
	contract, err := bindExampleIbc(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ExampleIbcCaller{contract: contract}, nil
}

// NewExampleIbcTransactor creates a new write-only instance of ExampleIbc, bound to a specific deployed contract.
func NewExampleIbcTransactor(address common.Address, transactor bind.ContractTransactor) (*ExampleIbcTransactor, error) {
	contract, err := bindExampleIbc(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ExampleIbcTransactor{contract: contract}, nil
}

// NewExampleIbcFilterer creates a new log filterer instance of ExampleIbc, bound to a specific deployed contract.
func NewExampleIbcFilterer(address common.Address, filterer bind.ContractFilterer) (*ExampleIbcFilterer, error) {
	contract, err := bindExampleIbc(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ExampleIbcFilterer{contract: contract}, nil
}

// bindExampleIbc binds a generic wrapper to an already deployed contract.
func bindExampleIbc(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ExampleIbcMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ExampleIbc *ExampleIbcRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ExampleIbc.Contract.ExampleIbcCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ExampleIbc *ExampleIbcRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ExampleIbc.Contract.ExampleIbcTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ExampleIbc *ExampleIbcRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ExampleIbc.Contract.ExampleIbcTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ExampleIbc *ExampleIbcCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ExampleIbc.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ExampleIbc *ExampleIbcTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ExampleIbc.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ExampleIbc *ExampleIbcTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ExampleIbc.Contract.contract.Transact(opts, method, params...)
}

// TransferCosmosDenomCall is a paid mutator transaction binding the contract method 0x6a47f271.
//
// Solidity: function transferCosmosDenomCall(string sourcePort, string sourceChannel, string denom, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns()
func (_ExampleIbc *ExampleIbcTransactor) TransferCosmosDenomCall(opts *bind.TransactOpts, sourcePort string, sourceChannel string, denom string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _ExampleIbc.contract.Transact(opts, "transferCosmosDenomCall", sourcePort, sourceChannel, denom, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// TransferCosmosDenomCall is a paid mutator transaction binding the contract method 0x6a47f271.
//
// Solidity: function transferCosmosDenomCall(string sourcePort, string sourceChannel, string denom, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns()
func (_ExampleIbc *ExampleIbcSession) TransferCosmosDenomCall(sourcePort string, sourceChannel string, denom string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferCosmosDenomCall(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, denom, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// TransferCosmosDenomCall is a paid mutator transaction binding the contract method 0x6a47f271.
//
// Solidity: function transferCosmosDenomCall(string sourcePort, string sourceChannel, string denom, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) returns()
func (_ExampleIbc *ExampleIbcTransactorSession) TransferCosmosDenomCall(sourcePort string, sourceChannel string, denom string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferCosmosDenomCall(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, denom, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// TransferERC20Call is a paid mutator transaction binding the contract method 0x413f0341.
//
// Solidity: function transferERC20Call(string sourcePort, string sourceChannel, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo, string kavaERC20Address) returns()
func (_ExampleIbc *ExampleIbcTransactor) TransferERC20Call(opts *bind.TransactOpts, sourcePort string, sourceChannel string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string, kavaERC20Address string) (*types.Transaction, error) {
	return _ExampleIbc.contract.Transact(opts, "transferERC20Call", sourcePort, sourceChannel, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo, kavaERC20Address)
}

// TransferERC20Call is a paid mutator transaction binding the contract method 0x413f0341.
//
// Solidity: function transferERC20Call(string sourcePort, string sourceChannel, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo, string kavaERC20Address) returns()
func (_ExampleIbc *ExampleIbcSession) TransferERC20Call(sourcePort string, sourceChannel string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string, kavaERC20Address string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferERC20Call(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo, kavaERC20Address)
}

// TransferERC20Call is a paid mutator transaction binding the contract method 0x413f0341.
//
// Solidity: function transferERC20Call(string sourcePort, string sourceChannel, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo, string kavaERC20Address) returns()
func (_ExampleIbc *ExampleIbcTransactorSession) TransferERC20Call(sourcePort string, sourceChannel string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string, kavaERC20Address string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferERC20Call(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo, kavaERC20Address)
}

// TransferKavaCall is a paid mutator transaction binding the contract method 0x1a3c0bc1.
//
// Solidity: function transferKavaCall(string sourcePort, string sourceChannel, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) payable returns()
func (_ExampleIbc *ExampleIbcTransactor) TransferKavaCall(opts *bind.TransactOpts, sourcePort string, sourceChannel string, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _ExampleIbc.contract.Transact(opts, "transferKavaCall", sourcePort, sourceChannel, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// TransferKavaCall is a paid mutator transaction binding the contract method 0x1a3c0bc1.
//
// Solidity: function transferKavaCall(string sourcePort, string sourceChannel, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) payable returns()
func (_ExampleIbc *ExampleIbcSession) TransferKavaCall(sourcePort string, sourceChannel string, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferKavaCall(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}

// TransferKavaCall is a paid mutator transaction binding the contract method 0x1a3c0bc1.
//
// Solidity: function transferKavaCall(string sourcePort, string sourceChannel, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo) payable returns()
func (_ExampleIbc *ExampleIbcTransactorSession) TransferKavaCall(sourcePort string, sourceChannel string, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferKavaCall(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo)
}
