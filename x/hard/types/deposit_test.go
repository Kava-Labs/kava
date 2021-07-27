package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/hard/types"
)

func TestDeposit_NormalizedDeposit(t *testing.T) {
	testCases := []struct {
		name      string
		deposit   types.Deposit
		expect    sdk.DecCoins
		expectErr string
	}{
		{
			name: "multiple denoms are calculated correctly",
			deposit: types.Deposit{
				Amount: sdk.NewCoins(
					sdk.NewInt64Coin("bnb", 100e8),
					sdk.NewInt64Coin("xrpb", 1e8),
				),
				Index: types.SupplyInterestFactors{
					{
						Denom: "xrpb",
						Value: sdk.MustNewDecFromStr("1.25"),
					},
					{
						Denom: "bnb",
						Value: sdk.MustNewDecFromStr("2.0"),
					},
				},
			},
			expect: sdk.NewDecCoins(
				sdk.NewInt64DecCoin("bnb", 50e8),
				sdk.NewInt64DecCoin("xrpb", 8e7),
			),
		},
		{
			name: "empty deposit amount returns empty dec coins",
			deposit: types.Deposit{
				Amount: sdk.Coins{},
				Index:  types.SupplyInterestFactors{},
			},
			expect: sdk.DecCoins{},
		},
		{
			name: "nil deposit amount returns empty dec coins",
			deposit: types.Deposit{
				Amount: nil,
				Index:  types.SupplyInterestFactors{},
			},
			expect: sdk.DecCoins{},
		},
		{
			name: "missing indexes return error",
			deposit: types.Deposit{
				Amount: sdk.NewCoins(
					sdk.NewInt64Coin("bnb", 100e8),
				),
				Index: types.SupplyInterestFactors{
					{
						Denom: "xrpb",
						Value: sdk.MustNewDecFromStr("1.25"),
					},
				},
			},
			expectErr: "missing interest factor",
		},
		{
			name: "invalid indexes return error",
			deposit: types.Deposit{
				Amount: sdk.NewCoins(
					sdk.NewInt64Coin("bnb", 100e8),
				),
				Index: types.SupplyInterestFactors{
					{
						Denom: "bnb",
						Value: sdk.MustNewDecFromStr("0.999999999999999999"),
					},
				},
			},
			expectErr: "< 1",
		},
		{
			name: "zero indexes return error rather than panicking",
			deposit: types.Deposit{
				Amount: sdk.NewCoins(
					sdk.NewInt64Coin("bnb", 100e8),
				),
				Index: types.SupplyInterestFactors{
					{
						Denom: "bnb",
						Value: sdk.MustNewDecFromStr("0"),
					},
				},
			},
			expectErr: "< 1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nb, err := tc.deposit.NormalizedDeposit()

			require.Equal(t, tc.expect, nb)

			if len(tc.expectErr) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
			}
		})
	}

}
