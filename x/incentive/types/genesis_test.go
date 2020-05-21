package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
			msg: "invalid NextClaimPeriodIDs",
			genesisState: GenesisState{
				PreviousBlockTime: now,
				NextClaimPeriodIDs: GenesisClaimPeriodIDs{
					GenesisClaimPeriodID{
						Denom: "",
						ID:    1,
					},
				},
			},
			expPass: false,
		},
		{
			msg: "invalid NextClaimPeriodIDs",
			genesisState: GenesisState{
				PreviousBlockTime: now,
				NextClaimPeriodIDs: GenesisClaimPeriodIDs{
					GenesisClaimPeriodID{
						Denom: "bnb",
						ID:    0,
					},
				},
			},
			expPass: false,
		},
		{
			msg: "dup NextClaimPeriodIDs",
			genesisState: GenesisState{
				PreviousBlockTime: now,
				NextClaimPeriodIDs: GenesisClaimPeriodIDs{
					GenesisClaimPeriodID{
						Denom: "bnb",
						ID:    1,
					},
					GenesisClaimPeriodID{
						Denom: "bnb",
						ID:    1,
					},
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
