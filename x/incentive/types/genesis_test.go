package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesisClaimPeriodIDsValidate(t *testing.T) {
	testCases := []struct {
		msg                   string
		genesisClaimPeriodIDs GenesisClaimPeriodIDs
		expPass               bool
	}{
		{
			"valid",
			GenesisClaimPeriodIDs{
				{Denom: "bnb", ID: 1},
			},
			true,
		},
		{
			"invalid denom",
			GenesisClaimPeriodIDs{
				{Denom: "", ID: 1},
			},
			false,
		},
		{
			"invalid ID",
			GenesisClaimPeriodIDs{
				{Denom: "bnb", ID: 0},
			},
			false,
		},
		{
			"duplicate",
			GenesisClaimPeriodIDs{
				{Denom: "bnb", ID: 1},
				{Denom: "bnb", ID: 1},
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.genesisClaimPeriodIDs.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}

func TestGenesisStateValidate(t *testing.T) {
	now := time.Now()
	_ = sdk.AccAddress(tmtypes.NewMockPV().GetPubKey().Address())

	rewards := Rewards{
		NewReward(
			true, "bnb", sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
			time.Hour*24*7, time.Hour*8766, time.Hour*24*14,
		),
	}

	rewardPeriods := RewardPeriods{}
	claimPeriods := ClaimPeriods{}
	claims := Claims{}
	gcps := GenesisClaimPeriodIDs{
		GenesisClaimPeriodID{
			Denom: "bnb",
			ID:    1,
		},
	}

	testCases := []struct {
		msg          string
		genesisState GenesisState
		expPass      bool
	}{
		{
			msg:          "default",
			genesisState: DefaultGenesisState(),
			expPass:      true,
		},
		{
			msg: "valid genesis",
			genesisState: NewGenesisState(
				NewParams(true, rewards),
				now, rewardPeriods, claimPeriods, claims, gcps,
			),
			expPass: true,
		},
		{
			msg: "invalid Params",
			genesisState: GenesisState{
				Params: Params{
					Active: true,
					Rewards: Rewards{
						Reward{},
					},
				},
			},
			expPass: false,
		},
		{
			msg: "zero PreviousBlockTime",
			genesisState: GenesisState{
				PreviousBlockTime: time.Time{},
			},
			expPass: false,
		},
		{
			msg: "invalid RewardsPeriod",
			genesisState: GenesisState{
				PreviousBlockTime: now,
				RewardPeriods: RewardPeriods{
					{Start: time.Time{}},
				},
			},
			expPass: false,
		},
		{
			msg: "invalid ClaimPeriods",
			genesisState: GenesisState{
				PreviousBlockTime: now,
				ClaimPeriods: ClaimPeriods{
					{ID: 0},
				},
			},
			expPass: false,
		},
		{
			msg: "invalid Claims",
			genesisState: GenesisState{
				PreviousBlockTime: now,
				Claims: Claims{
					{ClaimPeriodID: 0},
				},
			},
			expPass: false,
		},
		{
			msg: "invalid NextClaimPeriodIds",
			genesisState: GenesisState{
				PreviousBlockTime: now,
				NextClaimPeriodIDs: GenesisClaimPeriodIDs{
					{ID: 0},
				},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		err := tc.genesisState.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.msg)
		} else {
			require.Error(t, err, tc.msg)
		}
	}
}
