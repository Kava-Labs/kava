package main

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	v0_15cdp "github.com/kava-labs/kava/x/cdp/types"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

func TestCalculateRewardsForClaim(t *testing.T) {
	type expect struct {
		amount sdk.Int
		err    bool
	}
	testCases := []struct {
		name          string
		claimIndexes  v0_15incentive.RewardIndexes
		globalIndexes v0_15incentive.RewardIndexes
		cdps          v0_15cdp.CDPs
		expected      expect
	}{
		{
			name: "single cdp is synced",
			claimIndexes: v0_15incentive.RewardIndexes{
				{
					CollateralType: "busd-a",
					RewardFactor:   sdk.MustNewDecFromStr("0.1"),
				},
			},
			globalIndexes: v0_15incentive.RewardIndexes{
				{
					CollateralType: "busd-a",
					RewardFactor:   sdk.MustNewDecFromStr("0.2"),
				},
			},
			cdps: v0_15cdp.CDPs{{
				Owner:           sdk.AccAddress("address1"),
				Type:            "busd-a",
				Principal:       sdk.NewInt64Coin("usdx", 1000),
				AccumulatedFees: sdk.NewInt64Coin("usdx", 10),
			}},
			expected: expect{amount: sdk.NewInt(101)},
		},
		{
			name:         "missing claim index is assumed 0",
			claimIndexes: v0_15incentive.RewardIndexes{},
			globalIndexes: v0_15incentive.RewardIndexes{
				{
					CollateralType: "busd-a",
					RewardFactor:   sdk.MustNewDecFromStr("0.2"),
				},
			},
			cdps: v0_15cdp.CDPs{{
				Owner:           sdk.AccAddress("address1"),
				Type:            "busd-a",
				Principal:       sdk.NewInt64Coin("usdx", 1000),
				AccumulatedFees: sdk.NewInt64Coin("usdx", 10),
			}},
			expected: expect{amount: sdk.NewInt(202)},
		},
		{
			name: "multiple cdps are synced",
			claimIndexes: v0_15incentive.RewardIndexes{
				{
					CollateralType: "busd-a",
					RewardFactor:   sdk.MustNewDecFromStr("0.1"),
				},
				{
					CollateralType: "xrpb-a",
					RewardFactor:   sdk.MustNewDecFromStr("1"),
				},
			},
			globalIndexes: v0_15incentive.RewardIndexes{
				{
					CollateralType: "busd-a",
					RewardFactor:   sdk.MustNewDecFromStr("0.2"),
				},
				{
					CollateralType: "xrpb-a",
					RewardFactor:   sdk.MustNewDecFromStr("2"),
				},
			},
			cdps: v0_15cdp.CDPs{
				{
					Owner:           sdk.AccAddress("address1"),
					Type:            "busd-a",
					Principal:       sdk.NewInt64Coin("usdx", 1000),
					AccumulatedFees: sdk.NewInt64Coin("usdx", 100),
				},
				{
					Owner:           sdk.AccAddress("address1"),
					Type:            "xrpb-a",
					Principal:       sdk.NewInt64Coin("usdx", 10),
					AccumulatedFees: sdk.NewInt64Coin("usdx", 1),
				}},
			expected: expect{amount: sdk.NewInt(110 + 11)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc := rewardsCalculator{
				globalIndexes: tc.globalIndexes,
				cdps:          tc.cdps,
			}
			reward, err := rc.calculateRewardsForClaim(v0_15incentive.NewUSDXMintingClaim(
				sdk.AccAddress("address1"),
				sdk.NewInt64Coin("ukava", 0),
				tc.claimIndexes,
			))

			if tc.expected.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, v0_15incentive.USDXMintingRewardDenom, reward.Denom)
				require.Truef(t, tc.expected.amount.Equal(reward.Amount), "amount not equal %s, %s", tc.expected.amount, reward.Amount)
			}
		})
	}
}
