package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	coinsSingle = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)
	coinsZero   = sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt())
	addrs       = []sdk.AccAddress{
		sdk.AccAddress("test1"),
		sdk.AccAddress("test2"),
	}
)

func TestMsgCreateCDP(t *testing.T) {
	tests := []struct {
		description    string
		sender         sdk.AccAddress
		collateral     sdk.Coin
		principal      sdk.Coin
		collateralType string
		expectPass     bool
	}{
		{"create cdp", addrs[0], coinsSingle, coinsSingle, "type-a", true},
		{"create cdp no collateral", addrs[0], coinsZero, coinsSingle, "type-a", false},
		{"create cdp no debt", addrs[0], coinsSingle, coinsZero, "type-a", false},
		{"create cdp empty owner", sdk.AccAddress{}, coinsSingle, coinsSingle, "type-a", false},
		{"create cdp empty type", addrs[0], coinsSingle, coinsSingle, "", false},
	}

	for _, tc := range tests {
		msg := NewMsgCreateCDP(
			tc.sender,
			tc.collateral,
			tc.principal,
			tc.collateralType,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", tc.description)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", tc.description)
		}
	}
}

func TestMsgDeposit(t *testing.T) {
	tests := []struct {
		description    string
		sender         sdk.AccAddress
		depositor      sdk.AccAddress
		collateral     sdk.Coin
		collateralType string
		expectPass     bool
	}{
		{"deposit", addrs[0], addrs[1], coinsSingle, "type-a", true},
		{"deposit same owner", addrs[0], addrs[0], coinsSingle, "type-a", true},
		{"deposit no collateral", addrs[0], addrs[1], coinsZero, "type-a", false},
		{"deposit empty owner", sdk.AccAddress{}, addrs[1], coinsSingle, "type-a", false},
		{"deposit empty depositor", addrs[0], sdk.AccAddress{}, coinsSingle, "type-a", false},
		{"deposit empty type", addrs[0], addrs[0], coinsSingle, "", false},
	}

	for _, tc := range tests {
		msg := NewMsgDeposit(
			tc.sender,
			tc.depositor,
			tc.collateral,
			tc.collateralType,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", tc.description)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", tc.description)
		}
	}
}

func TestMsgWithdraw(t *testing.T) {
	tests := []struct {
		description    string
		sender         sdk.AccAddress
		depositor      sdk.AccAddress
		collateral     sdk.Coin
		collateralType string
		expectPass     bool
	}{
		{"withdraw", addrs[0], addrs[1], coinsSingle, "type-a", true},
		{"withdraw", addrs[0], addrs[0], coinsSingle, "type-a", true},
		{"withdraw no collateral", addrs[0], addrs[1], coinsZero, "type-a", false},
		{"withdraw empty owner", sdk.AccAddress{}, addrs[1], coinsSingle, "type-a", false},
		{"withdraw empty depositor", addrs[0], sdk.AccAddress{}, coinsSingle, "type-a", false},
		{"withdraw empty type", addrs[0], addrs[0], coinsSingle, "", false},
	}

	for _, tc := range tests {
		msg := NewMsgWithdraw(
			tc.sender,
			tc.depositor,
			tc.collateral,
			tc.collateralType,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", tc.description)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", tc.description)
		}
	}
}

func TestMsgDrawDebt(t *testing.T) {
	tests := []struct {
		description    string
		sender         sdk.AccAddress
		collateralType string
		principal      sdk.Coin
		expectPass     bool
	}{
		{"draw debt", addrs[0], sdk.DefaultBondDenom, coinsSingle, true},
		{"draw debt no debt", addrs[0], sdk.DefaultBondDenom, coinsZero, false},
		{"draw debt empty owner", sdk.AccAddress{}, sdk.DefaultBondDenom, coinsSingle, false},
		{"draw debt empty denom", sdk.AccAddress{}, "", coinsSingle, false},
	}

	for _, tc := range tests {
		msg := NewMsgDrawDebt(
			tc.sender,
			tc.collateralType,
			tc.principal,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", tc.description)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", tc.description)
		}
	}
}

func TestMsgRepayDebt(t *testing.T) {
	tests := []struct {
		description string
		sender      sdk.AccAddress
		denom       string
		payment     sdk.Coin
		expectPass  bool
	}{
		{"repay debt", addrs[0], sdk.DefaultBondDenom, coinsSingle, true},
		{"repay debt no payment", addrs[0], sdk.DefaultBondDenom, coinsZero, false},
		{"repay debt empty owner", sdk.AccAddress{}, sdk.DefaultBondDenom, coinsSingle, false},
		{"repay debt empty denom", sdk.AccAddress{}, "", coinsSingle, false},
	}

	for _, tc := range tests {
		msg := NewMsgRepayDebt(
			tc.sender,
			tc.denom,
			tc.payment,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", tc.description)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", tc.description)
		}
	}
}
