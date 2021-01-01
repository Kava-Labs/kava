package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MultiBankHooks combine multiple banks hooks, all hook functions are run in array sequence
type MultiBankHooks []BankHooks

// NewMultiBankHooks returns a new MultiBankHooks
func NewMultiBankHooks(hooks ...BankHooks) MultiBankHooks {
	return hooks
}

// BeforeSend function run before send
func (h MultiBankHooks) BeforeSend(ctx sdk.Context, sender sdk.AccAddress, receiver sdk.AccAddress, amount sdk.Coins) error {
	for i := range h {
		err := h[i].BeforeSend(ctx, sender, receiver, amount)
		if err != nil {
			return err
		}
	}
	return nil
}

// BeforeMultiSend function run before multi-send
func (h MultiBankHooks) BeforeMultiSend(ctx sdk.Context, inputs []Input, outputs []Output) error {
	for i := range h {
		err := h[i].BeforeMultiSend(ctx, inputs, outputs)
		if err != nil {
			return err
		}
	}
	return nil
}
