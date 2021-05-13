package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestClaimsValidate(t *testing.T) {
	owner := sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser1")))

	testCases := []struct {
		msg     string
		claims  USDXMintingClaims
		expPass bool
	}{
		{
			"valid",
			USDXMintingClaims{
				NewUSDXMintingClaim(owner, sdk.NewCoin("bnb", sdk.OneInt()), RewardIndexes{NewRewardIndex("bnb-a", sdk.ZeroDec())}),
			},
			true,
		},
		{
			"invalid owner",
			USDXMintingClaims{
				USDXMintingClaim{
					BaseClaim: BaseClaim{
						Owner: nil,
					},
				},
			},
			false,
		},
		{
			"invalid reward",
			USDXMintingClaims{
				{
					BaseClaim: BaseClaim{
						Owner:  owner,
						Reward: sdk.Coin{Denom: "", Amount: sdk.ZeroInt()},
					},
				},
			},
			false,
		},
		{
			"invalid collateral type",
			USDXMintingClaims{
				{
					BaseClaim: BaseClaim{
						Owner:  owner,
						Reward: sdk.NewCoin("bnb", sdk.OneInt()),
					},
					RewardIndexes: []RewardIndex{{"", sdk.ZeroDec()}},
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

func TestRewardIndexes(t *testing.T) {
	t.Run("With", func(t *testing.T) {
		var arbitraryDec = sdk.MustNewDecFromStr("0.1")

		type args struct {
			denom  string
			factor sdk.Dec
		}
		testcases := []struct {
			name          string
			rewardIndexes RewardIndexes
			args          args
			expected      RewardIndexes
		}{
			{
				name: "when index is not present, it's added and original isn't overwritten",
				rewardIndexes: RewardIndexes{
					NewRewardIndex("denom", arbitraryDec),
				},
				args: args{
					denom:  "otherdenom",
					factor: arbitraryDec,
				},
				expected: RewardIndexes{
					NewRewardIndex("denom", arbitraryDec),
					NewRewardIndex("otherdenom", arbitraryDec),
				},
			},
			{
				name: "when index is present, it's updated and original isn't overwritten",
				rewardIndexes: RewardIndexes{
					NewRewardIndex("denom", arbitraryDec),
				},
				args: args{
					denom:  "denom",
					factor: arbitraryDec.MulInt64(2),
				},
				expected: RewardIndexes{
					NewRewardIndex("denom", arbitraryDec.MulInt64(2)),
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				newIndexes := tc.rewardIndexes.With(tc.args.denom, tc.args.factor)

				require.Equal(t, tc.expected, newIndexes)
				require.NotEqual(t, tc.rewardIndexes, newIndexes) // check original slice not modified
			})
		}
	})
	t.Run("Get", func(t *testing.T) {
		var arbitraryDec = sdk.MustNewDecFromStr("0.1")

		type expected struct {
			factor sdk.Dec
			found  bool
		}
		testcases := []struct {
			name          string
			rewardIndexes RewardIndexes
			arg_denom     string
			expected      expected
		}{
			{
				name: "when index is present, it is found and returned",
				rewardIndexes: RewardIndexes{
					NewRewardIndex("denom", arbitraryDec),
				},
				arg_denom: "denom",
				expected: expected{
					factor: arbitraryDec,
					found:  true,
				},
			},
			{
				name: "when index is not present, it is not found",
				rewardIndexes: RewardIndexes{
					NewRewardIndex("denom", arbitraryDec),
				},
				arg_denom: "notpresent",
				expected: expected{
					found: false,
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				factor, found := tc.rewardIndexes.Get(tc.arg_denom)

				require.Equal(t, tc.expected.found, found)
				require.Equal(t, tc.expected.factor, factor)
			})
		}
	})
}

func TestMultiRewardIndexes(t *testing.T) {
	arbitraryRewardIndexes := RewardIndexes{
		{
			CollateralType: "reward",
			RewardFactor:   sdk.MustNewDecFromStr("0.1"),
		},
	}

	t.Run("Get", func(t *testing.T) {
		type expected struct {
			rewardIndexes RewardIndexes
			found         bool
		}
		testcases := []struct {
			name               string
			multiRewardIndexes MultiRewardIndexes
			arg_denom          string
			expected           expected
		}{
			{
				name: "when indexes are present, they are found and returned",
				multiRewardIndexes: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
				arg_denom: "denom",
				expected: expected{
					found:         true,
					rewardIndexes: arbitraryRewardIndexes,
				},
			},
			{
				name: "when indexes are not present, they are not found",
				multiRewardIndexes: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
				arg_denom: "notpresent",
				expected: expected{
					found: false,
				},
			},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				rewardIndexes, found := tc.multiRewardIndexes.Get(tc.arg_denom)

				require.Equal(t, tc.expected.found, found)
				require.Equal(t, tc.expected.rewardIndexes, rewardIndexes)
			})
		}
	})
	t.Run("With", func(t *testing.T) {
		type args struct {
			denom         string
			rewardIndexes RewardIndexes
		}
		testcases := []struct {
			name               string
			multiRewardIndexes MultiRewardIndexes
			args               args
			expected           MultiRewardIndexes
		}{
			{
				name: "when indexes are not present, add them and do not update original",
				multiRewardIndexes: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
				args: args{
					denom:         "otherdenom",
					rewardIndexes: arbitraryRewardIndexes,
				},
				expected: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
					{
						CollateralType: "otherdenom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
			},
			{
				name: "when indexes are present, update them and do not update original",
				multiRewardIndexes: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
				args: args{
					denom:         "denom",
					rewardIndexes: appendUniqueRewardIndex(arbitraryRewardIndexes),
				},
				expected: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  appendUniqueRewardIndex(arbitraryRewardIndexes),
					},
				},
			},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				oldIndexes := tc.multiRewardIndexes.copy()

				newIndexes := tc.multiRewardIndexes.With(tc.args.denom, tc.args.rewardIndexes)

				require.Equal(t, tc.expected, newIndexes)
				require.Equal(t, oldIndexes, tc.multiRewardIndexes)
			})
		}
	})
	t.Run("RemoveRewardIndex", func(t *testing.T) {
		testcases := []struct {
			name               string
			multiRewardIndexes MultiRewardIndexes
			arg_denom          string
			expected           MultiRewardIndexes
		}{
			{
				name: "when indexes are not present, do nothing",
				multiRewardIndexes: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
				arg_denom: "notpresent",
				expected: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
			},
			{
				name: "when indexes are present, remove them and do not update original",
				multiRewardIndexes: MultiRewardIndexes{
					{
						CollateralType: "denom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
					{
						CollateralType: "otherdenom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
				arg_denom: "denom",
				expected: MultiRewardIndexes{
					{
						CollateralType: "otherdenom",
						RewardIndexes:  arbitraryRewardIndexes,
					},
				},
			},
		}
		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				oldIndexes := tc.multiRewardIndexes.copy()

				newIndexes := tc.multiRewardIndexes.RemoveRewardIndex(tc.arg_denom)

				require.Equal(t, tc.expected, newIndexes)
				require.Equal(t, oldIndexes, tc.multiRewardIndexes)
			})
		}
	})
}

// TODO dedupe with copy in keeper
func appendUniqueRewardIndex(indexes RewardIndexes) RewardIndexes {
	const uniqueDenom = "uniquereward"

	for _, mri := range indexes {
		if mri.CollateralType == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique reward index with denom '%s', but denom already existed", uniqueDenom))
		}
	}

	return append(
		indexes,
		NewRewardIndex(uniqueDenom, sdk.MustNewDecFromStr("0.02")),
	)
}
