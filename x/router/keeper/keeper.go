package keeper

import (
	"github.com/kava-labs/kava/x/router/types"
)

// Keeper is the keeper for the module
type Keeper struct {
	earnKeeper    types.EarnKeeper
	liquidKeeper  types.LiquidKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	earnKeeper types.EarnKeeper,
	liquidKeeper types.LiquidKeeper,
	stakingKeeper types.StakingKeeper,
) Keeper {

	return Keeper{
		earnKeeper:    earnKeeper,
		liquidKeeper:  liquidKeeper,
		stakingKeeper: stakingKeeper,
	}
}
