package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/community/types"
)

const nanosecondsInOneSecond = int64(1000000000)

func (k Keeper) PayoutAccumulatedStakingRewards(ctx sdk.Context) {
	// get module parameters which define the amount of rewards to payout per second
	params := k.mustGetParams(ctx)
	currentBlockTime := ctx.BlockTime()
	state := k.GetStakingRewardsState(ctx)

	// we have un-initialized state -- set accumulation time and exit since there is nothing to do
	if state.LastAccumulationTime.IsZero() {
		state.LastAccumulationTime = currentBlockTime

		k.SetStakingRewardsState(ctx, state)

		return
	}

	// we get the duration since we last accumulated, then use nanoseconds for full precision available
	durationSinceLastPayout := currentBlockTime.Sub(state.LastAccumulationTime)
	nanosecondsSinceLastPayout := sdkmath.LegacyNewDec(durationSinceLastPayout.Nanoseconds())

	// We multiply by nanoseconds first, then divide by conversion to avoid loss of precision.
	// This multiplicaiton is also tested against very large values so we are safe from overflow
	// in normal operations.
	accumulatedRewards := nanosecondsSinceLastPayout.Mul(params.StakingRewardsPerSecond).QuoInt64(nanosecondsInOneSecond)
	// Ensure we add any error from previous truncations
	accumulatedRewards = accumulatedRewards.Add(state.LastTruncationError)

	// If the community pool balance is less than the accumulated rewards we only accumulate rewards up
	// to the pool balance.
	poolBalance := sdkmath.LegacyNewDecFromInt(k.bankKeeper.GetBalance(ctx, k.moduleAddress, "ukava").Amount)
	if poolBalance.LT(accumulatedRewards) {
		accumulatedRewards = poolBalance
	}

	// we truncate since we can only transfer whole units
	truncatedRewards := accumulatedRewards.TruncateDec()
	// the truncation error to carry over to the next accumulation
	truncationError := accumulatedRewards.Sub(truncatedRewards)

	if !truncatedRewards.IsZero() {
		transferAmount := sdk.NewCoins(sdk.NewCoin("ukava", truncatedRewards.TruncateInt()))

		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, authtypes.FeeCollectorName, transferAmount); err != nil {
			// we check for a valid balance and rewards can never be negative so panic since this will only
			// occur in cases where the chain is running in an invalid state
			panic(err)
		}
	}

	// update accumulation state
	state.LastAccumulationTime = currentBlockTime
	state.LastTruncationError = truncationError

	// save state
	k.SetStakingRewardsState(ctx, state)

	return
}
