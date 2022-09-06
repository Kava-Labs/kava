package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/router/types"
)

// Keeper is the keeper for the router module
type Keeper struct {
	cdc codec.Codec

	earnKeeper    types.EarnKeeper
	liquidKeeper  types.LiquidKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc codec.Codec,
	earnKeeper types.EarnKeeper,
	liquidKeeper types.LiquidKeeper,
	stakingKeeper types.StakingKeeper,
) Keeper {

	return Keeper{
		cdc:           cdc,
		earnKeeper:    earnKeeper,
		liquidKeeper:  liquidKeeper,
		stakingKeeper: stakingKeeper,
	}
}
