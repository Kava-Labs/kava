package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.Account
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.Account
	SetAccount(ctx sdk.Context, acc authexported.Account)
}

// BankKeeper defines the expected bank Keeper
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}