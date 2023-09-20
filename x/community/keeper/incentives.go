package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/types"
)

// CanPayoutStakingRewards return true if payout staking rewards can be paid out
func (k Keeper) CanPayoutStakingRewards(ctx sdk.Context) bool {
	return k.hasExecutedDisableInflationUpgrade(ctx)
}

// PayoutStakingRewards pays out staking rewards from the community account
// to the fee collector account for distribution to stakers.
func (k Keeper) PayoutStakingRewards(ctx sdk.Context) error {
	params, found := k.GetParams(ctx)
	if !found {
		return errorsmod.Wrapf(types.ErrStakingRewardsPayout, "params not found")
	}

	stakingRewards, err := k.calculateStakingRewards(ctx)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, k.feeCollectorName, sdk.NewCoins(stakingRewards)); err != nil {
		return err
	}

	if stakingRewards.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(stakingRewards.Amount.Int64()), "payout_staking_rewards")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePayoutRewards,
			sdk.NewAttribute(types.AttributeKeyRewardsPerSecond, params.StakingRewardsPerSecond.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, stakingRewards.Amount.String()),
		),
	)

	return nil
}

func (k Keeper) calculateStakingRewards(ctx sdk.Context) (sdk.Coin, error) {
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrStakingRewardsPayout, "missing previous block time")
	}

	params, found := k.GetParams(ctx)
	if !found {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrStakingRewardsPayout, "x/community params not found")
	}

	blockTime := ctx.BlockTime()
	secondsElapsed := sdkmath.NewInt(blockTime.Unix() - previousBlockTime.Unix())
	if secondsElapsed.IsNegative() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrStakingRewardsPayout, "negative block duration")
	}

	ukavaRewards := params.StakingRewardsPerSecond.Mul(secondsElapsed)

	return sdk.NewCoin("ukava", ukavaRewards), nil
}

// ShouldStartDisableInflationUpgrade returns true if the disable inflation upgrade should be started
func (k Keeper) ShouldStartDisableInflationUpgrade(ctx sdk.Context) bool {
	params, found := k.GetParams(ctx)
	if !found {
		return false
	}

	if k.hasExecutedDisableInflationUpgrade(ctx) {
		return false
	}

	blockTime := ctx.BlockTime()
	upgradeTime := params.UpgradeTimeDisableInflation

	// a vanilla kava chain should disable inflation on the first block if `upgradeTime` is not set.
	if upgradeTime.IsZero() {
		return true
	}

	return !blockTime.Before(upgradeTime)
}

// StartDisableInflationUpgrade disables x/mint and x/kavadist inflation
func (k Keeper) StartDisableInflationUpgrade(ctx sdk.Context) {
	logger := k.Logger(ctx)

	blockTime := ctx.BlockTime()

	logger.Info("disable inflation upgrade started")
	k.SetPreviousBlockTime(ctx, blockTime)

	// set x/min inflation to 0
	mintParams := k.mintKeeper.GetParams(ctx)
	mintParams.InflationMax = sdk.ZeroDec()
	mintParams.InflationMin = sdk.ZeroDec()
	k.mintKeeper.SetParams(ctx, mintParams)
	logger.Info("x/mint inflation set to 0")

	// disable kavadist inflation
	kavadistParams := k.kavadistKeeper.GetParams(ctx)
	kavadistParams.Active = false
	k.kavadistKeeper.SetParams(ctx, kavadistParams)

	logger.Info("x/kavadist inflation disabled")

	logger.Info("disable inflation upgrade finished successfully!")
}

func (k Keeper) hasExecutedDisableInflationUpgrade(ctx sdk.Context) bool {
	// previous block time is only set after we have started the disable inflation upgrade
	_, found := k.GetPreviousBlockTime(ctx)
	return found
}
