package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// MintPeriodInflation mints new tokens according to the inflation schedule specified in the parameters
func (k Keeper) MintPeriodInflation(ctx sdk.Context) error {
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
	for _, period := range params.Periods {
		// Case 1 - period is fully expired
		if period.End.Before(previousBlockTime) {
			continue
		}
		// Case 2 - period has ended since the previous block time
		if period.End.After(previousBlockTime) && period.End.Before(ctx.BlockTime()) {
			// calculate time elapsed relative to the periods end time
			timeElapsed := sdk.NewInt(period.End.Unix() - previousBlockTime.Unix())
			err := k.mintInflationaryCoins(ctx, period.Inflation, timeElapsed, types.GovDenom)
			if err != nil {
				return err
			}
			// update the value of previousBlockTime so that the next period starts from the end of the last
			// period and not the original value of previousBlockTime
			previousBlockTime = period.End
		}
		// Case 3 - period is ongoing
		if (period.Start.Before(previousBlockTime) || period.Start.Equal(previousBlockTime)) && period.End.After(ctx.BlockTime()) {
			// calculate time elapsed relative to the current block time
			timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())
			err := k.mintInflationaryCoins(ctx, period.Inflation, timeElapsed, types.GovDenom)
			if err != nil {
				return err
			}
		}
		// Case 4 - period hasn't started
		if period.Start.After(ctx.BlockTime()) || period.Start.Equal(ctx.BlockTime()) {
			continue
		}
	}
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
	return nil
}

func (k Keeper) mintInflationaryCoins(ctx sdk.Context, inflationRate sdk.Dec, timePeriods sdk.Int, denom string) error {
	totalSupply := k.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(denom)
	// used to scale accumulator calculations by 10^18
	scalar := sdk.NewInt(1000000000000000000)
	// convert inflation rate to integer
	inflationInt := inflationRate.Mul(sdk.NewDecFromInt(scalar)).TruncateInt()
	// calculate the multiplier (amount to multiply the total supply by to achieve the desired inflation)
	// multiply the result by 10^-18 because RelativePow returns the result scaled by 10^18
	accumulator := sdk.NewDecFromInt(cdptypes.RelativePow(inflationInt, timePeriods, scalar)).Mul(sdk.SmallestDec())
	// calculate the number of coins to mint
	amountToMint := (sdk.NewDecFromInt(totalSupply).Mul(accumulator)).Sub(sdk.NewDecFromInt(totalSupply)).TruncateInt()
	if amountToMint.IsZero() {
		return nil
	}
	err := k.supplyKeeper.MintCoins(ctx, types.KavaDistMacc, sdk.NewCoins(sdk.NewCoin(denom, amountToMint)))
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKavaDist,
			sdk.NewAttribute(types.AttributeKeyInflation, sdk.NewCoin(denom, amountToMint).String()),
		),
	)
	return nil
}
