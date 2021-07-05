package types

import (
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Accumulator struct {
	PreviousAccumulationTime time.Time
	Indexes                  RewardIndexes
}

func NewAccumulator(previousAccrual time.Time, indexes RewardIndexes) *Accumulator {
	return &Accumulator{
		PreviousAccumulationTime: previousAccrual,
		Indexes:                  indexes,
	}
}

// Accumulate updates the reward indexes for a new block.
func (acc *Accumulator) Accumulate(period MultiRewardPeriod, rewardSourceTotal sdk.Dec, currentTime time.Time) {
	accumulationDuration := acc.getTimeElapsedWithinLimits(acc.PreviousAccumulationTime, currentTime, period.Start, period.End)
	indexesIncrement := acc.calculateNewRewards(period.RewardsPerSecond, rewardSourceTotal, accumulationDuration)

	acc.Indexes = acc.Indexes.Add(indexesIncrement)
	acc.PreviousAccumulationTime = minTime(period.End, currentTime)
}

// getTimeElapsedWithinLimits returns the duration between start and end times, capped by min and max times.
// If the start and end range is outside the min to max time range then zero duration is returned.
func (acc *Accumulator) getTimeElapsedWithinLimits(start, end, limitMin, limitMax time.Time) time.Duration {
	if start.After(end) {
		panic(fmt.Sprintf("start time (%s) cannot be after end time (%s)", start, end))
	}
	if limitMin.After(limitMax) {
		panic(fmt.Sprintf("minimum limit time (%s) cannot be after maximum limit time (%s)", limitMin, limitMax))
	}
	if start.After(limitMax) || end.Before(limitMin) {
		// no intersection between the start-end and limitMin-limitMax time ranges
		return 0
	}
	return minTime(end, limitMax).Sub(maxTime(start, limitMin))
}

// calculateNewRewards calculates the amount to increase the global reward indexes for a given reward rate, duration, and source total.
// The total rewards to distribute in this block are given by reward rate * duration. This update to global indexes is this value divided
// by the source total. This gives rewards per unit of source.
// Note, duration is rounded to the nearest second to keep rewards calculation the same as in kava-7.
func (acc *Accumulator) calculateNewRewards(rewardsPerSecond sdk.Coins, rewardSourceTotal sdk.Dec, duration time.Duration) RewardIndexes {
	if rewardSourceTotal.IsZero() {
		// When the source total is zero, there is no users with deposits/borrows/delegations to pay out the current block's rewards to.
		// So drop the rewards and pay out nothing.
		return nil
	}
	durationSeconds := int64(math.RoundToEven(duration.Seconds()))
	increment := newRewardIndexesFromCoins(rewardsPerSecond)
	return increment.Mul(sdk.NewDec(durationSeconds)).Quo(rewardSourceTotal)
}

// minTime returns the earliest of two times.
func minTime(t1, t2 time.Time) time.Time {
	if t2.Before(t1) {
		return t2
	}
	return t1
}

// maxTime returns the latest of two times.
func maxTime(t1, t2 time.Time) time.Time {
	if t2.After(t1) {
		return t2
	}
	return t1
}

// newRewardIndexesFromCoins is a helper function to initialize a RewardIndexes slice with the values from a Coins slice.
func newRewardIndexesFromCoins(coins sdk.Coins) RewardIndexes {
	var indexes RewardIndexes
	for _, coin := range coins {
		indexes = append(indexes, NewRewardIndex(coin.Denom, coin.Amount.ToDec()))
	}
	return indexes
}
