package v0_15

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v0_15staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

func TestAddMissingDelegatorClaims(t *testing.T) {
	claims := v0_15incentive.DelegatorClaims{
		v0_15incentive.NewDelegatorClaim(
			sdk.AccAddress("address1"),
			nil,
			nil,
		),
		v0_15incentive.NewDelegatorClaim(
			sdk.AccAddress("address3"),
			nil,
			nil,
		),
	}

	delegations := v0_15staking.Delegations{
		{
			DelegatorAddress: sdk.AccAddress("address2"),
		},
		{
			DelegatorAddress: sdk.AccAddress("address1"),
		},
		{
			DelegatorAddress: sdk.AccAddress("address3"),
		},
		{
			DelegatorAddress: sdk.AccAddress("address3"), // there can be multiple delegations per delegator
		},
		{
			DelegatorAddress: sdk.AccAddress("address4"),
		},
		{
			DelegatorAddress: sdk.AccAddress("address4"),
		},
	}

	globalIndexes := v0_15incentive.MultiRewardIndexes{{
		CollateralType: "ukava",
		RewardIndexes: v0_15incentive.RewardIndexes{
			{
				CollateralType: "hard",
				RewardFactor:   sdk.MustNewDecFromStr("0.1"),
			},
			{
				CollateralType: "swp",
				RewardFactor:   sdk.MustNewDecFromStr("0.2"),
			},
		},
	}}

	expectedClaims := v0_15incentive.DelegatorClaims{
		v0_15incentive.NewDelegatorClaim(
			sdk.AccAddress("address1"),
			nil,
			nil,
		),
		v0_15incentive.NewDelegatorClaim(
			sdk.AccAddress("address3"),
			nil,
			nil,
		),
		v0_15incentive.NewDelegatorClaim(
			sdk.AccAddress("address2"),
			sdk.NewCoins(),
			globalIndexes,
		),
		v0_15incentive.NewDelegatorClaim(
			sdk.AccAddress("address4"),
			sdk.NewCoins(),
			globalIndexes,
		),
	}

	newClaims := addMissingDelegatorClaims(claims, delegations, globalIndexes)
	require.Equal(t, expectedClaims, newClaims)
}
