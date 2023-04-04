package ante_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

func newMsgGrant(granter sdk.AccAddress, grantee sdk.AccAddress, a authz.Authorization, expiration time.Time) *authz.MsgGrant {
	msg, err := authz.NewMsgGrant(granter, grantee, a, &expiration)
	if err != nil {
		panic(err)
	}
	return msg
}

func newMsgExec(grantee sdk.AccAddress, msgs []sdk.Msg) *authz.MsgExec {
	msg := authz.NewMsgExec(grantee, msgs)
	return &msg
}

func TestAuthzLimiterDecorator(t *testing.T) {
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(5)
	distantFuture := time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)

	validator := sdk.ValAddress(testAddresses[4])
	stakingAuthDelegate, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{validator}, nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, nil)
	require.NoError(t, err)
	stakingAuthUndelegate, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{validator}, nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE, nil)
	require.NoError(t, err)

	decorator := ante.NewAuthzLimiterDecorator(
		sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
		sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
	)

	testCases := []struct {
		name        string
		msgs        []sdk.Msg
		checkTx     bool
		expectedErr error
	}{
		{
			name: "a non blocked msg is not blocked",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(
					testAddresses[0],
					testAddresses[1],
					sdk.NewCoins(sdk.NewInt64Coin("ukava", 100e6)),
				),
			},
			checkTx: false,
		},
		{
			name: "a blocked msg is not blocked when not wrapped in MsgExec",
			msgs: []sdk.Msg{
				&evmtypes.MsgEthereumTx{},
			},
			checkTx: false,
		},
		{
			name: "when a MsgGrant contains a non blocked msg, it passes",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
					distantFuture,
				),
			},
			checkTx: false,
		},
		{
			name: "when a MsgGrant contains a non blocked msg, it passes",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					stakingAuthDelegate,
					distantFuture,
				),
			},
			checkTx: false,
		},
		{
			name: "when a MsgGrant contains a blocked msg, it is blocked",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					authz.NewGenericAuthorization(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
					distantFuture,
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "when a MsgGrant contains a blocked msg, it is blocked",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					stakingAuthUndelegate,
					distantFuture,
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "when a MsgExec contains a non blocked msg, it passes",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{banktypes.NewMsgSend(
						testAddresses[0],
						testAddresses[3],
						sdk.NewCoins(sdk.NewInt64Coin("ukava", 100e6)),
					)}),
			},
			checkTx: false,
		},
		{
			name: "when a MsgExec contains a blocked msg, it is blocked",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						&evmtypes.MsgEthereumTx{},
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "blocked msg surrounded by valid msgs is still blocked",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					stakingAuthDelegate,
					distantFuture,
				),
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						banktypes.NewMsgSend(
							testAddresses[0],
							testAddresses[3],
							sdk.NewCoins(sdk.NewInt64Coin("ukava", 100e6)),
						),
						&evmtypes.MsgEthereumTx{},
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "a nested MsgExec containing a blocked msg is still blocked",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						newMsgExec(
							testAddresses[2],
							[]sdk.Msg{
								&evmtypes.MsgEthereumTx{},
							},
						),
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "a nested MsgGrant containing a blocked msg is still blocked",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						newMsgGrant(
							testAddresses[0],
							testAddresses[1],
							authz.NewGenericAuthorization(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
							distantFuture,
						),
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
	}

	txConfig := app.MakeEncodingConfig().TxConfig

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := helpers.GenSignedMockTx(
				rand.New(rand.NewSource(time.Now().UnixNano())),
				txConfig,
				tc.msgs,
				sdk.NewCoins(),
				helpers.DefaultGenTxGas,
				"testing-chain-id",
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			)
			require.NoError(t, err)
			mmd := MockAnteHandler{}
			ctx := sdk.Context{}.WithIsCheckTx(tc.checkTx)
			_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
