package v0_15

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	v0_15hard "github.com/kava-labs/kava/x/hard/types"
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

func TestAddMissingHardClaims_Basic(t *testing.T) {
	claims := v0_15incentive.HardLiquidityProviderClaims{
		v0_15incentive.NewHardLiquidityProviderClaim(
			sdk.AccAddress("address1"),
			sdk.NewCoins(sdk.NewInt64Coin("hard", 1e9)),
			v0_15incentive.MultiRewardIndexes{},
			v0_15incentive.MultiRewardIndexes{},
		),
	}

	deposits := v0_15hard.Deposits{
		v0_15hard.NewDeposit(
			sdk.AccAddress("address1"),
			nil,
			nil,
		),
		v0_15hard.NewDeposit(
			sdk.AccAddress("address2"),
			nil,
			nil,
		),
	}
	borrows := v0_15hard.Borrows{
		v0_15hard.NewBorrow(
			sdk.AccAddress("address1"),
			nil, // only need the owner address for this test
			nil,
		),
		v0_15hard.NewBorrow(
			sdk.AccAddress("address3"),
			nil, // only need the owner address for this test
			nil,
		),
	}

	actualClaims := addMissingHardClaims(claims, deposits, borrows, nil, nil)

	expectedClaims := v0_15incentive.HardLiquidityProviderClaims{
		v0_15incentive.NewHardLiquidityProviderClaim(
			sdk.AccAddress("address1"),
			sdk.NewCoins(sdk.NewInt64Coin("hard", 1e9)),
			v0_15incentive.MultiRewardIndexes{},
			v0_15incentive.MultiRewardIndexes{},
		),
		v0_15incentive.NewHardLiquidityProviderClaim(
			sdk.AccAddress("address3"),
			sdk.NewCoins(),
			v0_15incentive.MultiRewardIndexes{},
			v0_15incentive.MultiRewardIndexes{},
		),
		v0_15incentive.NewHardLiquidityProviderClaim(
			sdk.AccAddress("address2"),
			sdk.NewCoins(),
			v0_15incentive.MultiRewardIndexes{},
			v0_15incentive.MultiRewardIndexes{},
		),
	}

	require.Equal(t, expectedClaims, actualClaims)
}

func TestAlignClaimIndexes_Basic(t *testing.T) {
	claims := v0_15incentive.HardLiquidityProviderClaims{
		v0_15incentive.NewHardLiquidityProviderClaim(
			sdk.AccAddress("address1"),
			sdk.NewCoins(sdk.NewInt64Coin("hard", 1e9)),
			v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "ukava",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "hard",
					RewardIndexes:  nil,
				},
			},
			v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "busd",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "xrpb",
					RewardIndexes:  nil,
				},
			},
		),
	}

	deposits := v0_15hard.Deposits{
		v0_15hard.NewDeposit(
			sdk.AccAddress("address1"),
			sdk.NewCoins(sdk.NewInt64Coin("ukava", 1)),
			nil,
		),
		v0_15hard.NewDeposit(
			sdk.AccAddress("address2"),
			nil,
			nil,
		),
	}
	borrows := v0_15hard.Borrows{
		v0_15hard.NewBorrow(
			sdk.AccAddress("address1"),
			sdk.NewCoins(sdk.NewInt64Coin("xrpb", 1)),
			nil,
		),
		v0_15hard.NewBorrow(
			sdk.AccAddress("address3"),
			nil,
			nil,
		),
	}

	actualClaims := alignClaimIndexes(claims, deposits, borrows, nil, nil)

	expectedClaims := v0_15incentive.HardLiquidityProviderClaims{
		v0_15incentive.NewHardLiquidityProviderClaim(
			sdk.AccAddress("address1"),
			sdk.NewCoins(sdk.NewInt64Coin("hard", 1e9)),
			v0_15incentive.MultiRewardIndexes{{
				CollateralType: "ukava",
				RewardIndexes:  nil,
			}},
			v0_15incentive.MultiRewardIndexes{{
				CollateralType: "xrpb",
				RewardIndexes:  nil,
			}},
		),
	}

	require.Equal(t, expectedClaims, actualClaims)
}

func TestAlignIndexes(t *testing.T) {
	globalRI := v0_15incentive.RewardIndexes{{
		CollateralType: "hard",
		RewardFactor:   sdk.OneDec(),
	}}
	globalIndexes := v0_15incentive.MultiRewardIndexes{
		{
			CollateralType: "ukava",
			RewardIndexes:  globalRI,
		},
		{
			CollateralType: "hard",
			RewardIndexes:  globalRI,
		},
	}

	testCases := []struct {
		name     string
		indexes  v0_15incentive.MultiRewardIndexes
		coins    sdk.Coins
		expected v0_15incentive.MultiRewardIndexes
	}{
		{
			name: "indexes matching coins are unchanged",
			indexes: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "ukava",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "hard", // not in alphabetic order
					RewardIndexes:  nil,
				},
			},
			coins: sdk.NewCoins(sdk.NewInt64Coin("hard", 1), sdk.NewInt64Coin("ukava", 1)),
			expected: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "ukava",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "hard", // order is preserved
					RewardIndexes:  nil,
				},
			},
		},
		{
			name: "missing indexes are added from global values",
			indexes: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "ukava",
					RewardIndexes:  nil,
				},
			},
			coins: sdk.NewCoins(sdk.NewInt64Coin("hard", 1), sdk.NewInt64Coin("ukava", 1)),
			expected: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "ukava",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "hard",
					RewardIndexes:  globalRI,
				},
			},
		},
		{
			name: "extra indexes are removed",
			indexes: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "ukava",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "hard",
					RewardIndexes:  nil,
				},
			},
			coins: sdk.NewCoins(sdk.NewInt64Coin("hard", 1)),
			expected: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "hard",
					RewardIndexes:  nil,
				},
			},
		},
		{
			name: "missing indexes are added even when not in global values",
			indexes: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "hard",
					RewardIndexes:  nil,
				},
			},
			coins: sdk.NewCoins(sdk.NewInt64Coin("hard", 1), sdk.NewInt64Coin("btcb", 1)),
			expected: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "hard",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "btcb",
					RewardIndexes:  v0_15incentive.RewardIndexes{},
				},
			},
		},
		{
			name: "empty coins results in empty indexes",
			indexes: v0_15incentive.MultiRewardIndexes{
				{
					CollateralType: "ukava",
					RewardIndexes:  nil,
				},
				{
					CollateralType: "hard",
					RewardIndexes:  nil,
				},
			},
			coins:    sdk.NewCoins(),
			expected: v0_15incentive.MultiRewardIndexes{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, alignIndexes(tc.indexes, tc.coins, globalIndexes))
		})
	}
}
