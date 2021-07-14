package types_test

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/incentive/types"
)

func TestMsgClaimVVesting_Validate(t *testing.T) {
	validAddress := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1")))

	type expectedErr struct {
		wraps error
		pass  bool
	}
	type msgArgs struct {
		sender         sdk.AccAddress
		receiver       sdk.AccAddress
		multiplierName string
		denomsToClaim  []string
	}
	tests := []struct {
		name    string
		msgArgs msgArgs
		expect  expectedErr
	}{
		{
			name: "large multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "large",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "medium multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "medium",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "small multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "small",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "empty denoms to claim is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "small",
				denomsToClaim:  []string{},
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "invalid sender",
			msgArgs: msgArgs{
				sender:         sdk.AccAddress{},
				receiver:       validAddress,
				multiplierName: "medium",
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "invalid receiver",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       sdk.AccAddress{},
				multiplierName: "medium",
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "invalid multiplier",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "huge",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "multiplier with capitalization is invalid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "Large",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "invalid claim denom",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "small",
				denomsToClaim:  []string{"a denom string that is invalid because it is much too long"},
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "too many claim denoms",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "small",
				denomsToClaim:  tooManyClaimDenoms(),
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
	}

	for _, tc := range tests {
		msgs := []sdk.Msg{
			types.NewMsgClaimHardRewardVVesting(
				tc.msgArgs.sender, tc.msgArgs.receiver, tc.msgArgs.multiplierName, tc.msgArgs.denomsToClaim,
			),
			types.NewMsgClaimDelegatorRewardVVesting(
				tc.msgArgs.sender, tc.msgArgs.receiver, tc.msgArgs.multiplierName, tc.msgArgs.denomsToClaim,
			),
			types.NewMsgClaimSwapRewardVVesting(
				tc.msgArgs.sender, tc.msgArgs.receiver, tc.msgArgs.multiplierName, tc.msgArgs.denomsToClaim,
			),
		}
		for _, msg := range msgs {
			t.Run(msg.Type()+" "+tc.name, func(t *testing.T) {

				err := msg.ValidateBasic()
				if tc.expect.pass {
					require.NoError(t, err)
				} else {
					require.Truef(t, errors.Is(err, tc.expect.wraps), "expected error '%s' was not actual '%s'", tc.expect.wraps, err)
				}
			})
		}
	}
}

func TestMsgClaim_Validate(t *testing.T) {
	validAddress := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1")))

	type expectedErr struct {
		wraps error
		pass  bool
	}
	type msgArgs struct {
		sender         sdk.AccAddress
		multiplierName string
		denomsToClaim  []string
	}
	tests := []struct {
		name    string
		msgArgs msgArgs
		expect  expectedErr
	}{
		{
			name: "large multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "large",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "medium multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "medium",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "small multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "small",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "empty denoms to claim is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "small",
				denomsToClaim:  []string{},
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "invalid sender",
			msgArgs: msgArgs{
				sender:         sdk.AccAddress{},
				multiplierName: "medium",
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "invalid multiplier",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "huge",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "multiplier with capitalization is invalid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "Large",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "invalid claim denom",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "small",
				denomsToClaim:  []string{"a denom string that is invalid because it is much too long"},
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "too many claim denoms",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "small",
				denomsToClaim:  tooManyClaimDenoms(),
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
	}

	for _, tc := range tests {
		msgs := []sdk.Msg{
			types.NewMsgClaimHardReward(
				tc.msgArgs.sender, tc.msgArgs.multiplierName, tc.msgArgs.denomsToClaim,
			),
			types.NewMsgClaimDelegatorReward(
				tc.msgArgs.sender, tc.msgArgs.multiplierName, tc.msgArgs.denomsToClaim,
			),
			types.NewMsgClaimSwapReward(
				tc.msgArgs.sender, tc.msgArgs.multiplierName, tc.msgArgs.denomsToClaim,
			),
		}
		for _, msg := range msgs {
			t.Run(msg.Type()+" "+tc.name, func(t *testing.T) {

				err := msg.ValidateBasic()
				if tc.expect.pass {
					require.NoError(t, err)
				} else {
					require.Truef(t, errors.Is(err, tc.expect.wraps), "expected error '%s' was not actual '%s'", tc.expect.wraps, err)
				}
			})
		}
	}
}

func TestMsgClaimUSDXMintingRewardVVesting_Validate(t *testing.T) {
	validAddress := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1")))

	type expectedErr struct {
		wraps error
		pass  bool
	}
	type msgArgs struct {
		sender         sdk.AccAddress
		receiver       sdk.AccAddress
		multiplierName string
	}
	tests := []struct {
		name    string
		msgArgs msgArgs
		expect  expectedErr
	}{
		{
			name: "large multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "large",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "medium multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "medium",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "small multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "small",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "invalid sender",
			msgArgs: msgArgs{
				sender:         sdk.AccAddress{},
				receiver:       validAddress,
				multiplierName: "medium",
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "invalid receiver",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       sdk.AccAddress{},
				multiplierName: "medium",
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "invalid multiplier",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "huge",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "multiplier with capitalization is invalid",
			msgArgs: msgArgs{
				sender:         validAddress,
				receiver:       validAddress,
				multiplierName: "Large",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			msg := types.NewMsgClaimUSDXMintingRewardVVesting(tc.msgArgs.sender, tc.msgArgs.receiver, tc.msgArgs.multiplierName)

			err := msg.ValidateBasic()
			if tc.expect.pass {
				require.NoError(t, err)
			} else {
				require.Truef(t, errors.Is(err, tc.expect.wraps), "expected error '%s' was not actual '%s'", tc.expect.wraps, err)
			}
		})
	}
}

func TestMsgClaimUSDXMintingReward_Validate(t *testing.T) {
	validAddress := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1")))

	type expectedErr struct {
		wraps error
		pass  bool
	}
	type msgArgs struct {
		sender         sdk.AccAddress
		multiplierName string
		denomsToClaim  []string
	}
	tests := []struct {
		name    string
		msgArgs msgArgs
		expect  expectedErr
	}{
		{
			name: "large multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "large",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "medium multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "medium",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "small multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "small",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "invalid sender",
			msgArgs: msgArgs{
				sender:         sdk.AccAddress{},
				multiplierName: "medium",
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "invalid multiplier",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "huge",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "multiplier with capitalization is invalid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "Large",
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgClaimUSDXMintingReward(tc.msgArgs.sender, tc.msgArgs.multiplierName)

			err := msg.ValidateBasic()
			if tc.expect.pass {
				require.NoError(t, err)
			} else {
				require.Truef(t, errors.Is(err, tc.expect.wraps), "expected error '%s' was not actual '%s'", tc.expect.wraps, err)
			}
		})
	}
}

func tooManyClaimDenoms() []string {
	claimDenoms := make([]string, types.MaxDenomsToClaim+1)
	for i := range claimDenoms {
		claimDenoms[i] = fmt.Sprintf("denom%d", i)
	}
	return claimDenoms
}
