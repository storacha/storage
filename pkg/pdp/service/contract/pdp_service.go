// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// PDPServiceMetaData contains all meta data concerning the PDPService contract.
var PDPServiceMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"NO_CHALLENGE_SCHEDULED\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"NO_PROVING_DEADLINE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"UPGRADE_INTERFACE_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"challengeWindow\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"getChallengesPerProof\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"getMaxProvingPeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"initChallengeWindowStart\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_pdpVerifierAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"nextChallengeWindowStart\",\"inputs\":[{\"name\":\"setId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nextProvingPeriod\",\"inputs\":[{\"name\":\"proofSetId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"challengeEpoch\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pdpVerifierAddress\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"possessionProven\",\"inputs\":[{\"name\":\"proofSetId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"challengeCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"proofSetCreated\",\"inputs\":[{\"name\":\"proofSetId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"creator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"proofSetDeleted\",\"inputs\":[{\"name\":\"proofSetId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"deletedLeafCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"provenThisPeriod\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"provingDeadlines\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proxiableUUID\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"rootsAdded\",\"inputs\":[{\"name\":\"proofSetId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"firstAdded\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"rootData\",\"type\":\"tuple[]\",\"internalType\":\"structPDPVerifier.RootData[]\",\"components\":[{\"name\":\"root\",\"type\":\"tuple\",\"internalType\":\"structCids.Cid\",\"components\":[{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"rawSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"rootsScheduledRemove\",\"inputs\":[{\"name\":\"proofSetId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"rootIds\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"thisChallengeWindowStart\",\"inputs\":[{\"name\":\"setId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"upgradeToAndCall\",\"inputs\":[{\"name\":\"newImplementation\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"event\",\"name\":\"FaultRecord\",\"inputs\":[{\"name\":\"proofSetId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"periodsFaulted\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"deadline\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Upgraded\",\"inputs\":[{\"name\":\"implementation\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC1967InvalidImplementation\",\"inputs\":[{\"name\":\"implementation\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC1967NonPayable\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FailedCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"UUPSUnauthorizedCallContext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UUPSUnsupportedProxiableUUID\",\"inputs\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]",
}

// PDPServiceABI is the input ABI used to generate the binding from.
// Deprecated: Use PDPServiceMetaData.ABI instead.
var PDPServiceABI = PDPServiceMetaData.ABI

// PDPService is an auto generated Go binding around an Ethereum contract.
type PDPService struct {
	PDPServiceCaller     // Read-only binding to the contract
	PDPServiceTransactor // Write-only binding to the contract
	PDPServiceFilterer   // Log filterer for contract events
}

// PDPServiceCaller is an auto generated read-only Go binding around an Ethereum contract.
type PDPServiceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PDPServiceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PDPServiceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PDPServiceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PDPServiceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PDPServiceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PDPServiceSession struct {
	Contract     *PDPService       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PDPServiceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PDPServiceCallerSession struct {
	Contract *PDPServiceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// PDPServiceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PDPServiceTransactorSession struct {
	Contract     *PDPServiceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// PDPServiceRaw is an auto generated low-level Go binding around an Ethereum contract.
type PDPServiceRaw struct {
	Contract *PDPService // Generic contract binding to access the raw methods on
}

// PDPServiceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PDPServiceCallerRaw struct {
	Contract *PDPServiceCaller // Generic read-only contract binding to access the raw methods on
}

// PDPServiceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PDPServiceTransactorRaw struct {
	Contract *PDPServiceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPDPService creates a new instance of PDPService, bound to a specific deployed contract.
func NewPDPService(address common.Address, backend bind.ContractBackend) (*PDPService, error) {
	contract, err := bindPDPService(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PDPService{PDPServiceCaller: PDPServiceCaller{contract: contract}, PDPServiceTransactor: PDPServiceTransactor{contract: contract}, PDPServiceFilterer: PDPServiceFilterer{contract: contract}}, nil
}

// NewPDPServiceCaller creates a new read-only instance of PDPService, bound to a specific deployed contract.
func NewPDPServiceCaller(address common.Address, caller bind.ContractCaller) (*PDPServiceCaller, error) {
	contract, err := bindPDPService(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PDPServiceCaller{contract: contract}, nil
}

// NewPDPServiceTransactor creates a new write-only instance of PDPService, bound to a specific deployed contract.
func NewPDPServiceTransactor(address common.Address, transactor bind.ContractTransactor) (*PDPServiceTransactor, error) {
	contract, err := bindPDPService(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PDPServiceTransactor{contract: contract}, nil
}

// NewPDPServiceFilterer creates a new log filterer instance of PDPService, bound to a specific deployed contract.
func NewPDPServiceFilterer(address common.Address, filterer bind.ContractFilterer) (*PDPServiceFilterer, error) {
	contract, err := bindPDPService(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PDPServiceFilterer{contract: contract}, nil
}

// bindPDPService binds a generic wrapper to an already deployed contract.
func bindPDPService(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PDPServiceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PDPService *PDPServiceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PDPService.Contract.PDPServiceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PDPService *PDPServiceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PDPService.Contract.PDPServiceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PDPService *PDPServiceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PDPService.Contract.PDPServiceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PDPService *PDPServiceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PDPService.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PDPService *PDPServiceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PDPService.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PDPService *PDPServiceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PDPService.Contract.contract.Transact(opts, method, params...)
}

// NOCHALLENGESCHEDULED is a free data retrieval call binding the contract method 0x462dd449.
//
// Solidity: function NO_CHALLENGE_SCHEDULED() view returns(uint256)
func (_PDPService *PDPServiceCaller) NOCHALLENGESCHEDULED(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "NO_CHALLENGE_SCHEDULED")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NOCHALLENGESCHEDULED is a free data retrieval call binding the contract method 0x462dd449.
//
// Solidity: function NO_CHALLENGE_SCHEDULED() view returns(uint256)
func (_PDPService *PDPServiceSession) NOCHALLENGESCHEDULED() (*big.Int, error) {
	return _PDPService.Contract.NOCHALLENGESCHEDULED(&_PDPService.CallOpts)
}

// NOCHALLENGESCHEDULED is a free data retrieval call binding the contract method 0x462dd449.
//
// Solidity: function NO_CHALLENGE_SCHEDULED() view returns(uint256)
func (_PDPService *PDPServiceCallerSession) NOCHALLENGESCHEDULED() (*big.Int, error) {
	return _PDPService.Contract.NOCHALLENGESCHEDULED(&_PDPService.CallOpts)
}

// NOPROVINGDEADLINE is a free data retrieval call binding the contract method 0x4ba9ac22.
//
// Solidity: function NO_PROVING_DEADLINE() view returns(uint256)
func (_PDPService *PDPServiceCaller) NOPROVINGDEADLINE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "NO_PROVING_DEADLINE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NOPROVINGDEADLINE is a free data retrieval call binding the contract method 0x4ba9ac22.
//
// Solidity: function NO_PROVING_DEADLINE() view returns(uint256)
func (_PDPService *PDPServiceSession) NOPROVINGDEADLINE() (*big.Int, error) {
	return _PDPService.Contract.NOPROVINGDEADLINE(&_PDPService.CallOpts)
}

// NOPROVINGDEADLINE is a free data retrieval call binding the contract method 0x4ba9ac22.
//
// Solidity: function NO_PROVING_DEADLINE() view returns(uint256)
func (_PDPService *PDPServiceCallerSession) NOPROVINGDEADLINE() (*big.Int, error) {
	return _PDPService.Contract.NOPROVINGDEADLINE(&_PDPService.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_PDPService *PDPServiceCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_PDPService *PDPServiceSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _PDPService.Contract.UPGRADEINTERFACEVERSION(&_PDPService.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_PDPService *PDPServiceCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _PDPService.Contract.UPGRADEINTERFACEVERSION(&_PDPService.CallOpts)
}

// ChallengeWindow is a free data retrieval call binding the contract method 0x861a1412.
//
// Solidity: function challengeWindow() pure returns(uint256)
func (_PDPService *PDPServiceCaller) ChallengeWindow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "challengeWindow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ChallengeWindow is a free data retrieval call binding the contract method 0x861a1412.
//
// Solidity: function challengeWindow() pure returns(uint256)
func (_PDPService *PDPServiceSession) ChallengeWindow() (*big.Int, error) {
	return _PDPService.Contract.ChallengeWindow(&_PDPService.CallOpts)
}

// ChallengeWindow is a free data retrieval call binding the contract method 0x861a1412.
//
// Solidity: function challengeWindow() pure returns(uint256)
func (_PDPService *PDPServiceCallerSession) ChallengeWindow() (*big.Int, error) {
	return _PDPService.Contract.ChallengeWindow(&_PDPService.CallOpts)
}

// GetChallengesPerProof is a free data retrieval call binding the contract method 0x47d3dfe7.
//
// Solidity: function getChallengesPerProof() pure returns(uint64)
func (_PDPService *PDPServiceCaller) GetChallengesPerProof(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "getChallengesPerProof")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetChallengesPerProof is a free data retrieval call binding the contract method 0x47d3dfe7.
//
// Solidity: function getChallengesPerProof() pure returns(uint64)
func (_PDPService *PDPServiceSession) GetChallengesPerProof() (uint64, error) {
	return _PDPService.Contract.GetChallengesPerProof(&_PDPService.CallOpts)
}

// GetChallengesPerProof is a free data retrieval call binding the contract method 0x47d3dfe7.
//
// Solidity: function getChallengesPerProof() pure returns(uint64)
func (_PDPService *PDPServiceCallerSession) GetChallengesPerProof() (uint64, error) {
	return _PDPService.Contract.GetChallengesPerProof(&_PDPService.CallOpts)
}

// GetMaxProvingPeriod is a free data retrieval call binding the contract method 0xf2f12333.
//
// Solidity: function getMaxProvingPeriod() pure returns(uint64)
func (_PDPService *PDPServiceCaller) GetMaxProvingPeriod(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "getMaxProvingPeriod")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetMaxProvingPeriod is a free data retrieval call binding the contract method 0xf2f12333.
//
// Solidity: function getMaxProvingPeriod() pure returns(uint64)
func (_PDPService *PDPServiceSession) GetMaxProvingPeriod() (uint64, error) {
	return _PDPService.Contract.GetMaxProvingPeriod(&_PDPService.CallOpts)
}

// GetMaxProvingPeriod is a free data retrieval call binding the contract method 0xf2f12333.
//
// Solidity: function getMaxProvingPeriod() pure returns(uint64)
func (_PDPService *PDPServiceCallerSession) GetMaxProvingPeriod() (uint64, error) {
	return _PDPService.Contract.GetMaxProvingPeriod(&_PDPService.CallOpts)
}

// InitChallengeWindowStart is a free data retrieval call binding the contract method 0x21918cea.
//
// Solidity: function initChallengeWindowStart() view returns(uint256)
func (_PDPService *PDPServiceCaller) InitChallengeWindowStart(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "initChallengeWindowStart")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// InitChallengeWindowStart is a free data retrieval call binding the contract method 0x21918cea.
//
// Solidity: function initChallengeWindowStart() view returns(uint256)
func (_PDPService *PDPServiceSession) InitChallengeWindowStart() (*big.Int, error) {
	return _PDPService.Contract.InitChallengeWindowStart(&_PDPService.CallOpts)
}

// InitChallengeWindowStart is a free data retrieval call binding the contract method 0x21918cea.
//
// Solidity: function initChallengeWindowStart() view returns(uint256)
func (_PDPService *PDPServiceCallerSession) InitChallengeWindowStart() (*big.Int, error) {
	return _PDPService.Contract.InitChallengeWindowStart(&_PDPService.CallOpts)
}

// NextChallengeWindowStart is a free data retrieval call binding the contract method 0x8bf96d28.
//
// Solidity: function nextChallengeWindowStart(uint256 setId) view returns(uint256)
func (_PDPService *PDPServiceCaller) NextChallengeWindowStart(opts *bind.CallOpts, setId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "nextChallengeWindowStart", setId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextChallengeWindowStart is a free data retrieval call binding the contract method 0x8bf96d28.
//
// Solidity: function nextChallengeWindowStart(uint256 setId) view returns(uint256)
func (_PDPService *PDPServiceSession) NextChallengeWindowStart(setId *big.Int) (*big.Int, error) {
	return _PDPService.Contract.NextChallengeWindowStart(&_PDPService.CallOpts, setId)
}

// NextChallengeWindowStart is a free data retrieval call binding the contract method 0x8bf96d28.
//
// Solidity: function nextChallengeWindowStart(uint256 setId) view returns(uint256)
func (_PDPService *PDPServiceCallerSession) NextChallengeWindowStart(setId *big.Int) (*big.Int, error) {
	return _PDPService.Contract.NextChallengeWindowStart(&_PDPService.CallOpts, setId)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PDPService *PDPServiceCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PDPService *PDPServiceSession) Owner() (common.Address, error) {
	return _PDPService.Contract.Owner(&_PDPService.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PDPService *PDPServiceCallerSession) Owner() (common.Address, error) {
	return _PDPService.Contract.Owner(&_PDPService.CallOpts)
}

// PdpVerifierAddress is a free data retrieval call binding the contract method 0xde4b6b71.
//
// Solidity: function pdpVerifierAddress() view returns(address)
func (_PDPService *PDPServiceCaller) PdpVerifierAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "pdpVerifierAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PdpVerifierAddress is a free data retrieval call binding the contract method 0xde4b6b71.
//
// Solidity: function pdpVerifierAddress() view returns(address)
func (_PDPService *PDPServiceSession) PdpVerifierAddress() (common.Address, error) {
	return _PDPService.Contract.PdpVerifierAddress(&_PDPService.CallOpts)
}

// PdpVerifierAddress is a free data retrieval call binding the contract method 0xde4b6b71.
//
// Solidity: function pdpVerifierAddress() view returns(address)
func (_PDPService *PDPServiceCallerSession) PdpVerifierAddress() (common.Address, error) {
	return _PDPService.Contract.PdpVerifierAddress(&_PDPService.CallOpts)
}

// ProvenThisPeriod is a free data retrieval call binding the contract method 0x7598a1cd.
//
// Solidity: function provenThisPeriod(uint256 ) view returns(bool)
func (_PDPService *PDPServiceCaller) ProvenThisPeriod(opts *bind.CallOpts, arg0 *big.Int) (bool, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "provenThisPeriod", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProvenThisPeriod is a free data retrieval call binding the contract method 0x7598a1cd.
//
// Solidity: function provenThisPeriod(uint256 ) view returns(bool)
func (_PDPService *PDPServiceSession) ProvenThisPeriod(arg0 *big.Int) (bool, error) {
	return _PDPService.Contract.ProvenThisPeriod(&_PDPService.CallOpts, arg0)
}

// ProvenThisPeriod is a free data retrieval call binding the contract method 0x7598a1cd.
//
// Solidity: function provenThisPeriod(uint256 ) view returns(bool)
func (_PDPService *PDPServiceCallerSession) ProvenThisPeriod(arg0 *big.Int) (bool, error) {
	return _PDPService.Contract.ProvenThisPeriod(&_PDPService.CallOpts, arg0)
}

// ProvingDeadlines is a free data retrieval call binding the contract method 0x11bc4865.
//
// Solidity: function provingDeadlines(uint256 ) view returns(uint256)
func (_PDPService *PDPServiceCaller) ProvingDeadlines(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "provingDeadlines", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProvingDeadlines is a free data retrieval call binding the contract method 0x11bc4865.
//
// Solidity: function provingDeadlines(uint256 ) view returns(uint256)
func (_PDPService *PDPServiceSession) ProvingDeadlines(arg0 *big.Int) (*big.Int, error) {
	return _PDPService.Contract.ProvingDeadlines(&_PDPService.CallOpts, arg0)
}

// ProvingDeadlines is a free data retrieval call binding the contract method 0x11bc4865.
//
// Solidity: function provingDeadlines(uint256 ) view returns(uint256)
func (_PDPService *PDPServiceCallerSession) ProvingDeadlines(arg0 *big.Int) (*big.Int, error) {
	return _PDPService.Contract.ProvingDeadlines(&_PDPService.CallOpts, arg0)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_PDPService *PDPServiceCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_PDPService *PDPServiceSession) ProxiableUUID() ([32]byte, error) {
	return _PDPService.Contract.ProxiableUUID(&_PDPService.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_PDPService *PDPServiceCallerSession) ProxiableUUID() ([32]byte, error) {
	return _PDPService.Contract.ProxiableUUID(&_PDPService.CallOpts)
}

// ThisChallengeWindowStart is a free data retrieval call binding the contract method 0x1506d198.
//
// Solidity: function thisChallengeWindowStart(uint256 setId) view returns(uint256)
func (_PDPService *PDPServiceCaller) ThisChallengeWindowStart(opts *bind.CallOpts, setId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _PDPService.contract.Call(opts, &out, "thisChallengeWindowStart", setId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ThisChallengeWindowStart is a free data retrieval call binding the contract method 0x1506d198.
//
// Solidity: function thisChallengeWindowStart(uint256 setId) view returns(uint256)
func (_PDPService *PDPServiceSession) ThisChallengeWindowStart(setId *big.Int) (*big.Int, error) {
	return _PDPService.Contract.ThisChallengeWindowStart(&_PDPService.CallOpts, setId)
}

// ThisChallengeWindowStart is a free data retrieval call binding the contract method 0x1506d198.
//
// Solidity: function thisChallengeWindowStart(uint256 setId) view returns(uint256)
func (_PDPService *PDPServiceCallerSession) ThisChallengeWindowStart(setId *big.Int) (*big.Int, error) {
	return _PDPService.Contract.ThisChallengeWindowStart(&_PDPService.CallOpts, setId)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _pdpVerifierAddress) returns()
func (_PDPService *PDPServiceTransactor) Initialize(opts *bind.TransactOpts, _pdpVerifierAddress common.Address) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "initialize", _pdpVerifierAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _pdpVerifierAddress) returns()
func (_PDPService *PDPServiceSession) Initialize(_pdpVerifierAddress common.Address) (*types.Transaction, error) {
	return _PDPService.Contract.Initialize(&_PDPService.TransactOpts, _pdpVerifierAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _pdpVerifierAddress) returns()
func (_PDPService *PDPServiceTransactorSession) Initialize(_pdpVerifierAddress common.Address) (*types.Transaction, error) {
	return _PDPService.Contract.Initialize(&_PDPService.TransactOpts, _pdpVerifierAddress)
}

// NextProvingPeriod is a paid mutator transaction binding the contract method 0xaa27ebcc.
//
// Solidity: function nextProvingPeriod(uint256 proofSetId, uint256 challengeEpoch, uint256 , bytes ) returns()
func (_PDPService *PDPServiceTransactor) NextProvingPeriod(opts *bind.TransactOpts, proofSetId *big.Int, challengeEpoch *big.Int, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "nextProvingPeriod", proofSetId, challengeEpoch, arg2, arg3)
}

// NextProvingPeriod is a paid mutator transaction binding the contract method 0xaa27ebcc.
//
// Solidity: function nextProvingPeriod(uint256 proofSetId, uint256 challengeEpoch, uint256 , bytes ) returns()
func (_PDPService *PDPServiceSession) NextProvingPeriod(proofSetId *big.Int, challengeEpoch *big.Int, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.NextProvingPeriod(&_PDPService.TransactOpts, proofSetId, challengeEpoch, arg2, arg3)
}

// NextProvingPeriod is a paid mutator transaction binding the contract method 0xaa27ebcc.
//
// Solidity: function nextProvingPeriod(uint256 proofSetId, uint256 challengeEpoch, uint256 , bytes ) returns()
func (_PDPService *PDPServiceTransactorSession) NextProvingPeriod(proofSetId *big.Int, challengeEpoch *big.Int, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.NextProvingPeriod(&_PDPService.TransactOpts, proofSetId, challengeEpoch, arg2, arg3)
}

// PossessionProven is a paid mutator transaction binding the contract method 0x356de02b.
//
// Solidity: function possessionProven(uint256 proofSetId, uint256 , uint256 , uint256 challengeCount) returns()
func (_PDPService *PDPServiceTransactor) PossessionProven(opts *bind.TransactOpts, proofSetId *big.Int, arg1 *big.Int, arg2 *big.Int, challengeCount *big.Int) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "possessionProven", proofSetId, arg1, arg2, challengeCount)
}

// PossessionProven is a paid mutator transaction binding the contract method 0x356de02b.
//
// Solidity: function possessionProven(uint256 proofSetId, uint256 , uint256 , uint256 challengeCount) returns()
func (_PDPService *PDPServiceSession) PossessionProven(proofSetId *big.Int, arg1 *big.Int, arg2 *big.Int, challengeCount *big.Int) (*types.Transaction, error) {
	return _PDPService.Contract.PossessionProven(&_PDPService.TransactOpts, proofSetId, arg1, arg2, challengeCount)
}

// PossessionProven is a paid mutator transaction binding the contract method 0x356de02b.
//
// Solidity: function possessionProven(uint256 proofSetId, uint256 , uint256 , uint256 challengeCount) returns()
func (_PDPService *PDPServiceTransactorSession) PossessionProven(proofSetId *big.Int, arg1 *big.Int, arg2 *big.Int, challengeCount *big.Int) (*types.Transaction, error) {
	return _PDPService.Contract.PossessionProven(&_PDPService.TransactOpts, proofSetId, arg1, arg2, challengeCount)
}

// ProofSetCreated is a paid mutator transaction binding the contract method 0x94d41b36.
//
// Solidity: function proofSetCreated(uint256 proofSetId, address creator, bytes ) returns()
func (_PDPService *PDPServiceTransactor) ProofSetCreated(opts *bind.TransactOpts, proofSetId *big.Int, creator common.Address, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "proofSetCreated", proofSetId, creator, arg2)
}

// ProofSetCreated is a paid mutator transaction binding the contract method 0x94d41b36.
//
// Solidity: function proofSetCreated(uint256 proofSetId, address creator, bytes ) returns()
func (_PDPService *PDPServiceSession) ProofSetCreated(proofSetId *big.Int, creator common.Address, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.ProofSetCreated(&_PDPService.TransactOpts, proofSetId, creator, arg2)
}

// ProofSetCreated is a paid mutator transaction binding the contract method 0x94d41b36.
//
// Solidity: function proofSetCreated(uint256 proofSetId, address creator, bytes ) returns()
func (_PDPService *PDPServiceTransactorSession) ProofSetCreated(proofSetId *big.Int, creator common.Address, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.ProofSetCreated(&_PDPService.TransactOpts, proofSetId, creator, arg2)
}

// ProofSetDeleted is a paid mutator transaction binding the contract method 0x26c249e3.
//
// Solidity: function proofSetDeleted(uint256 proofSetId, uint256 deletedLeafCount, bytes ) returns()
func (_PDPService *PDPServiceTransactor) ProofSetDeleted(opts *bind.TransactOpts, proofSetId *big.Int, deletedLeafCount *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "proofSetDeleted", proofSetId, deletedLeafCount, arg2)
}

// ProofSetDeleted is a paid mutator transaction binding the contract method 0x26c249e3.
//
// Solidity: function proofSetDeleted(uint256 proofSetId, uint256 deletedLeafCount, bytes ) returns()
func (_PDPService *PDPServiceSession) ProofSetDeleted(proofSetId *big.Int, deletedLeafCount *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.ProofSetDeleted(&_PDPService.TransactOpts, proofSetId, deletedLeafCount, arg2)
}

// ProofSetDeleted is a paid mutator transaction binding the contract method 0x26c249e3.
//
// Solidity: function proofSetDeleted(uint256 proofSetId, uint256 deletedLeafCount, bytes ) returns()
func (_PDPService *PDPServiceTransactorSession) ProofSetDeleted(proofSetId *big.Int, deletedLeafCount *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.ProofSetDeleted(&_PDPService.TransactOpts, proofSetId, deletedLeafCount, arg2)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PDPService *PDPServiceTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PDPService *PDPServiceSession) RenounceOwnership() (*types.Transaction, error) {
	return _PDPService.Contract.RenounceOwnership(&_PDPService.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PDPService *PDPServiceTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _PDPService.Contract.RenounceOwnership(&_PDPService.TransactOpts)
}

// RootsAdded is a paid mutator transaction binding the contract method 0x12d5d66f.
//
// Solidity: function rootsAdded(uint256 proofSetId, uint256 firstAdded, ((bytes),uint256)[] rootData, bytes ) returns()
func (_PDPService *PDPServiceTransactor) RootsAdded(opts *bind.TransactOpts, proofSetId *big.Int, firstAdded *big.Int, rootData []PDPVerifierRootData, arg3 []byte) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "rootsAdded", proofSetId, firstAdded, rootData, arg3)
}

// RootsAdded is a paid mutator transaction binding the contract method 0x12d5d66f.
//
// Solidity: function rootsAdded(uint256 proofSetId, uint256 firstAdded, ((bytes),uint256)[] rootData, bytes ) returns()
func (_PDPService *PDPServiceSession) RootsAdded(proofSetId *big.Int, firstAdded *big.Int, rootData []PDPVerifierRootData, arg3 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.RootsAdded(&_PDPService.TransactOpts, proofSetId, firstAdded, rootData, arg3)
}

// RootsAdded is a paid mutator transaction binding the contract method 0x12d5d66f.
//
// Solidity: function rootsAdded(uint256 proofSetId, uint256 firstAdded, ((bytes),uint256)[] rootData, bytes ) returns()
func (_PDPService *PDPServiceTransactorSession) RootsAdded(proofSetId *big.Int, firstAdded *big.Int, rootData []PDPVerifierRootData, arg3 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.RootsAdded(&_PDPService.TransactOpts, proofSetId, firstAdded, rootData, arg3)
}

// RootsScheduledRemove is a paid mutator transaction binding the contract method 0x4af7d1d2.
//
// Solidity: function rootsScheduledRemove(uint256 proofSetId, uint256[] rootIds, bytes ) returns()
func (_PDPService *PDPServiceTransactor) RootsScheduledRemove(opts *bind.TransactOpts, proofSetId *big.Int, rootIds []*big.Int, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "rootsScheduledRemove", proofSetId, rootIds, arg2)
}

// RootsScheduledRemove is a paid mutator transaction binding the contract method 0x4af7d1d2.
//
// Solidity: function rootsScheduledRemove(uint256 proofSetId, uint256[] rootIds, bytes ) returns()
func (_PDPService *PDPServiceSession) RootsScheduledRemove(proofSetId *big.Int, rootIds []*big.Int, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.RootsScheduledRemove(&_PDPService.TransactOpts, proofSetId, rootIds, arg2)
}

// RootsScheduledRemove is a paid mutator transaction binding the contract method 0x4af7d1d2.
//
// Solidity: function rootsScheduledRemove(uint256 proofSetId, uint256[] rootIds, bytes ) returns()
func (_PDPService *PDPServiceTransactorSession) RootsScheduledRemove(proofSetId *big.Int, rootIds []*big.Int, arg2 []byte) (*types.Transaction, error) {
	return _PDPService.Contract.RootsScheduledRemove(&_PDPService.TransactOpts, proofSetId, rootIds, arg2)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PDPService *PDPServiceTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PDPService *PDPServiceSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PDPService.Contract.TransferOwnership(&_PDPService.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PDPService *PDPServiceTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PDPService.Contract.TransferOwnership(&_PDPService.TransactOpts, newOwner)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_PDPService *PDPServiceTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _PDPService.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_PDPService *PDPServiceSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _PDPService.Contract.UpgradeToAndCall(&_PDPService.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_PDPService *PDPServiceTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _PDPService.Contract.UpgradeToAndCall(&_PDPService.TransactOpts, newImplementation, data)
}

// PDPServiceFaultRecordIterator is returned from FilterFaultRecord and is used to iterate over the raw logs and unpacked data for FaultRecord events raised by the PDPService contract.
type PDPServiceFaultRecordIterator struct {
	Event *PDPServiceFaultRecord // Event containing the contract specifics and raw log

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
func (it *PDPServiceFaultRecordIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PDPServiceFaultRecord)
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
		it.Event = new(PDPServiceFaultRecord)
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
func (it *PDPServiceFaultRecordIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PDPServiceFaultRecordIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PDPServiceFaultRecord represents a FaultRecord event raised by the PDPService contract.
type PDPServiceFaultRecord struct {
	ProofSetId     *big.Int
	PeriodsFaulted *big.Int
	Deadline       *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterFaultRecord is a free log retrieval operation binding the contract event 0xff5f076c63706be9f7eaafa8329db4a9ce9b9e3cd6e53470f05491e2043e1a81.
//
// Solidity: event FaultRecord(uint256 indexed proofSetId, uint256 periodsFaulted, uint256 deadline)
func (_PDPService *PDPServiceFilterer) FilterFaultRecord(opts *bind.FilterOpts, proofSetId []*big.Int) (*PDPServiceFaultRecordIterator, error) {

	var proofSetIdRule []interface{}
	for _, proofSetIdItem := range proofSetId {
		proofSetIdRule = append(proofSetIdRule, proofSetIdItem)
	}

	logs, sub, err := _PDPService.contract.FilterLogs(opts, "FaultRecord", proofSetIdRule)
	if err != nil {
		return nil, err
	}
	return &PDPServiceFaultRecordIterator{contract: _PDPService.contract, event: "FaultRecord", logs: logs, sub: sub}, nil
}

// WatchFaultRecord is a free log subscription operation binding the contract event 0xff5f076c63706be9f7eaafa8329db4a9ce9b9e3cd6e53470f05491e2043e1a81.
//
// Solidity: event FaultRecord(uint256 indexed proofSetId, uint256 periodsFaulted, uint256 deadline)
func (_PDPService *PDPServiceFilterer) WatchFaultRecord(opts *bind.WatchOpts, sink chan<- *PDPServiceFaultRecord, proofSetId []*big.Int) (event.Subscription, error) {

	var proofSetIdRule []interface{}
	for _, proofSetIdItem := range proofSetId {
		proofSetIdRule = append(proofSetIdRule, proofSetIdItem)
	}

	logs, sub, err := _PDPService.contract.WatchLogs(opts, "FaultRecord", proofSetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PDPServiceFaultRecord)
				if err := _PDPService.contract.UnpackLog(event, "FaultRecord", log); err != nil {
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

// ParseFaultRecord is a log parse operation binding the contract event 0xff5f076c63706be9f7eaafa8329db4a9ce9b9e3cd6e53470f05491e2043e1a81.
//
// Solidity: event FaultRecord(uint256 indexed proofSetId, uint256 periodsFaulted, uint256 deadline)
func (_PDPService *PDPServiceFilterer) ParseFaultRecord(log types.Log) (*PDPServiceFaultRecord, error) {
	event := new(PDPServiceFaultRecord)
	if err := _PDPService.contract.UnpackLog(event, "FaultRecord", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PDPServiceInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the PDPService contract.
type PDPServiceInitializedIterator struct {
	Event *PDPServiceInitialized // Event containing the contract specifics and raw log

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
func (it *PDPServiceInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PDPServiceInitialized)
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
		it.Event = new(PDPServiceInitialized)
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
func (it *PDPServiceInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PDPServiceInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PDPServiceInitialized represents a Initialized event raised by the PDPService contract.
type PDPServiceInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PDPService *PDPServiceFilterer) FilterInitialized(opts *bind.FilterOpts) (*PDPServiceInitializedIterator, error) {

	logs, sub, err := _PDPService.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &PDPServiceInitializedIterator{contract: _PDPService.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PDPService *PDPServiceFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *PDPServiceInitialized) (event.Subscription, error) {

	logs, sub, err := _PDPService.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PDPServiceInitialized)
				if err := _PDPService.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PDPService *PDPServiceFilterer) ParseInitialized(log types.Log) (*PDPServiceInitialized, error) {
	event := new(PDPServiceInitialized)
	if err := _PDPService.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PDPServiceOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the PDPService contract.
type PDPServiceOwnershipTransferredIterator struct {
	Event *PDPServiceOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *PDPServiceOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PDPServiceOwnershipTransferred)
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
		it.Event = new(PDPServiceOwnershipTransferred)
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
func (it *PDPServiceOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PDPServiceOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PDPServiceOwnershipTransferred represents a OwnershipTransferred event raised by the PDPService contract.
type PDPServiceOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PDPService *PDPServiceFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PDPServiceOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PDPService.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PDPServiceOwnershipTransferredIterator{contract: _PDPService.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PDPService *PDPServiceFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PDPServiceOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PDPService.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PDPServiceOwnershipTransferred)
				if err := _PDPService.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PDPService *PDPServiceFilterer) ParseOwnershipTransferred(log types.Log) (*PDPServiceOwnershipTransferred, error) {
	event := new(PDPServiceOwnershipTransferred)
	if err := _PDPService.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PDPServiceUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the PDPService contract.
type PDPServiceUpgradedIterator struct {
	Event *PDPServiceUpgraded // Event containing the contract specifics and raw log

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
func (it *PDPServiceUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PDPServiceUpgraded)
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
		it.Event = new(PDPServiceUpgraded)
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
func (it *PDPServiceUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PDPServiceUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PDPServiceUpgraded represents a Upgraded event raised by the PDPService contract.
type PDPServiceUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_PDPService *PDPServiceFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*PDPServiceUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _PDPService.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &PDPServiceUpgradedIterator{contract: _PDPService.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_PDPService *PDPServiceFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *PDPServiceUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _PDPService.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PDPServiceUpgraded)
				if err := _PDPService.contract.UnpackLog(event, "Upgraded", log); err != nil {
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

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_PDPService *PDPServiceFilterer) ParseUpgraded(log types.Log) (*PDPServiceUpgraded, error) {
	event := new(PDPServiceUpgraded)
	if err := _PDPService.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
