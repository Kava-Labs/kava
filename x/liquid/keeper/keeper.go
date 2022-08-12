package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/liquid/types"
)

// Keeper struct for the liquid module.
type Keeper struct {
	key sdk.StoreKey
	cdc codec.Codec

	paramSubspace types.ParamSubspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper returns a new keeper for the liquid module.
func NewKeeper(
	cdc codec.Codec, key sdk.StoreKey, paramstore types.ParamSubspace,
	ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		key:           key,
		paramSubspace: paramstore,
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
