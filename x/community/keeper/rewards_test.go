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
		inflation    sdk.Dec
		communityTax sdk.Dec
		perSecReward sdk.Dec
		expectedRate sdk.Dec
	}{
		{
			name:         "no inflation, no rewards per sec -> 0%",
			totalSupply:  sdkmath.ZeroInt(),
			totalBonded:  sdkmath.ZeroInt(),
			inflation:    sdk.ZeroDec(),
			communityTax: sdk.ZeroDec(),
			perSecReward: sdk.ZeroDec(),
			expectedRate: sdk.ZeroDec(),
		},

		// inflation-only sanity checks
		{
			name:         "inflation only: no bonded tokens -> 0%",
			totalSupply:  sdk.NewInt(42),
			totalBonded:  sdkmath.ZeroInt(),
			inflation:    sdk.OneDec(),
			communityTax: sdk.ZeroDec(),
			perSecReward: sdk.ZeroDec(),
			expectedRate: sdk.ZeroDec(),
		},
		{
			name:         "inflation only: 0% inflation -> 0%",
			totalSupply:  sdk.NewInt(123),
			totalBonded:  sdkmath.NewInt(45),
			inflation:    sdk.ZeroDec(),
			communityTax: sdk.ZeroDec(),
			perSecReward: sdk.ZeroDec(),
			expectedRate: sdk.ZeroDec(),
		},
		{
			name:         "inflation only: 100% bonded w/ 100% inflation -> 100%",
			totalSupply:  sdk.NewInt(42),
			totalBonded:  sdk.NewInt(42),
			inflation:    sdk.OneDec(),
			communityTax: sdk.ZeroDec(),
			perSecReward: sdk.ZeroDec(),
			expectedRate: sdk.OneDec(),
		},
		{
			name:         "inflation only: 100% community tax -> 0%",
			totalSupply:  sdk.NewInt(123),
			totalBonded:  sdkmath.NewInt(45),
			inflation:    sdk.MustNewDecFromStr("0.853"),
			communityTax: sdk.OneDec(),
			perSecReward: sdk.ZeroDec(),
			expectedRate: sdk.ZeroDec(),
		},
		{
			name:         "inflation only: Oct 2023 case",
			totalSupply:  sdk.NewInt(857570000),
			totalBonded:  sdk.NewInt(127680000),
			inflation:    sdk.MustNewDecFromStr("0.595"),
			communityTax: sdk.MustNewDecFromStr("0.9495"),
			perSecReward: sdk.ZeroDec(),
			// expect 20.18% staking reward
			expectedRate: sdk.MustNewDecFromStr("0.201815746984649124"),
		},

		// rewards-only sanity checks
		{
			name:         "rps only: no bonded tokens -> 0%",
			totalSupply:  sdk.NewInt(42),
			totalBonded:  sdkmath.ZeroInt(),
			inflation:    sdk.ZeroDec(),
			communityTax: sdk.ZeroDec(),
			perSecReward: sdk.MustNewDecFromStr("1234567.123456"),
			expectedRate: sdk.ZeroDec(),
		},
		{
			name:         "rps only: rps = total bonded / seconds in year -> basically 100%",
			totalSupply:  sdk.NewInt(12345),
			totalBonded:  sdkmath.NewInt(1234),
			inflation:    sdk.ZeroDec(),
			communityTax: sdk.ZeroDec(),
			perSecReward: sdk.NewDec(1234).Quo(sdk.NewDec(keeper.SecondsPerYear)),
			expectedRate: sdk.MustNewDecFromStr("0.999999999999987228"), // <-- for 6-decimal token, this is negligible rounding
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
