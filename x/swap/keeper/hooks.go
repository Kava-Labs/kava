package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetPoolShares(ctx sdk.Context, poolID string) (sdk.Dec, bool) {
	// FIXME return pool shares once merged with acceptance branch
	return sdk.Dec{}, false
}
