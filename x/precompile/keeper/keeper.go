package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Keeper of the precompile store.
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
}

// NewKeeper creates an precompile keeper.
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}
