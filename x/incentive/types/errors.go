package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

// Incentive module errors
var (
	ErrClaimNotFound                 = errorsmod.Register(ModuleName, 2, "no claimable rewards found for user")
	ErrRewardPeriodNotFound          = errorsmod.Register(ModuleName, 3, "no reward period found for collateral type")
	ErrInvalidAccountType            = errorsmod.Register(ModuleName, 4, "account type not supported")
	ErrNoClaimsFound                 = errorsmod.Register(ModuleName, 5, "no claimable rewards found")
	ErrInsufficientModAccountBalance = errorsmod.Register(ModuleName, 6, "module account has insufficient balance to pay claim")
	ErrAccountNotFound               = errorsmod.Register(ModuleName, 7, "account not found")
	ErrInvalidMultiplier             = errorsmod.Register(ModuleName, 8, "invalid rewards multiplier")
	ErrZeroClaim                     = errorsmod.Register(ModuleName, 9, "cannot claim - claim amount rounds to zero")
	ErrClaimExpired                  = errorsmod.Register(ModuleName, 10, "claim has expired")
	ErrInvalidClaimType              = errorsmod.Register(ModuleName, 11, "invalid claim type")
	ErrDecreasingRewardFactor        = errorsmod.Register(ModuleName, 13, "found new reward factor less than an old reward factor")
	ErrInvalidClaimDenoms            = errorsmod.Register(ModuleName, 14, "invalid claim denoms")
)
