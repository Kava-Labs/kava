package types // noalias

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected keeper interface for interacting with account
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// SavingsHooks event hooks for other keepers to run code in response to Savings modifications
type SavingsHooks interface {
	AfterSavingsDepositCreated(ctx context.Context, deposit Deposit)
	BeforeSavingsDepositModified(ctx context.Context, deposit Deposit, incomingDenoms []string)
}

type LiquidKeeper interface {
	GetStakedTokensForDerivatives(ctx sdk.Context, derivatives sdk.Coins) (sdk.Coin, error)
	IsDerivativeDenom(ctx sdk.Context, denom string) bool
}
