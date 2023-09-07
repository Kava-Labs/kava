package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/types"
)

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, bool) {
	store := ctx.KVStore(k.key)

	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.Params{}, false
	}

	params := types.Params{}
	k.cdc.MustUnmarshal(bz, &params)

	return params, true
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.key)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)
	return nil
}
