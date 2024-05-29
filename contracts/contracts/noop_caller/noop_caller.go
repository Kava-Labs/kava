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
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"string\",\"name\":\"indexedParam\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"param\",\"type\":\"string\"}],\"name\":\"Event\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"emitEvent\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"noop\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"noop_static_call\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60806040523461002f576100196100146100fa565b610196565b610021610034565b6106056101a4823961060590f35b61003a565b60405190565b600080fd5b601f801991011690565b634e487b7160e01b600052604160045260246000fd5b906100699061003f565b810190811060018060401b0382111761008157604052565b610049565b90610099610092610034565b928361005f565b565b600080fd5b60018060a01b031690565b6100b4906100a0565b90565b6100c0816100ab565b036100c757565b600080fd5b905051906100d9826100b7565b565b906020828203126100f5576100f2916000016100cc565b90565b61009b565b6101186107a98038038061010d81610086565b9283398101906100db565b90565b60001b90565b9061013260018060a01b039161011b565b9181191691161790565b90565b61015361014e610158926100a0565b61013c565b6100a0565b90565b6101649061013f565b90565b6101709061015b565b90565b90565b9061018b61018661019292610167565b610173565b8254610121565b9055565b6101a1906000610176565b56fe60806040526004361015610013575b61019a565b61001e60003561004d565b80635dfc2e4a146100485780637b0cb839146100435763a79ad1a50361000e57610165565b6100ac565b610079565b60e01c90565b60405190565b600080fd5b600080fd5b600091031261006e57565b61005e565b60000190565b346100a757610089366004610063565b610091610296565b610099610053565b806100a381610073565b0390f35b610059565b346100da576100bc366004610063565b6100c46103fb565b6100cc610053565b806100d681610073565b0390f35b610059565b5190565b60209181520190565b60005b838110610100575050906000910152565b8060209183015181850152016100ef565b601f801991011690565b61013a61014360209361014893610131816100df565b938480936100e3565b958691016100ec565b610111565b0190565b610162916020820191600081840391015261011b565b90565b3461019557610175366004610063565b610191610180610565565b610188610053565b9182918261014c565b0390f35b610059565b600080fd5b60001c90565b60018060a01b031690565b6101bc6101c19161019f565b6101a5565b90565b6101ce90546101b0565b90565b60018060a01b031690565b90565b6101f36101ee6101f8926101d1565b6101dc565b6101d1565b90565b610204906101df565b90565b610210906101fb565b90565b61021c906101df565b90565b61022890610213565b90565b600080fd5b634e487b7160e01b600052604160045260246000fd5b9061025090610111565b810190811067ffffffffffffffff82111761026a57604052565b610230565b60e01b90565b600091031261028057565b61005e565b61028d610053565b3d6000823e3d90fd5b6102b06102ab6102a660006101c4565b610207565b61021f565b635dfc2e4a90803b15610328576102d4916000916102cc610053565b93849261026f565b825281806102e460048201610073565b03915afa8015610323576102f6575b50565b6103169060003d811161031c575b61030e8183610246565b810190610275565b386102f3565b503d610304565b610285565b61022b565b905090565b60007f746573742d696e64657865642d706172616d0000000000000000000000000000910152565b6103666012809261032d565b61036f81610332565b0190565b61037c9061035a565b90565b610387610053565b8061039181610373565b03902090565b60209181520190565b60007f746573742d706172616d00000000000000000000000000000000000000000000910152565b6103d5600a602092610397565b6103de816103a0565b0190565b6103f890602081019060008183039101526103c8565b90565b7f39b8d23135cdeca3f85b347e5285f40c9b1de764cf9f8126e7f3b34d77ff0cf061042461037f565b9061042d610053565b80610437816103e2565b0390a2565b606090565b9061045461044d610053565b9283610246565b565b67ffffffffffffffff811161047457610470602091610111565b0190565b610230565b9061048b61048683610456565b610441565b918252565b3d6000146104ad576104a13d610479565b903d6000602084013e5b565b6104b561043c565b906104ab565b60207f6c65640000000000000000000000000000000000000000000000000000000000917f63616c6c20746f20707265636f6d70696c656420636f6e74726163742066616960008201520152565b6105166023604092610397565b61051f816104bb565b0190565b6105399060208101906000818303910152610509565b90565b1561054357565b61054b610053565b62461bcd60e51b81528061056160048201610523565b0390fd5b61056d61043c565b5060008060046105a8632efe172560e11b610599610589610053565b9384926020840190815201610073565b60208201810382520382610246565b6105b1826101c4565b90602081019051915afa6105cc6105c6610490565b9161053c565b9056fea2646970667358221220e4d3cbed0cd366decd5e1ca30677504172ecd37e64c24cdea36ea8f5e08b221d64736f6c63430008190033",
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

// EmitEvent is a paid mutator transaction binding the contract method 0x7b0cb839.
//
// Solidity: function emitEvent() returns()
func (_NoopCaller *NoopCallerTransactor) EmitEvent(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _NoopCaller.contract.Transact(opts, "emitEvent")
}

// EmitEvent is a paid mutator transaction binding the contract method 0x7b0cb839.
//
// Solidity: function emitEvent() returns()
func (_NoopCaller *NoopCallerSession) EmitEvent() (*types.Transaction, error) {
	return _NoopCaller.Contract.EmitEvent(&_NoopCaller.TransactOpts)
}

// EmitEvent is a paid mutator transaction binding the contract method 0x7b0cb839.
//
// Solidity: function emitEvent() returns()
func (_NoopCaller *NoopCallerTransactorSession) EmitEvent() (*types.Transaction, error) {
	return _NoopCaller.Contract.EmitEvent(&_NoopCaller.TransactOpts)
}

// NoopCallerEventIterator is returned from FilterEvent and is used to iterate over the raw logs and unpacked data for Event events raised by the NoopCaller contract.
type NoopCallerEventIterator struct {
	Event *NoopCallerEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *NoopCallerEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(NoopCallerEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(NoopCallerEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *NoopCallerEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *NoopCallerEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// NoopCallerEvent represents a Event event raised by the NoopCaller contract.
type NoopCallerEvent struct {
	IndexedParam common.Hash
	Param        string
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterEvent is a free log retrieval operation binding the contract event 0x39b8d23135cdeca3f85b347e5285f40c9b1de764cf9f8126e7f3b34d77ff0cf0.
//
// Solidity: event Event(string indexed indexedParam, string param)
func (_NoopCaller *NoopCallerFilterer) FilterEvent(opts *bind.FilterOpts, indexedParam []string) (*NoopCallerEventIterator, error) {

	var indexedParamRule []interface{}
	for _, indexedParamItem := range indexedParam {
		indexedParamRule = append(indexedParamRule, indexedParamItem)
	}

	logs, sub, err := _NoopCaller.contract.FilterLogs(opts, "Event", indexedParamRule)
	if err != nil {
		return nil, err
	}
	return &NoopCallerEventIterator{contract: _NoopCaller.contract, event: "Event", logs: logs, sub: sub}, nil
}

// WatchEvent is a free log subscription operation binding the contract event 0x39b8d23135cdeca3f85b347e5285f40c9b1de764cf9f8126e7f3b34d77ff0cf0.
//
// Solidity: event Event(string indexed indexedParam, string param)
func (_NoopCaller *NoopCallerFilterer) WatchEvent(opts *bind.WatchOpts, sink chan<- *NoopCallerEvent, indexedParam []string) (event.Subscription, error) {

	var indexedParamRule []interface{}
	for _, indexedParamItem := range indexedParam {
		indexedParamRule = append(indexedParamRule, indexedParamItem)
	}

	logs, sub, err := _NoopCaller.contract.WatchLogs(opts, "Event", indexedParamRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(NoopCallerEvent)
				if err := _NoopCaller.contract.UnpackLog(event, "Event", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEvent is a log parse operation binding the contract event 0x39b8d23135cdeca3f85b347e5285f40c9b1de764cf9f8126e7f3b34d77ff0cf0.
//
// Solidity: event Event(string indexed indexedParam, string param)
func (_NoopCaller *NoopCallerFilterer) ParseEvent(log types.Log) (*NoopCallerEvent, error) {
	event := new(NoopCallerEvent)
	if err := _NoopCaller.contract.UnpackLog(event, "Event", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
