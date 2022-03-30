package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MultiSavingsHooks combine multiple Savings hooks, all hook functions are run in array sequence
type MultiSavingsHooks []SavingsHooks

// NewMultiSavingsHooks returns a new MultiSavingsHooks
func NewMultiSavingsHooks(hooks ...SavingsHooks) MultiSavingsHooks {
	return hooks
}

// AfterSavingsDepositCreated runs after a deposit is created
func (s MultiSavingsHooks) AfterSavingsDepositCreated(ctx sdk.Context, deposit Deposit) {
	for i := range s {
		s[i].AfterSavingsDepositCreated(ctx, deposit)
	}
}

// BeforeSavingsDepositModified runs before a deposit is modified
func (s MultiSavingsHooks) BeforeSavingsDepositModified(ctx sdk.Context, deposit Deposit, incomingDenoms []string) {
	for i := range s {
		s[i].BeforeSavingsDepositModified(ctx, deposit, incomingDenoms)
	}
}
