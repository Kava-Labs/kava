package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

func (k Keeper) CollectStakingRewards(
	ctx sdk.Context,
	validator sdk.ValAddress,
	destinationModAccount string,
) (sdk.Coins, error) {
	macc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName)

	// Ensure withdraw address is as expected
	withdrawAddr := k.distributionKeeper.GetDelegatorWithdrawAddr(ctx, macc.GetAddress())
	if !withdrawAddr.Equals(macc.GetAddress()) {
		panic(fmt.Sprintf(
			"unexpected withdraw address for liquid staking module account, expected %s, got %s",
			macc.GetAddress(), withdrawAddr,
		))
	}

	rewards, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, macc.GetAddress(), validator)
	if err != nil {
		return nil, err
	}

	if rewards.IsZero() {
		return rewards, nil
	}

	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, destinationModAccount, rewards)
	if err != nil {
		return nil, err
	}

	return rewards, nil
}

func (k Keeper) CollectStakingRewardsByDenom(
	ctx sdk.Context,
	derivativeDenom string,
	destinationModAccount string,
) (sdk.Coins, error) {
	valAddr, err := types.ParseLiquidStakingTokenDenom(derivativeDenom)
	if err != nil {
		return nil, err
	}

	return k.CollectStakingRewards(ctx, valAddr, destinationModAccount)
}
