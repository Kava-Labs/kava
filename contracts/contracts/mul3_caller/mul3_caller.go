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
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"c\",\"type\":\"uint256\"}],\"name\":\"calcMul3\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"c\",\"type\":\"uint256\"}],\"name\":\"calcMul3Call\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"a\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"b\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"c\",\"type\":\"uint256\"}],\"name\":\"calcMul3WithError\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMul3\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"result\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMul3StaticCall\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234601c57600e6020565b6107f461002c82396107f490f35b6026565b60405190565b600080fdfe60806040526004361015610013575b6102ac565b61001e60003561006d565b80632eb84fe2146100685780634292eea214610063578063446a94a41461005e5780635c8062c2146100595763dbb45c000361000e57610278565b610242565b61020d565b610153565b6100b9565b60e01c90565b60405190565b600080fd5b600080fd5b600091031261008e57565b61007e565b90565b61009f90610093565b9052565b91906100b790600060208501940190610096565b565b346100e9576100c9366004610083565b6100e56100d46103a1565b6100dc610073565b918291826100a3565b0390f35b610079565b6100f781610093565b036100fe57565b600080fd5b90503590610110826100ee565b565b90916060828403126101485761014561012e8460008501610103565b9361013c8160208601610103565b93604001610103565b90565b61007e565b60000190565b346101825761016c610166366004610112565b9161047a565b610174610073565b8061017e8161014d565b0390f35b610079565b5190565b60209181520190565b60005b8381106101a8575050906000910152565b806020918301518185015201610197565b601f801991011690565b6101e26101eb6020936101f0936101d981610187565b9384809361018b565b95869101610194565b6101b9565b0190565b61020a91602082019160008184039101526101c3565b90565b3461023d5761021d366004610083565b610239610228610648565b610230610073565b918291826101f4565b0390f35b610079565b346102735761026f61025e610258366004610112565b916106b1565b610266610073565b918291826101f4565b0390f35b610079565b346102a75761029161028b366004610112565b91610722565b610299610073565b806102a38161014d565b0390f35b610079565b600080fd5b600090565b600360981b90565b60018060a01b031690565b90565b6102e06102db6102e5926102be565b6102c9565b6102be565b90565b6102f1906102cc565b90565b6102fd906102e8565b90565b610309906102cc565b90565b61031590610300565b90565b600080fd5b634e487b7160e01b600052604160045260246000fd5b9061033d906101b9565b810190811067ffffffffffffffff82111761035757604052565b61031d565b60e01b90565b9050519061036f826100ee565b565b9060208282031261038b5761038891600001610362565b90565b61007e565b610398610073565b3d6000823e3d90fd5b6103a96102b1565b506103dd60206103c76103c26103bd6102b6565b6102f4565b61030c565b632eb84fe2906103d5610073565b93849261035c565b825281806103ed6004820161014d565b03915afa90811561043257600091610404575b5090565b610425915060203d811161042b575b61041d8183610333565b810190610371565b38610400565b503d610413565b610390565b600091031261044257565b61007e565b604090610471610478949695939661046760608401986000850190610096565b6020830190610096565b0190610096565b565b9061049361048e6104896102b6565b6102f4565b61030c565b634292eea292919392813b156105115760006104c2916104cd82966104b6610073565b9889978896879561035c565b855260048501610447565b03925af1801561050c576104df575b50565b6104ff9060003d8111610505575b6104f78183610333565b810190610437565b386104dc565b503d6104ed565b610390565b610318565b606090565b9061052e610527610073565b9283610333565b565b67ffffffffffffffff811161054e5761054a6020916101b9565b0190565b61031d565b9061056561056083610530565b61051b565b918252565b3d6000146105875761057b3d610553565b903d6000602084013e5b565b61058f610516565b90610585565b60209181520190565b60207f6c65640000000000000000000000000000000000000000000000000000000000917f63616c6c20746f20707265636f6d70696c656420636f6e74726163742066616960008201520152565b6105f96023604092610595565b6106028161059e565b0190565b61061c90602081019060008183039101526105ec565b90565b1561062657565b61062e610073565b62461bcd60e51b81528061064460048201610606565b0390fd5b610650610516565b50600080600461068b63175c27f160e11b61067c61066c610073565b938492602084019081520161014d565b60208201810382520382610333565b6106936102b6565b90602081019051915afa6106ae6106a861056a565b9161061f565b90565b6106fb6004916106ec6000959486956106c8610516565b506236ed1760ea1b939190916106dc610073565b9687956020870190815201610447565b60208201810382520382610333565b6107036102b6565b9082602082019151925af161071f61071961056a565b9161061f565b90565b9061073b6107366107316102b6565b6102f4565b61030c565b63dbb45c0092919392813b156107b957600061076a91610775829661075e610073565b9889978896879561035c565b855260048501610447565b03925af180156107b457610787575b50565b6107a79060003d81116107ad575b61079f8183610333565b810190610437565b38610784565b503d610795565b610390565b61031856fea2646970667358221220e88155fe2561fb263a0c33b3cf460f278727610dce322e70605ef19d0811349064736f6c63430008190033",
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

// CalcMul3WithError is a paid mutator transaction binding the contract method 0x4292eea2.
//
// Solidity: function calcMul3WithError(uint256 a, uint256 b, uint256 c) returns()
func (_Mul3Caller *Mul3CallerTransactor) CalcMul3WithError(opts *bind.TransactOpts, a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.contract.Transact(opts, "calcMul3WithError", a, b, c)
}

// CalcMul3WithError is a paid mutator transaction binding the contract method 0x4292eea2.
//
// Solidity: function calcMul3WithError(uint256 a, uint256 b, uint256 c) returns()
func (_Mul3Caller *Mul3CallerSession) CalcMul3WithError(a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.Contract.CalcMul3WithError(&_Mul3Caller.TransactOpts, a, b, c)
}

// CalcMul3WithError is a paid mutator transaction binding the contract method 0x4292eea2.
//
// Solidity: function calcMul3WithError(uint256 a, uint256 b, uint256 c) returns()
func (_Mul3Caller *Mul3CallerTransactorSession) CalcMul3WithError(a *big.Int, b *big.Int, c *big.Int) (*types.Transaction, error) {
	return _Mul3Caller.Contract.CalcMul3WithError(&_Mul3Caller.TransactOpts, a, b, c)
}
