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
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"sourcePort\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"sourceChannel\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"denom\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"sender\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"revisionNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"revisionHeight\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"timeoutTimestamp\",\"type\":\"uint64\"}],\"name\":\"ibcTransferCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523461001f57610011610024565b610660610030823961066090f35b61002a565b60405190565b600080fdfe60806040526004361015610013575b61024c565b61001e60003561002d565b63405314fe0361000e57610206565b60e01c90565b60405190565b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b909182601f830112156100915781359167ffffffffffffffff831161008c57602001926001830284011161008757565b610052565b61004d565b610048565b90565b6100a281610096565b036100a957565b600080fd5b905035906100bb82610099565b565b67ffffffffffffffff1690565b6100d3816100bd565b036100da57565b600080fd5b905035906100ec826100ca565b565b91610120838303126101fb57600083013567ffffffffffffffff81116101f6578261011a918501610057565b929093602081013567ffffffffffffffff81116101f1578261013d918301610057565b929093604083013567ffffffffffffffff81116101ec5782610160918501610057565b92909361017082606083016100ae565b92608082013567ffffffffffffffff81116101e75783610191918401610057565b92909360a082013567ffffffffffffffff81116101e257816101b4918401610057565b9290936101df6101c78460c085016100df565b936101d58160e086016100df565b93610100016100df565b90565b610043565b610043565b610043565b610043565b610043565b61003e565b60000190565b34610247576102316102193660046100ee565b9c9b909b9a919a999299989398979497969596610596565b610239610033565b8061024381610200565b0390f35b610039565b600080fd5b60209181520190565b90826000939282370152565b601f801991011690565b919061028a816102838161028f95610251565b809561025a565b610266565b0190565b61029c90610096565b9052565b6102a9906100bd565b9052565b9994979b9e9d9b61031c6103116103539f9a610337996101009f98918f9261034c9f9799610303916103429f9a6103299c6102f59161012089019189830360008b0152610270565b918683036020880152610270565b926040818503910152610270565b9360608d0190610293565b8a830360808c0152610270565b9187830360a0890152610270565b9860c08501906102a0565b60e08301906102a0565b01906102a0565b565b634e487b7160e01b600052604160045260246000fd5b9061037590610266565b810190811067ffffffffffffffff82111761038f57604052565b610355565b6002600360981b0190565b906103b26103ab610033565b928361036b565b565b67ffffffffffffffff81116103d2576103ce602091610266565b0190565b610355565b906103e96103e4836103b4565b61039f565b918252565b606090565b3d600014610410576104043d6103d7565b903d6000602084013e5b565b6104186103ee565b9061040e565b90565b905090565b60207f743a200000000000000000000000000000000000000000000000000000000000917f63616c6c20746f20707265636f6d70696c65206661696c65642c206f7574707560008201520152565b61048060238092610421565b61048981610426565b0190565b5190565b60005b8381106104a5575050906000910152565b806020918301518185015201610494565b6104db6104d2926020926104c98161048d565b94858093610421565b93849101610491565b0190565b906104ec6104f292610474565b906104b6565b90565b9061051d610501610033565b80936105116020830191826104df565b9081038252038361036b565b565b61053e61054760209361054c936105358161048d565b93848093610251565b95869101610491565b610266565b0190565b610566916020820191600081840391015261051f565b90565b156105715750565b6105929061057d610033565b91829162461bcd60e51b835260048301610550565b0390fd5b9b97939c9894909a96929995919c633f74186b60e21b9c9b9d9a909192939495969798999a6105c3610033565b9e8f9e8f6020019081526004019d6105da9e6102ad565b6020820181038252036105ed908261036b565b6105f5610394565b6020820191515a9260008094938194f161060d6103f3565b6106169061041e565b61061f906104f5565b61062891610569565b56fea2646970667358221220ece4e639454630d9181fb34fea938ae5caf8ca0efec579812d6cd2c86411deab64736f6c63430008170033",
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

// IbcTransferCall is a paid mutator transaction binding the contract method 0x405314fe.
//
// Solidity: function ibcTransferCall(string sourcePort, string sourceChannel, string denom, uint256 amount, string sender, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp) returns()
func (_ExampleIbc *ExampleIbcTransactor) IbcTransferCall(opts *bind.TransactOpts, sourcePort string, sourceChannel string, denom string, amount *big.Int, sender string, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64) (*types.Transaction, error) {
	return _ExampleIbc.contract.Transact(opts, "ibcTransferCall", sourcePort, sourceChannel, denom, amount, sender, receiver, revisionNumber, revisionHeight, timeoutTimestamp)
}

// IbcTransferCall is a paid mutator transaction binding the contract method 0x405314fe.
//
// Solidity: function ibcTransferCall(string sourcePort, string sourceChannel, string denom, uint256 amount, string sender, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp) returns()
func (_ExampleIbc *ExampleIbcSession) IbcTransferCall(sourcePort string, sourceChannel string, denom string, amount *big.Int, sender string, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64) (*types.Transaction, error) {
	return _ExampleIbc.Contract.IbcTransferCall(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, denom, amount, sender, receiver, revisionNumber, revisionHeight, timeoutTimestamp)
}

// IbcTransferCall is a paid mutator transaction binding the contract method 0x405314fe.
//
// Solidity: function ibcTransferCall(string sourcePort, string sourceChannel, string denom, uint256 amount, string sender, string receiver, uint64 revisionNumber, uint64 revisionHeight, uint64 timeoutTimestamp) returns()
func (_ExampleIbc *ExampleIbcTransactorSession) IbcTransferCall(sourcePort string, sourceChannel string, denom string, amount *big.Int, sender string, receiver string, revisionNumber uint64, revisionHeight uint64, timeoutTimestamp uint64) (*types.Transaction, error) {
	return _ExampleIbc.Contract.IbcTransferCall(&_ExampleIbc.TransactOpts, sourcePort, sourceChannel, denom, amount, sender, receiver, revisionNumber, revisionHeight, timeoutTimestamp)
}
