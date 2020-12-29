package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// func TestRewardPeriodsValidate(t *testing.T) {
// 	now := time.Now()

// 	testCases := []struct {
// 		msg           string
// 		rewardPeriods RewardPeriods
// 		expPass       bool
// 	}{
// 		{
// 			"valid",
// 			RewardPeriods{
// 				NewRewardPeriod(true, "bnb", now, now.Add(time.Hour), sdk.NewCoin("bnb", sdk.OneInt()), now, Multipliers{NewMultiplier(Small, 1, sdk.MustNewDecFromStr("0.33"))}),
// 			},
// 			true,
// 		},
// 		{
// 			"zero start time",
// 			RewardPeriods{
// 				{Start: time.Time{}},
// 			},
// 			false,
// 		},
// 		{
// 			"zero end time",
// 			RewardPeriods{
// 				{Start: now, End: time.Time{}},
// 			},
// 			false,
// 		},
// 		{
// 			"zero end time",
// 			RewardPeriods{
// 				{Start: now, End: time.Time{}},
// 			},
// 			false,
// 		},
// 		{
// 			"start time > end time",
// 			RewardPeriods{
// 				{
// 					Start: now.Add(time.Hour),
// 					End:   now,
// 				},
// 			},
// 			false,
// 		},
// 		{
// 			"invalid reward",
// 			RewardPeriods{
// 				{
// 					Start:  now,
// 					End:    now.Add(time.Hour),
// 					Reward: sdk.Coin{Denom: "", Amount: sdk.ZeroInt()},
// 				},
// 			},
// 			false,
// 		},
// 		{
// 			"zero claim end time",
// 			RewardPeriods{
// 				{
// 					Start:    now,
// 					End:      now.Add(time.Hour),
// 					Reward:   sdk.NewCoin("bnb", sdk.OneInt()),
// 					ClaimEnd: time.Time{},
// 				},
// 			},
// 			false,
// 		},
// 		{
// 			"negative time lock",
// 			RewardPeriods{
// 				{
// 					Start:            now,
// 					End:              now.Add(time.Hour),
// 					Reward:           sdk.NewCoin("bnb", sdk.OneInt()),
// 					ClaimEnd:         now,
// 					ClaimMultipliers: Multipliers{NewMultiplier(Small, -1, sdk.MustNewDecFromStr("0.33"))},
// 				},
// 			},
// 			false,
// 		},
// 		{
// 			"invalid collateral type",
// 			RewardPeriods{
// 				{
// 					Start:            now,
// 					End:              now.Add(time.Hour),
// 					Reward:           sdk.NewCoin("bnb", sdk.OneInt()),
// 					ClaimEnd:         now,
// 					ClaimMultipliers: Multipliers{NewMultiplier(Small, 1, sdk.MustNewDecFromStr("0.33"))},
// 					CollateralType:   "",
// 				},
// 			},
// 			false,
// 		},
// 		{
// 			"duplicate reward period",
// 			RewardPeriods{
// 				NewRewardPeriod("bnb", now, now.Add(time.Hour), sdk.NewCoin("bnb", sdk.OneInt()), now, Multipliers{NewMultiplier(Small, 1, sdk.MustNewDecFromStr("0.33"))}),
// 				NewRewardPeriod("bnb", now, now.Add(time.Hour), sdk.NewCoin("bnb", sdk.OneInt()), now, Multipliers{NewMultiplier(Small, 1, sdk.MustNewDecFromStr("0.33"))}),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		err := tc.rewardPeriods.Validate()
// 		if tc.expPass {
// 			require.NoError(t, err, tc.msg)
// 		} else {
// 			require.Error(t, err, tc.msg)
// 		}
// 	}
// }

// func TestClaimPeriodsValidate(t *testing.T) {
// 	now := time.Now()

// 	testCases := []struct {
// 		msg          string
// 		claimPeriods ClaimPeriods
// 		expPass      bool
// 	}{
// 		{
// 			"valid",
// 			ClaimPeriods{
// 				NewClaimPeriod("bnb", 10, now, Multipliers{NewMultiplier(Small, 1, sdk.MustNewDecFromStr("0.33"))}),
// 			},
// 			true,
// 		},
// 		{
// 			"invalid ID",
// 			ClaimPeriods{
// 				{ID: 0},
// 			},
// 			false,
// 		},
// 		{
// 			"zero end time",
// 			ClaimPeriods{
// 				{ID: 10, End: time.Time{}},
// 			},
// 			false,
// 		},
// 		{
// 			"negative time lock",
// 			ClaimPeriods{
// 				{ID: 10, End: now, ClaimMultipliers: Multipliers{NewMultiplier(Small, -1, sdk.MustNewDecFromStr("0.33"))}},
// 			},
// 			false,
// 		},
// 		{
// 			"negative multiplier",
// 			ClaimPeriods{
// 				NewClaimPeriod("bnb", 10, now, Multipliers{NewMultiplier(Small, 1, sdk.MustNewDecFromStr("-0.33"))}),
// 			},
// 			false,
// 		},
// 		{
// 			"start time > end time",
// 			ClaimPeriods{
// 				{ID: 10, End: now, ClaimMultipliers: Multipliers{NewMultiplier(Small, -1, sdk.MustNewDecFromStr("0.33"))}},
// 			},
// 			false,
// 		},
// 		{
// 			"invalid collateral type",
// 			ClaimPeriods{
// 				{ID: 10, End: now, ClaimMultipliers: Multipliers{NewMultiplier(Small, -1, sdk.MustNewDecFromStr("0.33"))}, CollateralType: ""},
// 			},
// 			false,
// 		},
// 		{
// 			"duplicate reward period",
// 			ClaimPeriods{
// 				NewClaimPeriod("bnb", 10, now, Multipliers{NewMultiplier(Small, -1, sdk.MustNewDecFromStr("0.33"))}),
// 				NewClaimPeriod("bnb", 10, now, Multipliers{NewMultiplier(Small, -1, sdk.MustNewDecFromStr("0.33"))}),
// 			},
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		err := tc.claimPeriods.Validate()
// 		if tc.expPass {
// 			require.NoError(t, err, tc.msg)
// 		} else {
// 			require.Error(t, err, tc.msg)
// 		}
// 	}
// }

func TestClaimsValidate(t *testing.T) {
	owner := sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser1")))

	testCases := []struct {
		msg     string
		claims  Claims
		expPass bool
	}{
		{
			"valid",
			Claims{
				NewClaim(owner, sdk.NewCoin("bnb", sdk.OneInt()), "bnb-a", RewardIndex{"ukava", sdk.ZeroDec()}),
			},
			true,
		},
		{
			"invalid owner",
			Claims{
				{Owner: nil},
			},
			false,
		},
		{
			"invalid reward",
			Claims{
				{
					Owner:  owner,
					Reward: sdk.Coin{Denom: "", Amount: sdk.ZeroInt()},
				},
			},
			false,
		},
		{
			"invalid collateral type",
			Claims{
				{
					Owner:          owner,
					Reward:         sdk.NewCoin("bnb", sdk.OneInt()),
					RewardIndex:    RewardIndex{"ukava", sdk.ZeroDec()},
					CollateralType: "",
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.claims.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}
