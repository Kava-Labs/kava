package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// VestingAccountDecorator blocks MsgCreateVestingAccount from reaching the mempool
type VestingAccountDecorator struct{}

func NewVestingAccountDecorator() VestingAccountDecorator {
	return VestingAccountDecorator{}
}

func (vad VestingAccountDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		_, ok := msg.(*vesting.MsgCreateVestingAccount)
		if ok {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "MsgCreateVestingAccount not supported")

		}
	}
	return next(ctx, tx, simulate)
}
