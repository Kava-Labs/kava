package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	ErrClaimNotFound                 = sdkerrors.Register(ModuleName, 1, "no claim with input id found for owner and denom")
	ErrClaimPeriodNotFound           = sdkerrors.Register(ModuleName, 2, "no claim period found for id and denom")
	ErrInvalidAccountType            = sdkerrors.Register(ModuleName, 3, "account type not supported")
	ErrNoClaimsFound                 = sdkerrors.Register(ModuleName, 4, "no claims with denom found for address")
	ErrInsufficientModAccountBalance = sdkerrors.Register(ModuleName, 5, "module account has insufficient balance to pay claim")
)
