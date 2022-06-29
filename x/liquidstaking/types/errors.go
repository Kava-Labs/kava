package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = sdkerrors.Register(ModuleName, 2, "input must not be empty")
	// ErrNoDerivativeFound error for no derivative found in store
	ErrNoDerivativeFound = sdkerrors.Register(ModuleName, 3, "no derivative found")
	// ErrInvalidBurnAmount error for an invalid burn amount
	ErrInvalidBurnAmount               = sdkerrors.Register(ModuleName, 4, "invalid burn amount")
	ErrNoValidatorFound                = sdkerrors.New(ModuleName, 5, "validator does not exist")
	ErrNoDelegatorForAddress           = sdkerrors.New(ModuleName, 6, "delegator does not contain delegation")
	ErrOnlyBondDenomAllowdForTokenize  = sdkerrors.New(ModuleName, 7, "only bond denom is allowed for tokenize")
	ErrNotEnoughDelegationShares       = sdkerrors.New(ModuleName, 8, "not enough delegation shares")
	ErrExceedingFreeVestingDelegations = sdkerrors.New(ModuleName, 9, "trying to exceed vested free delegation for vesting account")
	ErrInvalidLiquidCoinDenom          = sdkerrors.New(ModuleName, 10, "invalid liquid staking coin denom")
	ErrNotEnoughBalance                = sdkerrors.New(ModuleName, 11, "insufficient balance of liquid staking coin")
	ErrInvalidDerivativeDenom          = sdkerrors.New(ModuleName, 12, "invalid derivative denom")
)
