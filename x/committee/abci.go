package committee

import (
	"github.com/kava-labs/kava/x/committee/keeper"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker runs at the start of every block.
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, k keeper.Keeper) {
	k.ProcessProposals(ctx)
}
