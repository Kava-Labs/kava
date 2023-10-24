package keeper_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/community/keeper"
)

func TestStakingRewardsCalculator(t *testing.T) {
	hugeInflation := new(big.Int).Exp(big.NewInt(2), big.NewInt(205), nil)
	hugeRewardsPerSec := new(big.Int).Exp(big.NewInt(2), big.NewInt(230), nil)

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
		//
		//
		// inflation-only
		//
		//
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
			totalSupply:  sdk.NewInt(857570000e6),
			totalBonded:  sdk.NewInt(127680000e6),
			inflation:    sdkmath.LegacyMustNewDecFromStr("0.595"),
			communityTax: sdkmath.LegacyMustNewDecFromStr("0.9495"),
			perSecReward: sdkmath.LegacyZeroDec(),
			// expect 20.18% staking reward
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.201815746984649122"), // verified manually
		},
		{
			name:         "inflation only: low inflation",
			totalSupply:  sdk.NewInt(857570000e6),
			totalBonded:  sdk.NewInt(127680000e6),
			inflation:    sdkmath.LegacyMustNewDecFromStr("0.0000000001"),
			communityTax: sdkmath.LegacyMustNewDecFromStr("0.9495"),
			perSecReward: sdkmath.LegacyZeroDec(),
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.000000000033918612"), // verified manually, rounded would be 0.000000000033918613
		},
		{
			name:         "inflation only: absurdly high inflation",
			totalSupply:  sdk.NewInt(857570000e6),
			totalBonded:  sdk.NewInt(127680000e6),
			inflation:    sdkmath.LegacyNewDecFromBigInt(hugeInflation), // 2^205. a higher exponent than this overflows.
			communityTax: sdkmath.LegacyMustNewDecFromStr("0.9495"),
			perSecReward: sdkmath.LegacyZeroDec(),
			// https://www.wolframalpha.com/input?i=%282%5E205%29+*+%281+-+0.9495%29+*+%28857570000e6+%2F127680000e6%29
			expectedRate: sdkmath.LegacyMustNewDecFromStr("17441635052648297161685283657196753398188161373334495592570113.113824561403508771"), // verified manually, would round up
		},
		//
		//
		// rewards-only
		//
		//
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
		{
			name:         "rps only: 10M kava / year rewards",
			totalSupply:  sdk.NewInt(870950000e6),
			totalBonded:  sdkmath.NewInt(130380000e6),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyMustNewDecFromStr("317097.919837645865043125"), // 10 million kava per year
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.076698880196349133"),      // verified manually
		},
		{
			name:         "rps only: 25M kava / year rewards",
			totalSupply:  sdk.NewInt(870950000e6),
			totalBonded:  sdkmath.NewInt(130380000e6),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyMustNewDecFromStr("792744.799594114662607813"), // 25 million kava per year
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.191747200490872833"),      // verified manually
		},
		{
			name:         "rps only: too much kava / year rewards",
			totalSupply:  sdk.NewInt(870950000e6),
			totalBonded:  sdkmath.NewInt(130380000e6),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyNewDecFromBigInt(hugeRewardsPerSec), // 2^230. a higher exponent than this overflows.
			// https://www.wolframalpha.com/input?i=%28%28365+*+24+*+3600%29+%2F+130380000e6%29+*+%282%5E230%29
			expectedRate: sdkmath.LegacyMustNewDecFromStr("417344440850566075319340506352140425426634017001007267992800590.431305795858260469"), // verified manually
		},
		{
			name:         "rps only: low kava / year rewards",
			totalSupply:  sdk.NewInt(870950000e6),
			totalBonded:  sdkmath.NewInt(130380000e6),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyMustNewDecFromStr("0.1"),
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.000000024187758858"), // verified manually, rounded would be 0.000000024187758859
		},
		{
			name:         "rps only: 1 ukava / year rewards",
			totalSupply:  sdk.NewInt(870950000e6),
			totalBonded:  sdkmath.NewInt(130380000e6),
			inflation:    sdkmath.LegacyZeroDec(),
			communityTax: sdkmath.LegacyZeroDec(),
			perSecReward: sdkmath.LegacyMustNewDecFromStr("0.000000031709791984"), // 1 ukava per year
			expectedRate: sdkmath.LegacyMustNewDecFromStr("0.000000000000007669"), // verified manually, rounded would be 0.000000000000007670
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rewardRate := keeper.CalculateStakingAnnualPercentage(
				tc.totalSupply,
				tc.totalBonded,
				tc.inflation,
				tc.communityTax,
				tc.perSecReward)
			require.Equal(t, tc.expectedRate, rewardRate)
		})
	}
}
