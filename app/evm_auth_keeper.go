package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

type EVMAuthKeeper struct {
	authkeeper.AccountKeeper
	proto func() authtypes.AccountI
}

var _ evmtypes.AccountKeeper = (*EVMAuthKeeper)(nil)

func NewEVMAuthKeeper(ak authkeeper.AccountKeeper, accountProto func() authtypes.AccountI) EVMAuthKeeper {
	return EVMAuthKeeper{
		AccountKeeper: ak,
		proto:         accountProto,
	}
}

func (ak EVMAuthKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		panic(err)
	}

	return ak.NewAccount(ctx, acc)
}
