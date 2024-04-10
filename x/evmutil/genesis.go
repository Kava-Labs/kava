package evmutil

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/x/evm/statedb"

	"github.com/kava-labs/kava/precompile/contracts/mul3"
	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, gs *types.GenesisState, ak types.AccountKeeper, evmKeeper statedb.Keeper) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	txConfig := statedb.TxConfig{
		BlockHash: common.Hash{},
		TxHash:    common.Hash{},
		TxIndex:   0,
		LogIndex:  0,
	}
	stateDB := statedb.New(ctx, evmKeeper, txConfig)

	// Set the nonce of the precompile's address (as is done when a contract is created) to ensure
	// that it is marked as non-empty and will not be cleaned up when the statedb is finalized.
	stateDB.SetNonce(mul3.ContractAddress, 1)
	// Set the code of the precompile's address to a non-zero length byte slice to ensure that the precompile
	// can be called from within Solidity contracts. Solidity adds a check before invoking a contract to ensure
	// that it does not attempt to invoke a non-existent contract.
	stateDB.SetCode(mul3.ContractAddress, []byte{0x1})

	err := stateDB.Commit()
	if err != nil {
		panic(err)
	}

	keeper.SetParams(ctx, gs.Params)

	// initialize module account
	if moduleAcc := ak.GetModuleAccount(ctx, types.ModuleName); moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	for _, account := range gs.Accounts {
		keeper.SetAccount(ctx, account)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	accounts := keeper.GetAllAccounts(ctx)
	return types.NewGenesisState(accounts, keeper.GetParams(ctx))
}
