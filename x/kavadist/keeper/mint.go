package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/kavadist/types"
)

func (k Keeper) MintPeriodRewards(ctx sdk.Context) sdk.Error {
	params := k.GetParams(ctx)
	if !params.Active {
		// TODO emit event
		return nil
	}
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
		k.SetPreviousBlockTime(ctx, previousBlockTime)
		return nil
	}
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())

	for _, p := range params.Periods {
		if p.Start.Before(ctx.BlockTime()) && p.End.After(ctx.BlockTime()) { // TODO do we need to handle equal?
			totalSupply := k.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)
			scalar := sdk.NewInt(1000000000000000000)
			inflationInt := p.Inflation.Mul(sdk.NewDecFromInt(scalar)).TruncateInt()
			accumulator := sdk.NewDecFromInt(cdptypes.RelativePow(inflationInt, timeElapsed, scalar)).Mul(sdk.SmallestDec())
			amountToMint := (sdk.NewDecFromInt(totalSupply).Mul(accumulator)).Sub(sdk.NewDecFromInt(totalSupply)).TruncateInt()
			err := k.supplyKeeper.MintCoins(ctx, types.KavaDistMacc, sdk.NewCoins(sdk.NewCoin(types.GovDenom, amountToMint)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// TEST considerations
// Sanity checks - doesn't mint for expired (past) or upcoming (future periods) - check supply doesn't change
// Does mint for non-expired periods - first check supply does change, next check the amount supply changes
// Mint one year's worth of coins - is the apr ~= the spr?
// Before merging - should be run on a local testnet, to check that the begin blocker's order is okay
