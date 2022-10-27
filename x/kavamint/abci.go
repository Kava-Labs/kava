package kavamint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavamint/keeper"
)

// BeginBlocker mints & distributes new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	// TODO: mint tokens for staking rewards
	// TODO: send staking tokens to auth fee collector
	// TODO: mint tokens for community pool
	// TODO: send tokens to community pool

	// TODO: emit event
}
