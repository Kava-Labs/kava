package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var _ sdk.AnteDecorator = VestingAccountDecorator{}

// VestingAccountDecorator blocks vesting messages from reaching the mempool
type VestingAccountDecorator struct {
	disabledMsgTypeUrls []string
}

func NewVestingAccountDecorator() VestingAccountDecorator {
	return VestingAccountDecorator{
		disabledMsgTypeUrls: []string{
			sdk.MsgTypeURL(&vesting.MsgCreateVestingAccount{}),
			sdk.MsgTypeURL(&vesting.MsgCreatePermanentLockedAccount{}),
			sdk.MsgTypeURL(&vesting.MsgCreatePeriodicVestingAccount{}),
		},
	}
}

func (vad VestingAccountDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		typeUrl := sdk.MsgTypeURL(msg)

		for _, disabledTypeUrl := range vad.disabledMsgTypeUrls {
			if typeUrl == disabledTypeUrl {
				return ctx, errorsmod.Wrapf(
					sdkerrors.ErrUnauthorized,
					"MsgTypeURL %s not supported",
					typeUrl,
				)
			}
		}
	}

	return next(ctx, tx, simulate)
}
