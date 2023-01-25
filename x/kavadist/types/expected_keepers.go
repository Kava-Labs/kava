package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

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

// DistKeeper defines the expected distribution keeper interface
type DistKeeper interface {
	DistributeFromFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// HardKeeper defines the contract needed to be fulfilled for Kava Lend dependencies.
type HardKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
}
