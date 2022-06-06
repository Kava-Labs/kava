package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var _ sdk.AnteDecorator = VestingAccountDecorator{}

// VestingAccountDecorator blocks MsgCreateVestingAccount from reaching the mempool
type VestingAccountDecorator struct{}

func NewVestingAccountDecorator() VestingAccountDecorator {
	return VestingAccountDecorator{}
}

func (vad VestingAccountDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		if _, ok := msg.(*vesting.MsgCreateVestingAccount); ok {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "MsgCreateVestingAccount not supported")
		}
	}

	return next(ctx, tx, simulate)
}
