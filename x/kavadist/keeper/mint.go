package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	err := k.mintIncentivePeriods(ctx, params.Periods, previousBlockTime)
	if err != nil {
		return err
	}

	coinsToDistribute, timeElapsed, err := k.mintInfrastructurePeriods(ctx, params.InfrastructureParams.InfrastructurePeriods, previousBlockTime)
	if err != nil {
		return err
	}

	err = k.distributeInfrastructureCoins(ctx, params.InfrastructureParams.PartnerRewards, params.InfrastructureParams.CoreRewards, timeElapsed, coinsToDistribute)
	if err != nil {
		return err
	}
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
	return nil
}

func (k Keeper) mintIncentivePeriods(ctx sdk.Context, periods types.Periods, previousBlockTime time.Time) error {
	var err error
	for _, period := range periods {
		switch {
		// Case 1 - period is fully expired
		case period.End.Before(previousBlockTime):
			continue

		// Case 2 - period has ended since the previous block time
		case period.End.After(previousBlockTime) && (period.End.Before(ctx.BlockTime()) || period.End.Equal(ctx.BlockTime())):
			// calculate time elapsed relative to the periods end time
			timeElapsed := sdk.NewInt(period.End.Unix() - previousBlockTime.Unix())
			_, err = k.mintInflationaryCoins(ctx, period.Inflation, timeElapsed, types.GovDenom)
			// update the value of previousBlockTime so that the next period starts from the end of the last
			// period and not the original value of previousBlockTime
			previousBlockTime = period.End

		// Case 3 - period is ongoing
		case (period.Start.Before(previousBlockTime) || period.Start.Equal(previousBlockTime)) && period.End.After(ctx.BlockTime()):
			// calculate time elapsed relative to the current block time
			timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())
			_, err = k.mintInflationaryCoins(ctx, period.Inflation, timeElapsed, types.GovDenom)

		// Case 4 - period hasn't started
		case period.Start.After(ctx.BlockTime()) || period.Start.Equal(ctx.BlockTime()):
			continue
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) mintInflationaryCoins(ctx sdk.Context, inflationRate sdk.Dec, timePeriods sdk.Int, denom string) (sdk.Coin, error) {
	totalSupply := k.bankKeeper.GetSupply(ctx, denom)
	// used to scale accumulator calculations by 10^18
	scalar := sdk.NewInt(1000000000000000000)
	// convert inflation rate to integer
	inflationInt := sdk.NewUintFromBigInt(inflationRate.Mul(sdk.NewDecFromInt(scalar)).TruncateInt().BigInt())
	timePeriodsUint := sdk.NewUintFromBigInt(timePeriods.BigInt())
	scalarUint := sdk.NewUintFromBigInt(scalar.BigInt())
	// calculate the multiplier (amount to multiply the total supply by to achieve the desired inflation)
	// multiply the result by 10^-18 because RelativePow returns the result scaled by 10^18
	accumulator := sdk.NewDecFromBigInt(sdk.RelativePow(inflationInt, timePeriodsUint, scalarUint).BigInt()).Mul(sdk.SmallestDec())
	// calculate the number of coins to mint
	amountToMint := (sdk.NewDecFromInt(totalSupply.Amount).Mul(accumulator)).Sub(sdk.NewDecFromInt(totalSupply.Amount)).TruncateInt()
	if amountToMint.IsZero() {
		return sdk.Coin{}, nil
	}
	err := k.bankKeeper.MintCoins(ctx, types.KavaDistMacc, sdk.NewCoins(sdk.NewCoin(denom, amountToMint)))
	if err != nil {
		return sdk.Coin{}, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKavaDist,
			sdk.NewAttribute(types.AttributeKeyInflation, sdk.NewCoin(denom, amountToMint).String()),
		),
	)

	return sdk.NewCoin(denom, amountToMint), nil
}
