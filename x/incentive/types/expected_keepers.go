package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// SupplyKeeper defines the expected supply keeper for module accounts
type SupplyKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
}

// CdpKeeper defines the expected cdp keeper for interacting with cdps
type CdpKeeper interface {
	IterateCdpsByDenom(ctx sdk.Context, denom string, cb func(cdp cdptypes.CDP) (stop bool))
	GetTotalPrincipal(ctx sdk.Context, collateralDenom string, principalDenom string) (total sdk.Int)
}

// AccountKeeper defines the expected keeper interface for interacting with account
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.Account
	SetAccount(ctx sdk.Context, acc authexported.Account)
}
