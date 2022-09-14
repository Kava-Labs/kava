package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

func (k Keeper) CollectStakingRewards(
	ctx sdk.Context,
	validator sdk.ValAddress,
	destinationModAccount string,
) (sdk.Coins, error) {
	modAcc := authtypes.NewModuleAddress(types.ModuleAccountName)
	k.Logger(ctx).Info("claimclaimclaim delegator:'" + modAcc.String() + "'")

	// ensure withdraw address is as expected
	withdrawAddr := k.distributionKeeper.GetDelegatorWithdrawAddr(ctx, modAcc)
	if !withdrawAddr.Equals(modAcc) {
		// TODO log error somewhere / panic? This case shouldn't happen.
		panic("unexpected withdraw address for liquid staking module account")
	}

	rewards, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, modAcc, validator)
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, destinationModAccount, rewards)
	if err != nil {
		panic(err) // TODO shouldn't happen?
	}
	return rewards, nil
}

func (k Keeper) CollectStakingRewardsByDenom(ctx sdk.Context, derivativeDenom string, destinationModAccount string) (sdk.Coins, error) {
	valAddr, err := types.ParseLiquidStakingTokenDenom(derivativeDenom)
	if err != nil {
		return nil, err
	}
	k.Logger(ctx).Info("claimclaimclaim validator: '" + valAddr.String() + "'")
	return k.CollectStakingRewards(ctx, valAddr, destinationModAccount)
}
