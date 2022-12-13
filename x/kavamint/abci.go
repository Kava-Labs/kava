package kavamint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/kavamint/keeper"
)

// BeginBlocker mints & distributes new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.KeeperI) {
	if err := k.AccumulateAndMintInflation(ctx); err != nil {
		panic(err)
	}
}
