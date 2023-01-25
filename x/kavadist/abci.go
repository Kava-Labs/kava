package kavadist

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/keeper"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	err := k.DistributeFunds(ctx)
	if err != nil {
		panic(err)
	}
}
