package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

// Errors used by the issuance module
var (
	ErrAssetNotFound           = errorsmod.Register(ModuleName, 2, "no asset with input denom found")
	ErrNotAuthorized           = errorsmod.Register(ModuleName, 3, "account not authorized")
	ErrAssetPaused             = errorsmod.Register(ModuleName, 4, "asset is paused")
	ErrAccountBlocked          = errorsmod.Register(ModuleName, 5, "account is blocked")
	ErrAccountAlreadyBlocked   = errorsmod.Register(ModuleName, 6, "account is already blocked")
	ErrAccountAlreadyUnblocked = errorsmod.Register(ModuleName, 7, "account is already unblocked")
	ErrIssueToModuleAccount    = errorsmod.Register(ModuleName, 8, "cannot issue tokens to module account")
	ErrExceedsSupplyLimit      = errorsmod.Register(ModuleName, 9, "asset supply over limit")
	ErrAssetUnblockable        = errorsmod.Register(ModuleName, 10, "asset does not support block/unblock functionality")
	ErrAccountNotFound         = errorsmod.Register(ModuleName, 11, "cannot block account that does not exist in state")
)
