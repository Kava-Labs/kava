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

func TestMsgClaim_Validate(t *testing.T) {
	validAddress := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))).String()

	type expectedErr struct {
		wraps error
		pass  bool
	}
	type msgArgs struct {
		sender        string
		denomsToClaim types.Selections
	}
	tests := []struct {
		name    string
		msgArgs msgArgs
		expect  expectedErr
	}{

		{
			name: "normal multiplier is valid",
			msgArgs: msgArgs{
				sender: validAddress,
				denomsToClaim: types.Selections{
					{
						Denom:          "hard",
						MultiplierName: "large",
					},
				},
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "empty multiplier name is invalid",
			msgArgs: msgArgs{
				sender: validAddress,
				denomsToClaim: types.Selections{
					{
						Denom:          "hard",
						MultiplierName: "",
					},
				},
			},
			expect: expectedErr{
				wraps: types.ErrInvalidMultiplier,
			},
		},
		{
			name: "empty denoms to claim is not valid",
			msgArgs: msgArgs{
				sender:        validAddress,
				denomsToClaim: types.Selections{},
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "nil denoms to claim is not valid",
			msgArgs: msgArgs{
				sender:        validAddress,
				denomsToClaim: nil,
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "invalid sender",
			msgArgs: msgArgs{
				sender: "",
				denomsToClaim: types.Selections{
					{
						Denom:          "hard",
						MultiplierName: "medium",
					},
				},
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "invalid claim denom",
			msgArgs: msgArgs{
				sender: validAddress,
				denomsToClaim: types.Selections{
					{
						Denom:          "a denom string that is invalid because it is much too long",
						MultiplierName: "medium",
					},
				},
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "too many claim denoms",
			msgArgs: msgArgs{
				sender:        validAddress,
				denomsToClaim: tooManySelections(),
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
		{
			name: "duplicated claim denoms",
			msgArgs: msgArgs{
				sender: validAddress,
				denomsToClaim: types.Selections{
					{
						Denom:          "hard",
						MultiplierName: "medium",
					},
					{
						Denom:          "hard",
						MultiplierName: "large",
					},
				},
			},
			expect: expectedErr{
				wraps: types.ErrInvalidClaimDenoms,
			},
		},
	}

	for _, tc := range tests {
		msgClaimHardReward := types.NewMsgClaimHardReward(tc.msgArgs.sender, tc.msgArgs.denomsToClaim)
		msgClaimDelegatorReward := types.NewMsgClaimDelegatorReward(tc.msgArgs.sender, tc.msgArgs.denomsToClaim)
		msgClaimSwapReward := types.NewMsgClaimSwapReward(tc.msgArgs.sender, tc.msgArgs.denomsToClaim)
		msgs := []sdk.Msg{&msgClaimHardReward, &msgClaimDelegatorReward, &msgClaimSwapReward}
		for _, msg := range msgs {
			t.Run(tc.name, func(t *testing.T) {

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

func TestMsgClaimUSDXMintingReward_Validate(t *testing.T) {
	validAddress := sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))).String()

	type expectedErr struct {
		wraps error
		pass  bool
	}
	type msgArgs struct {
		sender         string
		multiplierName string
	}
	tests := []struct {
		name    string
		msgArgs msgArgs
		expect  expectedErr
	}{
		{
			name: "normal multiplier is valid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "large",
			},
			expect: expectedErr{
				pass: true,
			},
		},
		{
			name: "invalid sender",
			msgArgs: msgArgs{
				sender:         "",
				multiplierName: "medium",
			},
			expect: expectedErr{
				wraps: sdkerrors.ErrInvalidAddress,
			},
		},
		{
			name: "empty multiplier is invalid",
			msgArgs: msgArgs{
				sender:         validAddress,
				multiplierName: "",
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

func tooManySelections() types.Selections {
	selections := make(types.Selections, types.MaxDenomsToClaim+1)
	for i := range selections {
		selections[i] = types.Selection{
			Denom:          fmt.Sprintf("denom%d", i),
			MultiplierName: "large",
		}
	}
	return selections
}
