package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrEmptyInput error for empty input
	ErrEmptyInput = sdkerrors.Register(ModuleName, 2, "input must not be empty")
	// ErrExpired error for posted price messages with expired price
	ErrExpired = sdkerrors.Register(ModuleName, 3, "price is expired")
	// ErrNoValidPrice error for posted price messages with expired price
	ErrNoValidPrice = sdkerrors.Register(ModuleName, 4, "all input prices are expired")
	// ErrInvalidMarket error for posted price messages for invalid markets
	ErrInvalidMarket = sdkerrors.Register(ModuleName, 5, "market does not exist")
	// ErrInvalidOracle error for posted price messages for invalid oracles
	ErrInvalidOracle = sdkerrors.Register(ModuleName, 6, "oracle does not exist or not authorized")
)
