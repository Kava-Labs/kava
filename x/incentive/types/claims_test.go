package types

import (
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
}
