package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

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
