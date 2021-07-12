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
	tests := []struct {
		name           string
		sender         sdk.AccAddress
		receiver       sdk.AccAddress
		multiplierName string
		expect         expectedErr
	}{
		{
			name:           "large multiplier is valid",
			sender:         validAddress,
			receiver:       validAddress,
			multiplierName: "large",
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name:           "medium multiplier is valid",
			sender:         validAddress,
			receiver:       validAddress,
			multiplierName: "medium",
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name:           "small multiplier is valid",
			sender:         validAddress,
			receiver:       validAddress,
			multiplierName: "small",
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name:           "invalid sender",
			sender:         sdk.AccAddress{},
			receiver:       validAddress,
			multiplierName: "medium",
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name:           "invalid receiver",
			sender:         validAddress,
			receiver:       sdk.AccAddress{},
			multiplierName: "medium",
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name:           "invalid multiplier",
			sender:         validAddress,
			receiver:       validAddress,
			multiplierName: "huge",
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
	}

	for _, tc := range tests {
		msgs := []sdk.Msg{
			types.NewMsgClaimUSDXMintingRewardVVesting(tc.sender, tc.receiver, tc.multiplierName),
			types.NewMsgClaimHardRewardVVesting(tc.sender, tc.receiver, tc.multiplierName),
			types.NewMsgClaimDelegatorRewardVVesting(tc.sender, tc.receiver, tc.multiplierName),
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
	tests := []struct {
		name           string
		sender         sdk.AccAddress
		multiplierName string
		expect         expectedErr
	}{
		{
			name:           "large multiplier is valid",
			sender:         validAddress,
			multiplierName: "large",
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name:           "medium multiplier is valid",
			sender:         validAddress,
			multiplierName: "medium",
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name:           "small multiplier is valid",
			sender:         validAddress,
			multiplierName: "small",
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name:           "invalid sender",
			sender:         sdk.AccAddress{},
			multiplierName: "medium",
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name:           "invalid multiplier",
			sender:         validAddress,
			multiplierName: "huge",
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
	}

	for _, tc := range tests {
		msgs := []sdk.Msg{
			types.NewMsgClaimUSDXMintingReward(tc.sender, tc.multiplierName),
			types.NewMsgClaimHardReward(tc.sender, tc.multiplierName),
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

func TestMsgClaimDelegatorReward_Validate(t *testing.T) {
	validAddress := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1")))

	tooManyClaimDenoms := make([]string, types.MaxDenomsToClaim+1)
	for i := range tooManyClaimDenoms {
		tooManyClaimDenoms[i] = fmt.Sprintf("denom%d", i)
	}

	type expectedErr struct {
		wraps error
		pass  bool
	}
	tests := []struct {
		name     string
		msg      types.MsgClaimDelegatorReward
		expected expectedErr
	}{
		{
			name: "valid msg passed validate",
			msg: types.MsgClaimDelegatorReward{
				Sender:         validAddress,
				MultiplierName: "large",
				DenomsToClaim:  nil,
			},
			expected: expectedErr{
				pass: true,
			},
		},
		{
			name: "invalid multiplier name",
			msg: types.MsgClaimDelegatorReward{
				Sender:         validAddress,
				MultiplierName: "invalid multiplier name",
				DenomsToClaim:  nil,
			},
			expected: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "invalid claim denom",
			msg: types.MsgClaimDelegatorReward{
				Sender:         validAddress,
				MultiplierName: "small",
				DenomsToClaim:  []string{"a denom string that is invalid because it is much too long"},
			},
			expected: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "too many claim denoms",
			msg: types.MsgClaimDelegatorReward{
				Sender:         validAddress,
				MultiplierName: "small",
				DenomsToClaim:  tooManyClaimDenoms,
			},
			expected: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "invalid sender",
			msg: types.MsgClaimDelegatorReward{
				Sender:         nil,
				MultiplierName: "medium",
				DenomsToClaim:  nil,
			},
			expected: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expected.pass {
				require.NoError(t, err)
			} else {
				require.Truef(t, errors.Is(err, tc.expected.wraps), "expected error '%s' was not actual '%s'", tc.expected.wraps, err)
			}
		})
	}
}
