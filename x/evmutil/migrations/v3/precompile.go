package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	"github.com/evmos/ethermint/x/evm/statedb"
	"math/big"
)

var ContractAddress = common.HexToAddress("0x0300000000000000000000000000000000000002")

// Keeper provide underlying storage of StateDB
type Keeper interface {
	// Read methods
	GetAccount(ctx sdk.Context, addr common.Address) *statedb.Account
	GetState(ctx sdk.Context, addr common.Address, key common.Hash) common.Hash
	GetCode(ctx sdk.Context, codeHash common.Hash) []byte
	// the callback returns false to break early
	ForEachStorage(ctx sdk.Context, addr common.Address, cb func(key, value common.Hash) bool)

	// Write methods, only called by `StateDB.Commit()`
	SetAccount(ctx sdk.Context, addr common.Address, account statedb.Account) error
	SetState(ctx sdk.Context, addr common.Address, key common.Hash, value []byte)
	SetCode(ctx sdk.Context, codeHash []byte, code []byte)
	SetBalance(ctx sdk.Context, addr common.Address, amount *big.Int) error
	DeleteAccount(ctx sdk.Context, addr common.Address) error
}

func Migrate(
	ctx sdk.Context,
	evmKeeper *evmkeeper.Keeper,
) error {
	txConfig := statedb.TxConfig{
		BlockHash: common.Hash{},
		TxHash:    common.Hash{},
		TxIndex:   0,
		LogIndex:  0,
	}
	stateDB := statedb.New(ctx, evmKeeper, txConfig)

	// Set the nonce of the precompile's address (as is done when a contract is created) to ensure
	// that it is marked as non-empty and will not be cleaned up when the statedb is finalized.
	stateDB.SetNonce(ContractAddress, 1)
	// Set the code of the precompile's address to a non-zero length byte slice to ensure that the precompile
	// can be called from within Solidity contracts. Solidity adds a check before invoking a contract to ensure
	// that it does not attempt to invoke a non-existent contract.
	stateDB.SetCode(ContractAddress, []byte{0x1})

	return nil
}
