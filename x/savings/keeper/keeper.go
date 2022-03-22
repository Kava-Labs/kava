package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/savings/types"
)

// Keeper struct for savings module
type Keeper struct {
	// key used to access the stores from Context
	key sdk.StoreKey
	// Codec for binary encoding/decoding
	cdc codec.Codec
	// The reference to the Paramstore to get and set savings specific params
	paramSubspace paramtypes.Subspace
}

// NewKeeper returns a new keeper for the savings module.
func NewKeeper(
	cdc codec.Codec, key sdk.StoreKey, paramstore paramtypes.Subspace,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		key:           key,
		paramSubspace: paramstore,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
