package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

func (k Keeper) mintInfrastructurePeriods(ctx sdk.Context, periods types.Periods, previousBlockTime time.Time) (sdk.Coin, sdk.Int, error) {
	var err error
	coinsMinted := sdk.NewCoin(types.GovDenom, sdk.ZeroInt())
	timeElapsed := sdk.ZeroInt()
	for _, period := range periods {
		switch {
		// Case 1 - period is fully expired
		case period.End.Before(previousBlockTime):
			continue

		// Case 2 - period has ended since the previous block time
		case period.End.After(previousBlockTime) && (period.End.Before(ctx.BlockTime()) || period.End.Equal(ctx.BlockTime())):
			// calculate time elapsed relative to the periods end time
			timeElapsed = sdk.NewInt(period.End.Unix() - previousBlockTime.Unix())
			coins, errI := k.mintInflationaryCoins(ctx, period.Inflation, timeElapsed, types.GovDenom)
			err = errI
			if !coins.IsZero() {
				coinsMinted = coinsMinted.Add(coins)
			}
			// update the value of previousBlockTime so that the next period starts from the end of the last
			// period and not the original value of previousBlockTime
			previousBlockTime = period.End

		// Case 3 - period is ongoing
		case (period.Start.Before(previousBlockTime) || period.Start.Equal(previousBlockTime)) && period.End.After(ctx.BlockTime()):
			// calculate time elapsed relative to the current block time
			timeElapsed = sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())
			coins, errI := k.mintInflationaryCoins(ctx, period.Inflation, timeElapsed, types.GovDenom)
			if !coins.IsZero() {
				coinsMinted = coinsMinted.Add(coins)
			}
			err = errI

		// Case 4 - period hasn't started
		case period.Start.After(ctx.BlockTime()) || period.Start.Equal(ctx.BlockTime()):
			timeElapsed = sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())
			continue
		}

		if err != nil {
			return sdk.Coin{}, sdk.Int{}, err
		}
	}
	return coinsMinted, timeElapsed, nil
}

func (k Keeper) distributeInfrastructureCoins(ctx sdk.Context, partnerRewards types.PartnerRewards, coreRewards types.CoreRewards, timeElapsed sdk.Int, coinsToDistribute sdk.Coin) error {
	if timeElapsed.IsZero() {
		return nil
	}
	if coinsToDistribute.IsZero() {
		return nil
	}
	for _, pr := range partnerRewards {
		coinsToSend := sdk.NewCoin(types.GovDenom, pr.RewardsPerSecond.Amount.Mul(timeElapsed))
		// TODO check balance, log if insufficient and return rather than error
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, pr.Address, sdk.NewCoins(coinsToSend))
		if err != nil {
			return err
		}
		neg, updatedCoins := safeSub(coinsToDistribute, coinsToSend)
		if neg {
			return fmt.Errorf("negative coins")
		}
		coinsToDistribute = updatedCoins
	}
	for _, cr := range coreRewards {
		coinsToSend := sdk.NewCoin(types.GovDenom, coinsToDistribute.Amount.ToDec().Mul(cr.Weight).RoundInt())
		// TODO check balance, log if insufficient and return rather than error
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, cr.Address, sdk.NewCoins(coinsToSend))
		if err != nil {
			return err
		}
		neg, updatedCoins := safeSub(coinsToDistribute, coinsToSend)
		if neg {
			return fmt.Errorf("negative coins")
		}
		coinsToDistribute = updatedCoins
	}
	return nil
}

func safeSub(a, b sdk.Coin) (bool, sdk.Coin) {
	isNeg := a.IsLT(b)
	if isNeg {
		return true, sdk.Coin{}
	}
	return false, a.Sub(b)
}
