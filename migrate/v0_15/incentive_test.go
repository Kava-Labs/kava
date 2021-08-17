package v0_15

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

func TestMigrateDelegatorRewardIndexes(t *testing.T) {
	type expect struct {
		err     bool
		indexes v0_15incentive.MultiRewardIndexes
	}
	testCases := []struct {
		name     string
		oldRIs   v0_14incentive.RewardIndexes
		expected expect
	}{
		{
			name: "single index is migrated correctly",
			oldRIs: v0_14incentive.RewardIndexes{{
				CollateralType: "ukava",
				RewardFactor:   sdk.MustNewDecFromStr("0.1"),
			}},
			expected: expect{
				indexes: v0_15incentive.MultiRewardIndexes{{
					CollateralType: "ukava",
					RewardIndexes: v0_15incentive.RewardIndexes{{
						CollateralType: "hard",
						RewardFactor:   sdk.MustNewDecFromStr("0.1"),
					}},
				}},
			},
		},
		{
			name:   "empty index is migrated correctly",
			oldRIs: v0_14incentive.RewardIndexes{},
			expected: expect{
				indexes: v0_15incentive.MultiRewardIndexes{{
					CollateralType: "ukava",
					RewardIndexes:  v0_15incentive.RewardIndexes{},
				}},
			},
		},
		{
			name: "too many indexes errors",
			oldRIs: v0_14incentive.RewardIndexes{
				{
					CollateralType: "ukava",
					RewardFactor:   sdk.MustNewDecFromStr("0.1"),
				},
				{
					CollateralType: "btcb",
					RewardFactor:   sdk.MustNewDecFromStr("0.2"),
				},
			},
			expected: expect{
				err: true,
			},
		},
		{
			name: "incorrect rewarded denom errors",
			oldRIs: v0_14incentive.RewardIndexes{{
				CollateralType: "btcb",
				RewardFactor:   sdk.MustNewDecFromStr("0.1"),
			}},
			expected: expect{
				err: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualIndexes, err := migrateDelegatorRewardIndexes(tc.oldRIs)
			if tc.expected.err {
				require.Error(t, err)
			} else {
				require.Equal(t, tc.expected.indexes, actualIndexes)
			}
		})
	}
}
