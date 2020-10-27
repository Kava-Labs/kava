package types // noalias

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"

	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) exported.ModuleAccountI
	GetSupply(ctx sdk.Context) (supply exported.SupplyI)
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// AccountKeeper defines the expected keeper interface for interacting with account
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.Account
	SetAccount(ctx sdk.Context, acc authexported.Account)
}

// StakingKeeper defines the expected keeper interface for the staking keeper
type StakingKeeper interface {
	IterateLastValidators(ctx sdk.Context, fn func(index int64, validator stakingexported.ValidatorI) (stop bool))
	IterateValidators(sdk.Context, func(index int64, validator stakingexported.ValidatorI) (stop bool))
	IterateAllDelegations(ctx sdk.Context, cb func(delegation stakingtypes.Delegation) (stop bool))
	GetBondedPool(ctx sdk.Context) (bondedPool exported.ModuleAccountI)
	BondDenom(ctx sdk.Context) (res string)
}

// PricefeedKeeper defines the expected interface for the pricefeed
type PricefeedKeeper interface {
	GetCurrentPrice(sdk.Context, string) (pftypes.CurrentPrice, error)
	GetLiveMarketIDByDenom(sdk.Context, string) (string, bool)
}
