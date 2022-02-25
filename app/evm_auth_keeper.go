package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

type EVMAccountKeeper struct {
	authkeeper.AccountKeeper
	proto func() authtypes.AccountI
}

var _ evmtypes.AccountKeeper = (*EVMAccountKeeper)(nil)

func NewEVMAccountKeeper(ak authkeeper.AccountKeeper, accountProto func() authtypes.AccountI) EVMAccountKeeper {
	return EVMAccountKeeper{
		AccountKeeper: ak,
		proto:         accountProto,
	}
}

func (ak EVMAccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		panic(err)
	}

	return ak.NewAccount(ctx, acc)
}
