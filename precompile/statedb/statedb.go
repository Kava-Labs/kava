package statedb

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	ethermint_statedb "github.com/evmos/ethermint/x/evm/statedb"
	evm "github.com/evmos/ethermint/x/evm/vm"
	"github.com/evmos/ethermint/x/evm/vm/geth"

	"github.com/kava-labs/kava/x/evmutil/types"
)

type IBCTransferKeeper interface {
	Transfer(goCtx context.Context, msg *ibctransfertypes.MsgTransfer) (*ibctransfertypes.MsgTransferResponse, error)
}

type EvmutilKeeper interface {
	GetEnabledConversionPairFromERC20Address(
		ctx sdk.Context,
		address types.InternalEVMAddress,
	) (types.ConversionPair, error)

	ConvertERC20ToCoin(
		goCtx context.Context,
		msg *types.MsgConvertERC20ToCoin,
	) (*types.MsgConvertERC20ToCoinResponse, error)
}

type StateDBWithContext interface {
	vm.StateDB
	Context() context.Context
}

type StateDBWithKeepers interface {
	StateDBWithContext
	IBCTransferKeeper() IBCTransferKeeper
	EvmutilKeeper() EvmutilKeeper
}

type stateDB struct {
	StateDBWithContext
	ibcTransferKeeper IBCTransferKeeper
	evmutilKeeper     EvmutilKeeper
}

func NewStateDB(stateDBWithContext StateDBWithContext, ibcTransferKeeper IBCTransferKeeper, evmutilKeeper EvmutilKeeper) StateDBWithKeepers {
	return &stateDB{
		StateDBWithContext: stateDBWithContext,
		ibcTransferKeeper:  ibcTransferKeeper,
		evmutilKeeper:      evmutilKeeper,
	}
}

func GetEVMConstructor(
	ibcTransferKeeper IBCTransferKeeper,
	evmutilKeeper EvmutilKeeper,
) evm.Constructor {
	return func(
		blockCtx vm.BlockContext,
		txCtx vm.TxContext,
		stateDB vm.StateDB,
		chainConfig *params.ChainConfig,
		config vm.Config,
		customPrecompiles evm.PrecompiledContracts,
	) evm.EVM {
		stateDBWithContext, ok := stateDB.(*ethermint_statedb.StateDB)
		if !ok {
			panic("unsupported statedb")
		}
		customStateDB := NewStateDB(stateDBWithContext, ibcTransferKeeper, evmutilKeeper)
		return geth.NewEVM(blockCtx, txCtx, customStateDB, chainConfig, config, customPrecompiles)
	}
}

func (s *stateDB) Context() context.Context {
	return s.StateDBWithContext.Context()
}

func (s *stateDB) IBCTransferKeeper() IBCTransferKeeper {
	return s.ibcTransferKeeper
}

func (s *stateDB) EvmutilKeeper() EvmutilKeeper {
	return s.evmutilKeeper
}
