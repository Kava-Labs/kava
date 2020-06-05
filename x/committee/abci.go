package committee

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker runs at the start of every block.
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, k Keeper) {
	// enact proposals ignoring their expiry time - they could have received enough votes last block before expiring this block
	k.EnactPassedProposals(ctx)
	k.CloseExpiredProposals(ctx)
}
