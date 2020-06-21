package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/issuance/types"
)

// Keeper keeper for the issuance module
type Keeper struct {
	key           sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	accountKeeper types.AccountKeeper
	supplyKeeper  types.SupplyKeeper
}

// NewKeeper returns a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, sk types.SupplyKeeper) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore,
		supplyKeeper:  sk,
	}
}
