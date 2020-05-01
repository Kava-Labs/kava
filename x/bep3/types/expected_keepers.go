package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SupplyKeeper defines the expected supply Keeper
type SupplyKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
