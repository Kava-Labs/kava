package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/community/types"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name        string
		params      types.Params
		expectedErr string
	}{
		{
			name: "valid parms",
			params: types.Params{
				UpgradeTimeDisableInflation: time.Time{},
				RewardsPerSecond: sdk.NewCoin(
					"ukava",
					sdkmath.NewInt(1000),
				),
			},
			expectedErr: "",
		},
		{
			name: "invalid rewards per second",
			params: types.Params{
				UpgradeTimeDisableInflation: time.Time{},
				RewardsPerSecond: sdk.Coin{
					Denom:  "ukava",
					Amount: sdkmath.NewInt(-1),
				},
			},
			expectedErr: "invalid rewards per second: negative coin amount: -1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
