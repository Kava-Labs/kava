package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// MintPeriodRewards mints new tokens according to the inflation schedule specified in the paramters
func (k Keeper) MintPeriodRewards(ctx sdk.Context) sdk.Error {
	params := k.GetParams(ctx)
	if !params.Active {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeKavaDist,
				sdk.NewAttribute(types.AttributeKeyStatus, types.AttributeValueInactive),
			),
		)
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
		if p.Start.Before(ctx.BlockTime()) && (p.End.After(ctx.BlockTime()) || p.End.Equal(ctx.BlockTime())) {
			totalSupply := k.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.GovDenom)
			scalar := sdk.NewInt(1000000000000000000)
			inflationInt := p.Inflation.Mul(sdk.NewDecFromInt(scalar)).TruncateInt()
			accumulator := sdk.NewDecFromInt(cdptypes.RelativePow(inflationInt, timeElapsed, scalar)).Mul(sdk.SmallestDec())
			amountToMint := (sdk.NewDecFromInt(totalSupply).Mul(accumulator)).Sub(sdk.NewDecFromInt(totalSupply)).TruncateInt()
			err := k.supplyKeeper.MintCoins(ctx, types.KavaDistMacc, sdk.NewCoins(sdk.NewCoin(types.GovDenom, amountToMint)))
			if err != nil {
				return err
			}
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeKavaDist,
					sdk.NewAttribute(types.AttributeKeyInflation, amountToMint.String()),
				),
			)
		}
	}
	return nil
}
