package types

import (
	"context"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetModuleAccount(context.Context, sdk.ModuleAccountI)

	// moved in from supply
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// SwapHooks are event hooks called when a user's deposit to a swap pool changes.
type SwapHooks interface {
	AfterPoolDepositCreated(ctx context.Context, poolID string, depositor sdk.AccAddress, sharedOwned sdkmath.Int)
	BeforePoolDepositModified(ctx context.Context, poolID string, depositor sdk.AccAddress, sharedOwned sdkmath.Int)
}
