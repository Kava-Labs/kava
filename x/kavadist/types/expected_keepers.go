package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// DistKeeper defines the expected distribution keeper interface
type DistKeeper interface {
	DistributeFromFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error
}

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authTypes.ModuleAccountI
	SetModuleAccount(ctx sdk.Context, macc authTypes.ModuleAccountI)
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authTypes.AccountI
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}
