package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/kavadist/types"
)

func (k Keeper) MintPeriodRewards(ctx sdk.Context) {
	params := k.GetParams(ctx)
	previousBlockTime := k.GetPreviousBlockTime(ctx)
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())

	for _, p := range params.Periods {
		if p.Start.After(ctx.BlockTime()) && p.End.Before(ctx.BlockTime()) {
			totalSupply := k.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)

		}
	}
}
