package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// CommunityKeeper defines the expected community keeper interface
type CommunityKeeper interface {
	DistributeFromCommunityPool(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Coins) error
	FundCommunityPool(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Coins) error
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
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	BondDenom(sdk.Context) string
	TotalBondedTokens(sdk.Context) sdk.Int
}
