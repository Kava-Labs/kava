// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package greeter

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

// GreeterMetaData contains all meta data concerning the Greeter contract.
var GreeterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_greeting\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"greet\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_greeting\",\"type\":\"string\"}],\"name\":\"setGreeting\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162000bf238038062000bf28339818101604052810190620000379190620001e3565b80600090816200004891906200047f565b505062000566565b6000604051905090565b600080fd5b600080fd5b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b620000b9826200006e565b810181811067ffffffffffffffff82111715620000db57620000da6200007f565b5b80604052505050565b6000620000f062000050565b9050620000fe8282620000ae565b919050565b600067ffffffffffffffff8211156200012157620001206200007f565b5b6200012c826200006e565b9050602081019050919050565b60005b83811015620001595780820151818401526020810190506200013c565b60008484015250505050565b60006200017c620001768462000103565b620000e4565b9050828152602081018484840111156200019b576200019a62000069565b5b620001a884828562000139565b509392505050565b600082601f830112620001c857620001c762000064565b5b8151620001da84826020860162000165565b91505092915050565b600060208284031215620001fc57620001fb6200005a565b5b600082015167ffffffffffffffff8111156200021d576200021c6200005f565b5b6200022b84828501620001b0565b91505092915050565b600081519050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806200028757607f821691505b6020821081036200029d576200029c6200023f565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b600060088302620003077fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82620002c8565b620003138683620002c8565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b6000620003606200035a62000354846200032b565b62000335565b6200032b565b9050919050565b6000819050919050565b6200037c836200033f565b620003946200038b8262000367565b848454620002d5565b825550505050565b600090565b620003ab6200039c565b620003b881848462000371565b505050565b5b81811015620003e057620003d4600082620003a1565b600181019050620003be565b5050565b601f8211156200042f57620003f981620002a3565b6200040484620002b8565b8101602085101562000414578190505b6200042c6200042385620002b8565b830182620003bd565b50505b505050565b600082821c905092915050565b6000620004546000198460080262000434565b1980831691505092915050565b60006200046f838362000441565b9150826002028217905092915050565b6200048a8262000234565b67ffffffffffffffff811115620004a657620004a56200007f565b5b620004b282546200026e565b620004bf828285620003e4565b600060209050601f831160018114620004f75760008415620004e2578287015190505b620004ee858262000461565b8655506200055e565b601f1984166200050786620002a3565b60005b8281101562000531578489015182556001820191506020850194506020810190506200050a565b868310156200055157848901516200054d601f89168262000441565b8355505b6001600288020188555050505b505050505050565b61067c80620005766000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c8063a41368621461003b578063cfae321714610057575b600080fd5b61005560048036038101906100509190610274565b610075565b005b61005f610088565b60405161006c919061033c565b60405180910390f35b80600090816100849190610574565b5050565b6060600080546100979061038d565b80601f01602080910402602001604051908101604052809291908181526020018280546100c39061038d565b80156101105780601f106100e557610100808354040283529160200191610110565b820191906000526020600020905b8154815290600101906020018083116100f357829003601f168201915b5050505050905090565b6000604051905090565b600080fd5b600080fd5b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b61018182610138565b810181811067ffffffffffffffff821117156101a05761019f610149565b5b80604052505050565b60006101b361011a565b90506101bf8282610178565b919050565b600067ffffffffffffffff8211156101df576101de610149565b5b6101e882610138565b9050602081019050919050565b82818337600083830152505050565b6000610217610212846101c4565b6101a9565b90508281526020810184848401111561023357610232610133565b5b61023e8482856101f5565b509392505050565b600082601f83011261025b5761025a61012e565b5b813561026b848260208601610204565b91505092915050565b60006020828403121561028a57610289610124565b5b600082013567ffffffffffffffff8111156102a8576102a7610129565b5b6102b484828501610246565b91505092915050565b600081519050919050565b600082825260208201905092915050565b60005b838110156102f75780820151818401526020810190506102dc565b60008484015250505050565b600061030e826102bd565b61031881856102c8565b93506103288185602086016102d9565b61033181610138565b840191505092915050565b600060208201905081810360008301526103568184610303565b905092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806103a557607f821691505b6020821081036103b8576103b761035e565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b6000600883026104207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826103e3565b61042a86836103e3565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b600061047161046c61046784610442565b61044c565b610442565b9050919050565b6000819050919050565b61048b83610456565b61049f61049782610478565b8484546103f0565b825550505050565b600090565b6104b46104a7565b6104bf818484610482565b505050565b5b818110156104e3576104d86000826104ac565b6001810190506104c5565b5050565b601f821115610528576104f9816103be565b610502846103d3565b81016020851015610511578190505b61052561051d856103d3565b8301826104c4565b50505b505050565b600082821c905092915050565b600061054b6000198460080261052d565b1980831691505092915050565b6000610564838361053a565b9150826002028217905092915050565b61057d826102bd565b67ffffffffffffffff81111561059657610595610149565b5b6105a0825461038d565b6105ab8282856104e7565b600060209050601f8311600181146105de57600084156105cc578287015190505b6105d68582610558565b86555061063e565b601f1984166105ec866103be565b60005b82811015610614578489015182556001820191506020850194506020810190506105ef565b86831015610631578489015161062d601f89168261053a565b8355505b6001600288020188555050505b50505050505056fea2646970667358221220c103277e349bc27f877da8c13495470deae31d171e2ad22b71f97202ea17ae9564736f6c63430008130033",
}

// GreeterABI is the input ABI used to generate the binding from.
// Deprecated: Use GreeterMetaData.ABI instead.
var GreeterABI = GreeterMetaData.ABI

// GreeterBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GreeterMetaData.Bin instead.
var GreeterBin = GreeterMetaData.Bin

// DeployGreeter deploys a new Ethereum contract, binding an instance of Greeter to it.
func DeployGreeter(auth *bind.TransactOpts, backend bind.ContractBackend, _greeting string) (common.Address, *types.Transaction, *Greeter, error) {
	parsed, err := GreeterMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GreeterBin), backend, _greeting)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Greeter{GreeterCaller: GreeterCaller{contract: contract}, GreeterTransactor: GreeterTransactor{contract: contract}, GreeterFilterer: GreeterFilterer{contract: contract}}, nil
}

// Greeter is an auto generated Go binding around an Ethereum contract.
type Greeter struct {
	GreeterCaller     // Read-only binding to the contract
	GreeterTransactor // Write-only binding to the contract
	GreeterFilterer   // Log filterer for contract events
}

// GreeterCaller is an auto generated read-only Go binding around an Ethereum contract.
type GreeterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreeterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GreeterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreeterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GreeterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreeterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GreeterSession struct {
	Contract     *Greeter          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GreeterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GreeterCallerSession struct {
	Contract *GreeterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// GreeterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GreeterTransactorSession struct {
	Contract     *GreeterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// GreeterRaw is an auto generated low-level Go binding around an Ethereum contract.
type GreeterRaw struct {
	Contract *Greeter // Generic contract binding to access the raw methods on
}

// GreeterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GreeterCallerRaw struct {
	Contract *GreeterCaller // Generic read-only contract binding to access the raw methods on
}

// GreeterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GreeterTransactorRaw struct {
	Contract *GreeterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGreeter creates a new instance of Greeter, bound to a specific deployed contract.
func NewGreeter(address common.Address, backend bind.ContractBackend) (*Greeter, error) {
	contract, err := bindGreeter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Greeter{GreeterCaller: GreeterCaller{contract: contract}, GreeterTransactor: GreeterTransactor{contract: contract}, GreeterFilterer: GreeterFilterer{contract: contract}}, nil
}

// NewGreeterCaller creates a new read-only instance of Greeter, bound to a specific deployed contract.
func NewGreeterCaller(address common.Address, caller bind.ContractCaller) (*GreeterCaller, error) {
	contract, err := bindGreeter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GreeterCaller{contract: contract}, nil
}

// NewGreeterTransactor creates a new write-only instance of Greeter, bound to a specific deployed contract.
func NewGreeterTransactor(address common.Address, transactor bind.ContractTransactor) (*GreeterTransactor, error) {
	contract, err := bindGreeter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GreeterTransactor{contract: contract}, nil
}

// NewGreeterFilterer creates a new log filterer instance of Greeter, bound to a specific deployed contract.
func NewGreeterFilterer(address common.Address, filterer bind.ContractFilterer) (*GreeterFilterer, error) {
	contract, err := bindGreeter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GreeterFilterer{contract: contract}, nil
}

// bindGreeter binds a generic wrapper to an already deployed contract.
func bindGreeter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GreeterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Greeter *GreeterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Greeter.Contract.GreeterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Greeter *GreeterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Greeter.Contract.GreeterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Greeter *GreeterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Greeter.Contract.GreeterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Greeter *GreeterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Greeter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Greeter *GreeterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Greeter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Greeter *GreeterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Greeter.Contract.contract.Transact(opts, method, params...)
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() view returns(string)
func (_Greeter *GreeterCaller) Greet(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Greeter.contract.Call(opts, &out, "greet")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() view returns(string)
func (_Greeter *GreeterSession) Greet() (string, error) {
	return _Greeter.Contract.Greet(&_Greeter.CallOpts)
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() view returns(string)
func (_Greeter *GreeterCallerSession) Greet() (string, error) {
	return _Greeter.Contract.Greet(&_Greeter.CallOpts)
}

// SetGreeting is a paid mutator transaction binding the contract method 0xa4136862.
//
// Solidity: function setGreeting(string _greeting) returns()
func (_Greeter *GreeterTransactor) SetGreeting(opts *bind.TransactOpts, _greeting string) (*types.Transaction, error) {
	return _Greeter.contract.Transact(opts, "setGreeting", _greeting)
}

// SetGreeting is a paid mutator transaction binding the contract method 0xa4136862.
//
// Solidity: function setGreeting(string _greeting) returns()
func (_Greeter *GreeterSession) SetGreeting(_greeting string) (*types.Transaction, error) {
	return _Greeter.Contract.SetGreeting(&_Greeter.TransactOpts, _greeting)
}

// SetGreeting is a paid mutator transaction binding the contract method 0xa4136862.
//
// Solidity: function setGreeting(string _greeting) returns()
func (_Greeter *GreeterTransactorSession) SetGreeting(_greeting string) (*types.Transaction, error) {
	return _Greeter.Contract.SetGreeting(&_Greeter.TransactOpts, _greeting)
}
