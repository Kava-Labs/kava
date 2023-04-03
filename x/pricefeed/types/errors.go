package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = errorsmod.Register(ModuleName, 2, "input must not be empty")
	// ErrExpired error for posted price messages with expired price
	ErrExpired = errorsmod.Register(ModuleName, 3, "price is expired")
	// ErrNoValidPrice error for posted price messages with expired price
	ErrNoValidPrice = errorsmod.Register(ModuleName, 4, "all input prices are expired")
	// ErrInvalidMarket error for posted price messages for invalid markets
	ErrInvalidMarket = errorsmod.Register(ModuleName, 5, "market does not exist")
	// ErrInvalidOracle error for posted price messages for invalid oracles
	ErrInvalidOracle = errorsmod.Register(ModuleName, 6, "oracle does not exist or not authorized")
	// ErrAssetNotFound error for not found asset
	ErrAssetNotFound = errorsmod.Register(ModuleName, 7, "asset not found")
)
