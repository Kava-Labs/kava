package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/bank/internal/types"
)

var _ types.BankHooks = BaseKeeper{}

// BeforeSend call hook if registered
func (k BaseKeeper) BeforeSend(ctx sdk.Context, sender, receiver sdk.AccAddress, amount sdk.Coins) error {
	if k.hooks != nil {
		return k.hooks.BeforeSend(ctx, sender, receiver, amount)
	}
	return nil
}

// BeforeMultiSend call hook if registered
func (k BaseKeeper) BeforeMultiSend(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error {
	if k.hooks != nil {
		return k.hooks.BeforeMultiSend(ctx, inputs, outputs)
	}
	return nil
}
