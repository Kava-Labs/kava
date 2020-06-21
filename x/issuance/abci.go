package issuance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/issuance/keeper"
)

// BeginBlocker iterates over each asset and seizes coins from blocked addresses by returning them to the asset owner
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)
	for _, asset := range params.Assets {
		err := k.SeizeCoinsFromBlockedAddresses(ctx, asset.Denom)
		if err != nil {
			panic(err)
		}
	}
}
