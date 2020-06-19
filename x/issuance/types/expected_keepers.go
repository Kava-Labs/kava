package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// SupplyKeeper defines the expected supply keeper for module accounts (noalias)
type SupplyKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}
