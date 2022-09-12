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
	cdc codec.Codec

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper

	derivativeDenom string
}

// NewKeeper returns a new keeper for the liquid module.
func NewKeeper(
	cdc codec.Codec,
	ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper,
	derivativeDenom string,
) Keeper {

	return Keeper{
		cdc:             cdc,
		accountKeeper:   ak,
		bankKeeper:      bk,
		stakingKeeper:   sk,
		derivativeDenom: derivativeDenom,
	}
}

// NewDefaultKeeper returns a new keeper for the liquid module with default values.
func NewDefaultKeeper(
	cdc codec.Codec,
	ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper,
) Keeper {

	return NewKeeper(cdc, ak, bk, sk, types.DefaultDerivativeDenom)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
