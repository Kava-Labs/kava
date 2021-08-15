package v0_15

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/incentive"
)

// d parses a string into an sdk.Dec type.
// It is an alias for sdk.MustNewDecFromStr.
var d = sdk.MustNewDecFromStr

func TestSyncUSDXMintingClaims(t *testing.T) {
	claims := incentive.USDXMintingClaims{
		incentive.NewUSDXMintingClaim(
			sdk.AccAddress("address1"),
			sdk.NewInt64Coin(incentive.USDXMintingRewardDenom, 1e9),
			incentive.RewardIndexes{
				{CollateralType: "bnb-a", RewardFactor: d("0.1")},
			},
		),
		incentive.NewUSDXMintingClaim(
			sdk.AccAddress("address2"),
			sdk.NewInt64Coin(incentive.USDXMintingRewardDenom, 0),
			incentive.RewardIndexes{
				{CollateralType: "bnb-a", RewardFactor: d("0")},
				{CollateralType: "xrpb-a", RewardFactor: d("0.2")},
			},
		),
		incentive.NewUSDXMintingClaim(
			sdk.AccAddress("address3"),
			sdk.NewInt64Coin(incentive.USDXMintingRewardDenom, 0),
			incentive.RewardIndexes{},
		),
	}

	indexes := incentive.MultiRewardIndexes{
		incentive.NewMultiRewardIndex(
			"bnb-a",
			incentive.RewardIndexes{
				{CollateralType: incentive.USDXMintingRewardDenom, RewardFactor: d("0.5")},
			},
		),
		incentive.NewMultiRewardIndex(
			"xrpb-a",
			incentive.RewardIndexes{
				{CollateralType: incentive.USDXMintingRewardDenom, RewardFactor: d("0.8")},
			},
		),
	}

	expectedIndexes := incentive.RewardIndexes{
		{CollateralType: "bnb-a", RewardFactor: d("0.5")},
		{CollateralType: "xrpb-a", RewardFactor: d("0.8")},
	}

	syncedClaims := syncUSDXMintingClaims(claims, indexes)

	for i, claim := range syncedClaims {
		// check fields are unchanged
		require.Equal(t, claim.Owner, claims[i].Owner)
		require.Equal(t, claim.Reward, claims[i].Reward)
		// except for indexes which have been overwritten
		require.Equal(t, expectedIndexes, claim.RewardIndexes)
	}
}
