package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// ---------- Module Parameters ----------
// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.CdpParams {
	var p types.CdpParams
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, cdpParams types.CdpParams) {
	k.paramSubspace.SetParamSet(ctx, &cdpParams)
}