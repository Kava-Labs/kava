package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/keeper"
)

func TestStakingRewardsCalculator(t *testing.T) {
	testCases := []struct {
		name         string
		totalSupply  sdkmath.Int
		totalBonded  sdkmath.Int
		inflation    sdkmath.LegacyDec
		communityTax sdkmath.LegacyDec
		perSecReward sdkmath.LegacyDec
		expectedRate sdkmath.LegacyDec
	}{
		{
			name:         "no inflation, no rewards per sec -> 0%",
			totalSupply:  sdkmath.ZeroInt(),
			totalBonded:  sdkmath.ZeroInt(),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyZeroDec(),
			expectedRate: sdkmath.LegacyZeroDec(),
		},

		// inflation-only sanity checks
		{
			name:         "inflation only: no bonded tokens -> 0%",
			totalSupply:  sdk.NewInt(42),
			totalBonded:  sdkmath.ZeroInt(),
			inflation:    sdkmath.LegacyOneDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyZeroDec(),
			expectedRate: sdkmath.LegacyZeroDec(),
		},
		{
			name:         "inflation only: 0% inflation -> 0%",
			totalSupply:  sdk.NewInt(123),
			totalBonded:  sdkmath.NewInt(45),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyZeroDec(),
			expectedRate: sdkmath.LegacyZeroDec(),
		},
		{
			name:         "inflation only: 100% bonded w/ 100% inflation -> 100%",
			totalSupply:  sdk.NewInt(42),
			totalBonded:  sdk.NewInt(42),
			inflation:    sdkmath.LegacyOneDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyZeroDec(),
			expectedRate: sdkmath.LegacyOneDec(),
		},
		{
			name:         "inflation only: 100% community tax -> 0%",
			totalSupply:  sdk.NewInt(123),
			totalBonded:  sdkmath.NewInt(45),
			inflation:    sdkmath.LegacyMustNewDecFromStr("0.853"),
			communityTax: sdkmath.LegacyOneDec(),
			perSecReward: sdkmath.LegacyZeroDec(),
			expectedRate: sdkmath.LegacyZeroDec(),
		},
		{
			name:         "inflation only: Oct 2023 case",
			totalSupply:  sdk.NewInt(857570000),
			totalBonded:  sdk.NewInt(127680000),
			inflation:    sdkmath.LegacyMustNewDecFromStr("0.595"),
			communityTax: sdkmath.LegacyMustNewDecFromStr("0.9495"),
			perSecReward: sdkmath.LegacyZeroDec(),
			// expect 20.18% staking reward
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.201815746984649124"),
		},

		// rewards-only sanity checks
		{
			name:         "rps only: no bonded tokens -> 0%",
			totalSupply:  sdk.NewInt(42),
			totalBonded:  sdkmath.ZeroInt(),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyMustNewDecFromStr("1234567.123456"),
			expectedRate: sdkmath.LegacyZeroDec(),
		},
		{
			name:         "rps only: rps = total bonded / seconds in year -> basically 100%",
			totalSupply:  sdk.NewInt(12345),
			totalBonded:  sdkmath.NewInt(1234),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyNewDec(1234).Quo(sdkmath.LegacyNewDec(keeper.SecondsPerYear)),
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.999999999999987228"), // <-- for 6-decimal token, this is negligible rounding
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calc := keeper.StakingRewardCalculator{
				TotalSupply:      tc.totalSupply,
				TotalBonded:      tc.totalBonded,
				InflationRate:    tc.inflation,
				CommunityTax:     tc.communityTax,
				RewardsPerSecond: tc.perSecReward,
			}

			rewardRate := calc.GetAnnualizedRate()
			require.Equal(t, tc.expectedRate, rewardRate)
		})
	}
}
