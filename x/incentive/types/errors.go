package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	ErrClaimNotFound                 = sdkerrors.Register(ModuleName, 1, "no claim with input id found for owner and collateral type")
	ErrClaimPeriodNotFound           = sdkerrors.Register(ModuleName, 2, "no claim period found for id and collateral type")
	ErrInvalidAccountType            = sdkerrors.Register(ModuleName, 3, "account type not supported")
	ErrNoClaimsFound                 = sdkerrors.Register(ModuleName, 4, "no claims with collateral type found for address")
	ErrInsufficientModAccountBalance = sdkerrors.Register(ModuleName, 5, "module account has insufficient balance to pay claim")
)
