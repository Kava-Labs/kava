package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

// Incentive module errors
var (
	ErrClaimNotFound                 = sdkerrors.Register(ModuleName, 2, "no claimable rewards found for user")
	ErrRewardPeriodNotFound          = sdkerrors.Register(ModuleName, 3, "no reward period found for collateral type")
	ErrInvalidAccountType            = sdkerrors.Register(ModuleName, 4, "account type not supported")
	ErrNoClaimsFound                 = sdkerrors.Register(ModuleName, 5, "no claimable rewards found")
	ErrInsufficientModAccountBalance = sdkerrors.Register(ModuleName, 6, "module account has insufficient balance to pay claim")
	ErrAccountNotFound               = sdkerrors.Register(ModuleName, 7, "account not found")
	ErrInvalidMultiplier             = sdkerrors.Register(ModuleName, 8, "invalid rewards multiplier")
	ErrZeroClaim                     = sdkerrors.Register(ModuleName, 9, "cannot claim - claim amount rounds to zero")
	ErrClaimExpired                  = sdkerrors.Register(ModuleName, 10, "claim has expired")
	ErrInvalidClaimType              = sdkerrors.Register(ModuleName, 11, "invalid claim type")
	ErrInvalidClaimOwner             = sdkerrors.Register(ModuleName, 12, "invalid claim owner")
	ErrDecreasingRewardFactor        = sdkerrors.Register(ModuleName, 13, "found new reward factor less than an old reward factor")
	ErrInvalidClaimDenoms            = sdkerrors.Register(ModuleName, 14, "invalid claim denoms")
)
