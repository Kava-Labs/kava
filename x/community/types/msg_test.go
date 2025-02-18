package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/types"
)

func TestFundCommunityPool_ValidateBasic(t *testing.T) {
	validCoins := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewIntFromUint64(1e6)),
		sdk.NewCoin("some-denom", sdkmath.NewIntFromUint64(1e4)),
	)
	testCases := []struct {
		name       string
		shouldPass bool
		message    types.MsgFundCommunityPool
	}{
		{
			name:       "valid message",
			shouldPass: true,
			message:    types.NewMsgFundCommunityPool(app.RandomAddress(), validCoins),
		},
		{
			name:       "invalid - bad depositor",
			shouldPass: false,
			message: types.MsgFundCommunityPool{
				Depositor: "not-an-address",
				Amount:    validCoins,
			},
		},
		{
			name:       "invalid - empty coins",
			shouldPass: false,
			message: types.MsgFundCommunityPool{
				Depositor: app.RandomAddress().String(),
				Amount:    sdk.NewCoins(),
			},
		},
		{
			name:       "invalid - nil coins",
			shouldPass: false,
			message: types.MsgFundCommunityPool{
				Depositor: app.RandomAddress().String(),
				Amount:    nil,
			},
		},
		{
			name:       "invalid - zero coins",
			shouldPass: false,
			message: types.MsgFundCommunityPool{
				Depositor: app.RandomAddress().String(),
				Amount: sdk.NewCoins(
					sdk.NewCoin("ukava", sdk.ZeroInt()),
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.message.ValidateBasic()
			if tc.shouldPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name        string
		message     types.MsgUpdateParams
		expectedErr error
	}{
		{
			name:        "valid message",
			message:     types.NewMsgUpdateParams(app.RandomAddress(), types.DefaultParams()),
			expectedErr: nil,
		},
		{
			name: "invalid - bad authority",
			message: types.MsgUpdateParams{
				Authority: "not-an-address",
				Params:    types.DefaultParams(),
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid - empty authority",
			message: types.MsgUpdateParams{
				Authority: "",
				Params:    types.DefaultParams(),
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid - invalid params",
			message: types.MsgUpdateParams{
				Authority: app.RandomAddress().String(),
				Params:    types.Params{},
			},
			expectedErr: types.ErrInvalidParams,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.message.ValidateBasic()
			if tc.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.expectedErr)
			}
		})
	}
}

func TestFundCommunityPool_GetSigners(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		address := app.RandomAddress()
		signers := types.MsgFundCommunityPool{
			Depositor: address.String(),
		}.GetSigners()
		require.Len(t, signers, 1)
		require.Equal(t, address, signers[0])
	})

	t.Run("panics when depositor is invalid", func(t *testing.T) {
		require.Panics(t, func() {
			types.MsgFundCommunityPool{
				Depositor: "not-an-address",
			}.GetSigners()
		})
	})
}

func TestMsgUpdateParams_GetSigners(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		address := app.RandomAddress()
		signers := types.MsgUpdateParams{
			Authority: address.String(),
		}.GetSigners()
		require.Len(t, signers, 1)
		require.Equal(t, address, signers[0])
	})

	t.Run("panics when depositor is invalid", func(t *testing.T) {
		require.Panics(t, func() {
			types.MsgUpdateParams{
				Authority: "not-an-address",
			}.GetSigners()
		})
	})
}
