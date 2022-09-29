package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestAccumulator(t *testing.T) {
	t.Run("getTimeElapsedWithinLimits", func(t *testing.T) {
		type args struct {
			start, end         time.Time
			limitMin, limitMax time.Time
		}
		testcases := []struct {
			name     string
			args     args
			expected time.Duration
		}{
			{
				name: "given time range is before limits and is non zero, return 0 duration",
				args: args{
					start:    time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					end:      time.Date(1998, 1, 1, 0, 0, 1, 0, time.UTC),
					limitMin: time.Date(2098, 1, 1, 0, 0, 0, 0, time.UTC),
					limitMax: time.Date(2098, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				expected: 0,
			},
			{
				name: "given time range is after limits and is non zero, return 0 duration",
				args: args{
					start:    time.Date(2098, 1, 1, 0, 0, 0, 0, time.UTC),
					end:      time.Date(2098, 1, 1, 0, 0, 1, 0, time.UTC),
					limitMin: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					limitMax: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				expected: 0,
			},
			{
				name: "given time range is within limits and is non zero, return duration",
				args: args{
					start:    time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					end:      time.Date(1998, 1, 1, 0, 0, 1, 0, time.UTC),
					limitMin: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					limitMax: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				expected: time.Second,
			},
			{
				name: "given time range is within limits and is zero, return 0 duration",
				args: args{
					start:    time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					end:      time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					limitMin: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					limitMax: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				expected: 0,
			},
			{
				name: "given time range overlaps limitMax and is non zero, return capped duration",
				args: args{
					start:    time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					end:      time.Date(1998, 1, 1, 0, 0, 2, 0, time.UTC),
					limitMin: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					limitMax: time.Date(1998, 1, 1, 0, 0, 1, 0, time.UTC),
				},
				expected: time.Second,
			},
			{
				name: "given time range overlaps limitMin and is non zero, return capped duration",
				args: args{
					start:    time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					end:      time.Date(1998, 1, 1, 0, 0, 2, 0, time.UTC),
					limitMin: time.Date(1998, 1, 1, 0, 0, 1, 0, time.UTC),
					limitMax: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				expected: time.Second,
			},
			{
				name: "given time range is larger than limits, return capped duration",
				args: args{
					start:    time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					end:      time.Date(1998, 1, 1, 0, 0, 10, 0, time.UTC),
					limitMin: time.Date(1998, 1, 1, 0, 0, 1, 0, time.UTC),
					limitMax: time.Date(1998, 1, 1, 0, 0, 9, 0, time.UTC),
				},
				expected: 8 * time.Second,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				acc := &Accumulator{}
				duration := acc.getTimeElapsedWithinLimits(tc.args.start, tc.args.end, tc.args.limitMin, tc.args.limitMax)

				require.Equal(t, tc.expected, duration)
			})
		}
	})
	t.Run("calculateNewRewards", func(t *testing.T) {
		type args struct {
			rewardsPerSecond  sdk.Coins
			duration          time.Duration
			totalSourceShares sdk.Dec
		}
		testcases := []struct {
			name     string
			args     args
			expected RewardIndexes
		}{
			{
				name: "rewards calculated normally",
				args: args{
					rewardsPerSecond:  cs(c("hard", 1000), c("swap", 100)),
					duration:          10 * time.Second,
					totalSourceShares: d("1000"),
				},
				expected: RewardIndexes{
					{CollateralType: "hard", RewardFactor: d("10")},
					{CollateralType: "swap", RewardFactor: d("1")},
				},
			},
			{
				name: "duration is rounded to nearest even second",
				args: args{
					rewardsPerSecond:  cs(c("hard", 1000)),
					duration:          10*time.Second + 500*time.Millisecond,
					totalSourceShares: d("1000"),
				},
				expected: RewardIndexes{
					{CollateralType: "hard", RewardFactor: d("10")},
				},
			},
			{
				name: "reward indexes have enough precision for extreme params",
				args: args{
					rewardsPerSecond:  cs(c("anydenom", 1)),    // minimum possible rewards
					duration:          1 * time.Second,         // minimum possible duration (beyond zero as it's rounded)
					totalSourceShares: d("100000000000000000"), // approximate shares in a $1B pool of 10^8 precision assets
				},
				expected: RewardIndexes{
					// smallest reward amount over smallest accumulation duration does not go past 10^-18 decimal precision
					{CollateralType: "anydenom", RewardFactor: d("0.000000000000000010")},
				},
			},
			{
				name: "when duration is zero there is no rewards",
				args: args{
					rewardsPerSecond:  cs(c("hard", 1000)),
					duration:          0,
					totalSourceShares: d("1000"),
				},
				expected: nil,
			},
			{
				name: "when rewards per second are nil there is no rewards",
				args: args{
					rewardsPerSecond:  cs(),
					duration:          10 * time.Second,
					totalSourceShares: d("1000"),
				},
				expected: nil,
			},
			{
				name: "when the source total is zero there is no rewards",
				args: args{
					rewardsPerSecond:  cs(c("hard", 1000)),
					duration:          10 * time.Second,
					totalSourceShares: d("0"),
				},
				expected: nil,
			},
			{
				name: "when all args are zero there is no rewards",
				args: args{
					rewardsPerSecond:  cs(),
					duration:          0,
					totalSourceShares: d("0"),
				},
				expected: nil,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				acc := &Accumulator{}
				indexes := acc.calculateNewRewards(
					sdk.NewDecCoinsFromCoins(tc.args.rewardsPerSecond...),
					tc.args.totalSourceShares,
					tc.args.duration,
				)

				require.Equal(t, tc.expected, indexes)
			})
		}
	})
	t.Run("Accumulate", func(t *testing.T) {
		type args struct {
			accumulator       Accumulator
			period            MultiRewardPeriod
			totalSourceShares sdk.Dec
			currentTime       time.Time
		}
		testcases := []struct {
			name     string
			args     args
			expected Accumulator
		}{
			{
				name: "normal",
				args: args{
					accumulator: Accumulator{
						PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
						Indexes: RewardIndexes{
							{CollateralType: "hard", RewardFactor: d("0.1")},
							{CollateralType: "swap", RewardFactor: d("0.2")},
						},
					},
					period: MultiRewardPeriod{
						Start:            time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
						End:              time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						RewardsPerSecond: cs(c("hard", 1000)),
					},
					totalSourceShares: d("1000"),
					currentTime:       time.Date(1998, 1, 1, 0, 0, 5, 0, time.UTC),
				},
				expected: Accumulator{
					PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 5, 0, time.UTC),
					Indexes: RewardIndexes{
						{CollateralType: "hard", RewardFactor: d("5.1")},
						{CollateralType: "swap", RewardFactor: d("0.2")},
					},
				},
			},
			{
				name: "empty reward indexes are added to correctly",
				args: args{
					accumulator: Accumulator{
						PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
						Indexes:                  RewardIndexes{},
					},
					period: MultiRewardPeriod{
						Start:            time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
						End:              time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						RewardsPerSecond: cs(c("hard", 1000)),
					},
					totalSourceShares: d("1000"),
					currentTime:       time.Date(1998, 1, 1, 0, 0, 5, 0, time.UTC),
				},
				expected: Accumulator{
					PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 5, 0, time.UTC),
					Indexes:                  RewardIndexes{{CollateralType: "hard", RewardFactor: d("5.0")}},
				},
			},
			{
				name: "empty reward indexes are unchanged when there's no rewards",
				args: args{
					accumulator: Accumulator{
						PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
						Indexes:                  RewardIndexes{},
					},
					period: MultiRewardPeriod{
						Start:            time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
						End:              time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						RewardsPerSecond: cs(),
					},
					totalSourceShares: d("1000"),
					currentTime:       time.Date(1998, 1, 1, 0, 0, 5, 0, time.UTC),
				},
				expected: Accumulator{
					PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 5, 0, time.UTC),
					Indexes:                  RewardIndexes{},
				},
			},
			{
				name: "when a period is enclosed within block the accumulation time is set to the period end time",
				args: args{
					accumulator: Accumulator{
						PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
						Indexes:                  RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}},
					},
					period: MultiRewardPeriod{
						Start:            time.Date(1998, 1, 1, 0, 0, 5, 0, time.UTC),
						End:              time.Date(1998, 1, 1, 0, 0, 7, 0, time.UTC),
						RewardsPerSecond: cs(c("hard", 1000)),
					},
					totalSourceShares: d("1000"),
					currentTime:       time.Date(1998, 1, 1, 0, 0, 10, 0, time.UTC),
				},
				expected: Accumulator{
					PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 7, 0, time.UTC),
					Indexes:                  RewardIndexes{{CollateralType: "hard", RewardFactor: d("2.1")}},
				},
			},
			{
				name: "accumulation duration is capped at param start when previous stored time is in the distant past",
				// This could happend in the default time value time.Time{} was accidentally stored, or if a reward period was
				// removed from the params, then added back a long time later.
				args: args{
					accumulator: Accumulator{
						PreviousAccumulationTime: time.Time{},
						Indexes:                  RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}},
					},
					period: MultiRewardPeriod{
						Start:            time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
						End:              time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						RewardsPerSecond: cs(c("hard", 1000)),
					},
					totalSourceShares: d("1000"),
					currentTime:       time.Date(1998, 1, 1, 0, 0, 10, 0, time.UTC),
				},
				expected: Accumulator{
					PreviousAccumulationTime: time.Date(1998, 1, 1, 0, 0, 10, 0, time.UTC),
					Indexes:                  RewardIndexes{{CollateralType: "hard", RewardFactor: d("10.1")}},
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				tc.args.accumulator.Accumulate(tc.args.period, tc.args.totalSourceShares, tc.args.currentTime)
				require.Equal(t, tc.expected, tc.args.accumulator)
			})
		}
	})
}

func TestMinTime(t *testing.T) {
	type args struct {
		t1, t2 time.Time
	}
	testcases := []struct {
		name     string
		args     args
		expected time.Time
	}{
		{
			name: "last arg greater than first",
			args: args{
				t1: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
				t2: time.Date(1998, 1, 1, 0, 0, 0, 1, time.UTC),
			},
			expected: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "first arg greater than last",
			args: args{
				t2: time.Date(1998, 1, 1, 0, 0, 0, 1, time.UTC),
				t1: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "first and last args equal",
			args: args{
				t2: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
				t1: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, minTime(tc.args.t1, tc.args.t2))
		})
	}
}

func TestMaxTime(t *testing.T) {
	type args struct {
		t1, t2 time.Time
	}
	testcases := []struct {
		name     string
		args     args
		expected time.Time
	}{
		{
			name: "last arg greater than first",
			args: args{
				t1: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
				t2: time.Date(1998, 1, 1, 0, 0, 0, 1, time.UTC),
			},
			expected: time.Date(1998, 1, 1, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "first arg greater than last",
			args: args{
				t2: time.Date(1998, 1, 1, 0, 0, 0, 1, time.UTC),
				t1: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: time.Date(1998, 1, 1, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "first and last args equal",
			args: args{
				t2: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
				t1: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, maxTime(tc.args.t1, tc.args.t2))
		})
	}
}
