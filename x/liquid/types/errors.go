package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoValidatorFound           = sdkerrors.New(ModuleName, 2, "validator does not exist")
	ErrNoDelegatorForAddress      = sdkerrors.New(ModuleName, 3, "delegator does not contain delegation")
	ErrInvalidDenom               = sdkerrors.New(ModuleName, 4, "invalid denom")
	ErrNotEnoughDelegationShares  = sdkerrors.New(ModuleName, 5, "not enough delegation shares")
	ErrRedelegationsNotCompleted  = sdkerrors.New(ModuleName, 6, "active redelegations cannot be transferred")
	ErrUntransferableShares       = sdkerrors.New(ModuleName, 7, "shares cannot be transferred")
	ErrSelfDelegationBelowMinimum = sdkerrors.Register(ModuleName, 8, "validator's self delegation must be greater than their minimum self delegation")
)
