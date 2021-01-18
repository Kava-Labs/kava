package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MultiHARDHooks combine multiple HARD hooks, all hook functions are run in array sequence
type MultiHARDHooks []HARDHooks

// NewMultiHARDHooks returns a new MultiHARDHooks
func NewMultiHARDHooks(hooks ...HARDHooks) MultiHARDHooks {
	return hooks
}

// AfterDepositCreated runs after a deposit is created
func (h MultiHARDHooks) AfterDepositCreated(ctx sdk.Context, deposit Deposit) {
	for i := range h {
		h[i].AfterDepositCreated(ctx, deposit)
	}
}

// BeforeDepositModified runs before a deposit is modified
func (h MultiHARDHooks) BeforeDepositModified(ctx sdk.Context, deposit Deposit) {
	for i := range h {
		h[i].BeforeDepositModified(ctx, deposit)
	}
}

// AfterDepositModified runs after a deposit is modified
func (h MultiHARDHooks) AfterDepositModified(ctx sdk.Context, deposit Deposit) {
	for i := range h {
		h[i].AfterDepositModified(ctx, deposit)
	}
}

// AfterBorrowCreated runs after a borrow is created
func (h MultiHARDHooks) AfterBorrowCreated(ctx sdk.Context, borrow Borrow) {
	for i := range h {
		h[i].AfterBorrowCreated(ctx, borrow)
	}
}

// BeforeBorrowModified runs before a borrow is modified
func (h MultiHARDHooks) BeforeBorrowModified(ctx sdk.Context, borrow Borrow) {
	for i := range h {
		h[i].BeforeBorrowModified(ctx, borrow)
	}
}

// AfterBorrowModified runs after a borrow is modified
func (h MultiHARDHooks) AfterBorrowModified(ctx sdk.Context, borrow Borrow) {
	for i := range h {
		h[i].AfterBorrowModified(ctx, borrow)
	}
}
