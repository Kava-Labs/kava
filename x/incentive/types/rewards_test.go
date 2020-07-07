package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestRewardPeriodsValidate(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		msg           string
		rewardPeriods RewardPeriods
		expPass       bool
	}{
		{
			"valid",
			RewardPeriods{
				NewRewardPeriod("bnb", now, now.Add(time.Hour), sdk.NewCoin("bnb", sdk.OneInt()), now, 10),
			},
			true,
		},
		{
			"zero start time",
			RewardPeriods{
				{Start: time.Time{}},
			},
			false,
		},
		{
			"zero end time",
			RewardPeriods{
				{Start: now, End: time.Time{}},
			},
			false,
		},
		{
			"zero end time",
			RewardPeriods{
				{Start: now, End: time.Time{}},
			},
			false,
		},
		{
			"start time > end time",
			RewardPeriods{
				{
					Start: now.Add(time.Hour),
					End:   now,
				},
			},
			false,
		},
		{
			"invalid reward",
			RewardPeriods{
				{
					Start:  now,
					End:    now.Add(time.Hour),
					Reward: sdk.Coin{Denom: "", Amount: sdk.ZeroInt()},
				},
			},
			false,
		},
		{
			"zero claim end time",
			RewardPeriods{
				{
					Start:    now,
					End:      now.Add(time.Hour),
					Reward:   sdk.NewCoin("bnb", sdk.OneInt()),
					ClaimEnd: time.Time{},
				},
			},
			false,
		},
		{
			"zero claim time lock",
			RewardPeriods{
				{
					Start:         now,
					End:           now.Add(time.Hour),
					Reward:        sdk.NewCoin("bnb", sdk.OneInt()),
					ClaimEnd:      now,
					ClaimTimeLock: 0,
				},
			},
			false,
		},
		{
			"invalid denom",
			RewardPeriods{
				{
					Start:         now,
					End:           now.Add(time.Hour),
					Reward:        sdk.NewCoin("bnb", sdk.OneInt()),
					ClaimEnd:      now,
					ClaimTimeLock: 10,
					Denom:         "",
				},
			},
			false,
		},
		{
			"duplicate reward period",
			RewardPeriods{
				NewRewardPeriod("bnb", now, now.Add(time.Hour), sdk.NewCoin("bnb", sdk.OneInt()), now, 10),
				NewRewardPeriod("bnb", now, now.Add(time.Hour), sdk.NewCoin("bnb", sdk.OneInt()), now, 10),
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.rewardPeriods.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}

func TestClaimPeriodsValidate(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		msg          string
		claimPeriods ClaimPeriods
		expPass      bool
	}{
		{
			"valid",
			ClaimPeriods{
				NewClaimPeriod("bnb", 10, now, 100),
			},
			true,
		},
		{
			"invalid ID",
			ClaimPeriods{
				{ID: 0},
			},
			false,
		},
		{
			"zero end time",
			ClaimPeriods{
				{ID: 10, End: time.Time{}},
			},
			false,
		},
		{
			"zero time lock",
			ClaimPeriods{
				{ID: 10, End: now, TimeLock: 0},
			},
			false,
		},
		{
			"start time > end time",
			ClaimPeriods{
				{ID: 10, End: now, TimeLock: 0},
			},
			false,
		},
		{
			"invalid denom",
			ClaimPeriods{
				{ID: 10, End: now, TimeLock: 100, Denom: ""},
			},
			false,
		},
		{
			"duplicate reward period",
			ClaimPeriods{
				NewClaimPeriod("bnb", 10, now, 100),
				NewClaimPeriod("bnb", 10, now, 100),
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.claimPeriods.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}

func TestClaimsValidate(t *testing.T) {
	mockPrivKey := tmtypes.NewMockPV()
	pubkey, err := mockPrivKey.GetPubKey()
	require.NoError(t, err)
	owner := sdk.AccAddress(pubkey.Address())

	testCases := []struct {
		msg     string
		claims  Claims
		expPass bool
	}{
		{
			"valid",
			Claims{
				NewClaim(owner, sdk.NewCoin("bnb", sdk.OneInt()), "bnb", 10),
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
			"zero claim period id",
			Claims{
				{
					Owner:         owner,
					Reward:        sdk.NewCoin("bnb", sdk.OneInt()),
					ClaimPeriodID: 0,
				},
			},
			false,
		},
		{
			"invalid denom",
			Claims{
				{
					Owner:         owner,
					Reward:        sdk.NewCoin("bnb", sdk.OneInt()),
					ClaimPeriodID: 10,
					Denom:         "",
				},
			},
			false,
		},
		{
			"duplicate reward period",
			Claims{
				NewClaim(owner, sdk.NewCoin("bnb", sdk.OneInt()), "bnb", 10),
				NewClaim(owner, sdk.NewCoin("bnb", sdk.OneInt()), "bnb", 10),
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
