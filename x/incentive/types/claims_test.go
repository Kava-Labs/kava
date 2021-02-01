package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
