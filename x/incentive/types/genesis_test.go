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
				{CollateralType: "bnb", ID: 1},
			},
			true,
		},
		{
			"invalid collateral type",
			GenesisClaimPeriodIDs{
				{CollateralType: "", ID: 1},
			},
			false,
		},
		{
			"invalid ID",
			GenesisClaimPeriodIDs{
				{CollateralType: "bnb", ID: 0},
			},
			false,
		},
		{
			"duplicate",
			GenesisClaimPeriodIDs{
				{CollateralType: "bnb", ID: 1},
				{CollateralType: "bnb", ID: 1},
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
	mockPrivKey := tmtypes.NewMockPV()
	pubkey, err := mockPrivKey.GetPubKey()
	require.NoError(t, err)
	owner := sdk.AccAddress(pubkey.Address())

	rewards := Rewards{
		NewReward(
			true, "bnb", sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
			time.Hour*24*7, time.Hour*8766, time.Hour*24*14,
		),
	}
	rewardPeriods := RewardPeriods{NewRewardPeriod("bnb", now, now.Add(time.Hour), sdk.NewCoin("bnb", sdk.OneInt()), now, 10)}
	claimPeriods := ClaimPeriods{NewClaimPeriod("bnb", 10, now, 100)}
	claims := Claims{NewClaim(owner, sdk.NewCoin("bnb", sdk.OneInt()), "bnb", 10)}
	gcps := GenesisClaimPeriodIDs{{CollateralType: "bnb", ID: 1}}

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
