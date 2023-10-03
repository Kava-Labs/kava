package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/kava-labs/kava/x/community/types"
)

type stakingRewardsStateTestCase struct {
	name                string
	stakingRewardsState types.StakingRewardsState
	expectedErr         string
}

var stakingRewardsStateTestCases = []stakingRewardsStateTestCase{
	{
		name:                "default stakingRewardsState are valid",
		stakingRewardsState: types.DefaultStakingRewardsState(),
		expectedErr:         "",
	},
	{
		name: "valid example state",
		stakingRewardsState: types.StakingRewardsState{
			LastAccumulationTime: time.Now(),
			LastTruncationError:  newDecFromString("0.10000000000000000"),
		},
		expectedErr: "",
	},
	{
		name: "last accumulation time can be zero",
		stakingRewardsState: types.StakingRewardsState{
			LastAccumulationTime: time.Time{},
			LastTruncationError:  sdkmath.LegacyZeroDec(),
		},
		expectedErr: "",
	},
	{
		name: "nil last truncation error is invalid",
		stakingRewardsState: types.StakingRewardsState{
			LastAccumulationTime: time.Now(),
			LastTruncationError:  sdkmath.LegacyDec{},
		},
		expectedErr: "LastTruncationError should not be nil",
	},
	{
		name: "negative last truncation error is invalid",
		stakingRewardsState: types.StakingRewardsState{
			LastAccumulationTime: time.Now(),
			LastTruncationError:  newDecFromString("-0.10000000000000000"),
		},
		expectedErr: "LastTruncationError should not be negative",
	},
	{
		name: "last truncation error equal to 1 is invalid",
		stakingRewardsState: types.StakingRewardsState{
			LastAccumulationTime: time.Now(),
			LastTruncationError:  newDecFromString("1.00000000000000000"),
		},
		expectedErr: "LastTruncationError should not be greater or equal to 1",
	},
	{
		name: "last truncation error greater than 1 is invalid",
		stakingRewardsState: types.StakingRewardsState{
			LastAccumulationTime: time.Now(),
			LastTruncationError:  newDecFromString("1.00000000000000000"),
		},
		expectedErr: "LastTruncationError should not be greater or equal to 1",
	},
	{
		name: "last truncation error can not be set if last accumulation time is zero",
		stakingRewardsState: types.StakingRewardsState{
			LastAccumulationTime: time.Time{},
			LastTruncationError:  newDecFromString("0.10000000000000000"),
		},
		expectedErr: "LastTruncationError should be zero if last accumulation time is zero",
	},
}

func TestStakingRewardsStateValidate(t *testing.T) {
	for _, tc := range stakingRewardsStateTestCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.stakingRewardsState.Validate()

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

// newDecFromString returns a new sdkmath.Int from a string
func newDecFromString(str string) sdkmath.LegacyDec {
	num, err := sdkmath.LegacyNewDecFromStr(str)
	if err != nil {
		panic(err)
	}
	return num
}
