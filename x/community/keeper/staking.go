package keeper

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/community/types"
)

const nanosecondsInOneSecond = int64(1000000000)

// PayoutAccumulatedStakingRewards calculates and transfers taking rewards to the fee collector address
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

	// get the denom for staking
	stakingRewardDenom := k.stakingKeeper.BondDenom(ctx)

	// we fetch the community pool balance to ensure only accumulate rewards up to the current balance
	communityPoolBalance := sdkmath.LegacyNewDecFromInt(k.bankKeeper.GetBalance(ctx, k.moduleAddress, stakingRewardDenom).Amount)

	// calculate staking reward payout capped to community pool balance
	truncatedRewards, truncationError := calculateStakingRewards(
		currentBlockTime,
		state.LastAccumulationTime,
		state.LastTruncationError,
		params.StakingRewardsPerSecond,
		communityPoolBalance,
	)

	// only payout if the truncated rewards are non-zero
	if !truncatedRewards.IsZero() {
		transferAmount := sdk.NewCoins(sdk.NewCoin(stakingRewardDenom, truncatedRewards))

		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleAccountName, authtypes.FeeCollectorName, transferAmount); err != nil {
			// we check for a valid balance and rewards can never be negative so panic since this will only
			// occur in cases where the chain is running in an invalid state
			panic(err)
		}

		// emit event with amount transferred
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeStakingRewardsPaid,
				sdk.NewAttribute(types.AttributeKeyStakingRewardAmount, transferAmount.String()),
			),
		)

	}

	// update accumulation state
	state.LastAccumulationTime = currentBlockTime
	// if the community pool balance is zero, this also resets the truncation error
	state.LastTruncationError = truncationError

	// save state
	k.SetStakingRewardsState(ctx, state)
}

// calculateStakingRewards takees the currentBlockTime, state of last accumulation, rewards per second, and the community pool balance
// in order to calculate the total payout since the last accumulation time.  It returns the truncated payout amount and the truncation error.
func calculateStakingRewards(currentBlockTime, lastAccumulationTime time.Time, lastTruncationError, stakingRewardsPerSecond, communityPoolBalance sdkmath.LegacyDec) (sdkmath.Int, sdkmath.LegacyDec) {
	// we get the duration since we last accumulated, then use nanoseconds for full precision available
	durationSinceLastPayout := currentBlockTime.Sub(lastAccumulationTime)
	nanosecondsSinceLastPayout := sdkmath.LegacyNewDec(durationSinceLastPayout.Nanoseconds())

	// We multiply by nanoseconds first, then divide by conversion to avoid loss of precision.
	// This multiplicaiton is also tested against very large values so we are safe from overflow
	// in normal operations.
	accumulatedRewards := nanosecondsSinceLastPayout.Mul(stakingRewardsPerSecond).QuoInt64(nanosecondsInOneSecond)
	// Ensure we add any error from previous truncations
	accumulatedRewards = accumulatedRewards.Add(lastTruncationError)

	if communityPoolBalance.LT(accumulatedRewards) {
		accumulatedRewards = communityPoolBalance
	}

	// we truncate since we can only transfer whole units
	truncatedRewards := accumulatedRewards.TruncateDec()
	// the truncation error to carry over to the next accumulation
	truncationError := accumulatedRewards.Sub(truncatedRewards)

	return truncatedRewards.TruncateInt(), truncationError
}
