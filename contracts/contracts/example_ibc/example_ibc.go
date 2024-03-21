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
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transferCosmosDenomCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"kavaERC20Address\",\"type\":\"string\"}],\"name\":\"transferERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"kavaERC20Address\",\"type\":\"string\"}],\"name\":\"transferERC20Call\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"memo\",\"type\":\"string\"}],\"name\":\"transferKavaCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
	Bin: "0x60806040523461001f57610011610024565b610cbd6100308239610cbd90f35b61002a565b60405190565b600080fdfe60806040526004361015610013575b61052a565b61001e60003561005d565b80631a3c0bc114610058578063413f0341146100535780636a47f2711461004e5763f75b65440361000e576104e4565b61049e565b610347565b6101d2565b60e01c90565b60405190565b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b909182601f830112156100bc5781359167ffffffffffffffff83116100b75760200192600183028401116100b257565b61007d565b610078565b610073565b67ffffffffffffffff1690565b6100d7816100c1565b036100de57565b600080fd5b905035906100f0826100ce565b565b909160e0828403126101c757600082013567ffffffffffffffff81116101c2578361011e918401610082565b929093602082013567ffffffffffffffff81116101bd5781610141918401610082565b929093604082013567ffffffffffffffff81116101b85783610164918401610082565b92909361017481606084016100e3565b9261018282608085016100e3565b926101908360a083016100e3565b9260c082013567ffffffffffffffff81116101b3576101af9201610082565b9091565b61006e565b61006e565b61006e565b61006e565b610069565b60000190565b6101f26101e03660046100f2565b9998909897919796929695939561083f565b6101fa610063565b80610204816101cc565b0390f35b600080fd5b90565b6102198161020d565b0361022057565b600080fd5b9050359061023282610210565b565b9190916101208184031261034257600081013567ffffffffffffffff811161033d5783610262918301610082565b929093602083013567ffffffffffffffff81116103385781610285918501610082565b9290936102958360408301610225565b92606082013567ffffffffffffffff811161033357816102b6918401610082565b9290936102c683608084016100e3565b926102d48160a085016100e3565b926102e28260c083016100e3565b9260e082013567ffffffffffffffff811161032e5783610303918401610082565b92909361010082013567ffffffffffffffff8111610329576103259201610082565b9091565b61006e565b61006e565b61006e565b61006e565b61006e565b610069565b346103885761037261035a366004610234565b9c9b909b9a919a999299989398979497969596610980565b61037a610063565b80610384816101cc565b0390f35b610208565b916101208383031261049957600083013567ffffffffffffffff811161049457826103b9918501610082565b929093602081013567ffffffffffffffff811161048f57826103dc918301610082565b929093604083013567ffffffffffffffff811161048a57826103ff918501610082565b92909361040f8260608301610225565b92608082013567ffffffffffffffff81116104855783610430918401610082565b9290936104408160a084016100e3565b9261044e8260c085016100e3565b9261045c8360e083016100e3565b9261010082013567ffffffffffffffff81116104805761047c9201610082565b9091565b61006e565b61006e565b61006e565b61006e565b61006e565b610069565b346104df576104c96104b136600461038d565b9c9b909b9a919a999299989398979497969596610ab9565b6104d1610063565b806104db816101cc565b0390f35b610208565b346105255761050f6104f7366004610234565b9c9b909b9a919a999299989398979497969596610bd3565b610517610063565b80610521816101cc565b0390f35b610208565b600080fd5b60209181520190565b90826000939282370152565b601f801991011690565b9190610568816105618161056d9561052f565b8095610538565b610544565b0190565b61057a906100c1565b9052565b96999792946105ee966105c16105fb9d9b976105e4976105b28c6105da986105cf9860e0830192600081850391015261054e565b918c602081850391015261054e565b9189830360408b015261054e565b986060870190610571565b6080850190610571565b60a0830190610571565b60c081850391015261054e565b90565b634e487b7160e01b600052604160045260246000fd5b9061061e90610544565b810190811067ffffffffffffffff82111761063857604052565b6105fe565b6002600360981b0190565b9061065b610654610063565b9283610614565b565b67ffffffffffffffff811161067b57610677602091610544565b0190565b6105fe565b9061069261068d8361065d565b610648565b918252565b606090565b3d6000146106b9576106ad3d610680565b903d6000602084013e5b565b6106c1610697565b906106b7565b90565b905090565b60207f743a200000000000000000000000000000000000000000000000000000000000917f63616c6c20746f20707265636f6d70696c65206661696c65642c206f7574707560008201520152565b610729602380926106ca565b610732816106cf565b0190565b5190565b60005b83811061074e575050906000910152565b80602091830151818501520161073d565b61078461077b9260209261077281610736565b948580936106ca565b9384910161073a565b0190565b9061079561079b9261071d565b9061075f565b90565b906107c66107aa610063565b80936107ba602083019182610788565b90810382520383610614565b565b6107e76107f06020936107f5936107de81610736565b9384809361052f565b9586910161073a565b610544565b0190565b61080f91602082019160008184039101526107c8565b90565b1561081a5750565b61083b90610826610063565b91829162461bcd60e51b8352600483016107f9565b0390fd5b989490979399959199969296630119704160e61b99989a979091929394959697610867610063565b9b8c9b60208d019081526004019a61087e9b61057e565b6020820181038252036108919082610614565b61089961063d565b3491602081019051905a93600094938594f16108b361069c565b6108bc906106c7565b6108c59061079e565b6108ce91610812565b565b6108d99061020d565b9052565b9c9a969461096e9a9561097d9f9d999a94978f61093561092a6109619b6109579a61091c6109429961094d9b61012088019188830360008a015261054e565b91858303602087015261054e565b9460408301906108d0565b606081850391015261054e565b9760808d0190610571565b60a08b0190610571565b60c0890190610571565b86830360e088015261054e565b9261010081850391015261054e565b90565b9b97939c9894909a96929995919c633dd6d95160e21b9c9b9d9a909192939495969798999a6109ad610063565b9e8f9e8f6020019081526004019d6109c49e6108dd565b6020820181038252036109d79082610614565b6109df61063d565b6020820191515a9260008094938194f16109f761069c565b610a00906106c7565b610a099061079e565b610a1291610812565b565b999c9a95610a71610ab69f9d9992610a7c92610a9498610a638f9398610aa89f989a610a55610a9e9f9b610a899d61012089019189830360008b015261054e565b91868303602088015261054e565b92604081850391015261054e565b9360608c01906108d0565b89830360808b015261054e565b9860a0870190610571565b60c0850190610571565b60e0830190610571565b61010081850391015261054e565b90565b9b97939c9894909a96929995919c633a9a241160e11b9c9b9d9a909192939495969798999a610ae6610063565b9e8f9e8f6020019081526004019d610afd9e610a14565b602082018103825203610b109082610614565b610b1861063d565b6020820191515a9260008094938194f1610b3061069c565b610b39906106c7565b610b429061079e565b610b4b91610812565b565b60018060a01b031690565b90565b610b6f610b6a610b7492610b4d565b610b58565b610b4d565b90565b610b8090610b5b565b90565b610b8c90610b77565b90565b610b9890610b5b565b90565b610ba490610b8f565b90565b600080fd5b60e01b90565b6000910312610bbd57565b610069565b610bca610063565b3d6000823e3d90fd5b9b959c969099939c9a94979198929a610bfa610bf5610bf061063d565b610b83565b610b9b565b9c999b9d989091929394959697988d3b15610c8257610c17610063565b9e8f9d8e9d8e610c2a63f75b6544610bac565b81526004019d610c399e6108dd565b03815a6000948591f18015610c7d57610c50575b50565b610c709060003d8111610c76575b610c688183610614565b810190610bb2565b38610c4d565b503d610c5e565b610bc2565b610ba756fea26469706673582212201329a76351fd95a7f91f9d28051770bf957a3bc552df0dd239faee58f1cbde3b64736f6c63430008170033",
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

// TransferERC20 is a paid mutator transaction binding the contract method 0xf75b6544.
//
// Solidity: function transferERC20(string sourcePort, string sourceChannel, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo, string kavaERC20Address) returns()
func (_ExampleIbc *ExampleIbcTransactor) TransferERC20(opts *bind.TransactOpts, sourcePort string, sourceChannel string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string, kavaERC20Address string) (*types.Transaction, error) {
	return _ExampleIbc.contract.Transact(opts, "transferERC20", sourcePort, sourceChannel, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo, kavaERC20Address)
}

// TransferERC20 is a paid mutator transaction binding the contract method 0xf75b6544.
//
// Solidity: function transferERC20(string sourcePort, string sourceChannel, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo, string kavaERC20Address) returns()
func (_ExampleIbc *ExampleIbcSession) TransferERC20(sourcePort string, sourceChannel string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string, kavaERC20Address string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferERC20(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo, kavaERC20Address)
}

// TransferERC20 is a paid mutator transaction binding the contract method 0xf75b6544.
//
// Solidity: function transferERC20(string sourcePort, string sourceChannel, uint256 amount, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp, string memo, string kavaERC20Address) returns()
func (_ExampleIbc *ExampleIbcTransactorSession) TransferERC20(sourcePort string, sourceChannel string, amount *big.Int, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64, memo string, kavaERC20Address string) (*types.Transaction, error) {
	return _ExampleIbc.Contract.TransferERC20(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, amount, receiver, revisionNumber, revisionHeight, timeoutTimestamp, memo, kavaERC20Address)
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
