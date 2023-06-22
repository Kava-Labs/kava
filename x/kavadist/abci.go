package kavadist

import (
	"github.com/kava-labs/kava/x/kavadist/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	err := k.MintPeriodInflation(ctx)
	if err != nil {
		panic(err)
	}
}
