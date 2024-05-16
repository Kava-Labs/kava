package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank/types"
)

// Enforce that Keeper implements the expected keeper interfaces
var _ types.BankKeeper = Keeper{}

// Keeper defines the precisebank module's keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

func (k Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	panic("unimplemented")
}

func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	panic("unimplemented")
}
