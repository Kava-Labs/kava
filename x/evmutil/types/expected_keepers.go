package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	evmtypes.BankKeeper

	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
