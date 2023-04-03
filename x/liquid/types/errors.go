package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrNoValidatorFound           = errorsmod.Register(ModuleName, 2, "validator does not exist")
	ErrNoDelegatorForAddress      = errorsmod.Register(ModuleName, 3, "delegator does not contain delegation")
	ErrInvalidDenom               = errorsmod.Register(ModuleName, 4, "invalid denom")
	ErrNotEnoughDelegationShares  = errorsmod.Register(ModuleName, 5, "not enough delegation shares")
	ErrRedelegationsNotCompleted  = errorsmod.Register(ModuleName, 6, "active redelegations cannot be transferred")
	ErrUntransferableShares       = errorsmod.Register(ModuleName, 7, "shares cannot be transferred")
	ErrSelfDelegationBelowMinimum = errorsmod.Register(ModuleName, 8, "validator's self delegation must be greater than their minimum self delegation")
)
