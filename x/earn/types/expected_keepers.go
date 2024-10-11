package types

import (
	"context"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
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

// DistributionKeeper defines the expected interface needed for community-pool deposits to earn vaults
type DistributionKeeper interface {
	GetFeePool(ctx sdk.Context) (feePool disttypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool disttypes.FeePool)
	GetDistributionAccount(ctx sdk.Context) sdk.ModuleAccountI
	DistributeFromFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error
}

// LiquidKeeper defines the expected interface needed for derivative to staked token conversions.
type LiquidKeeper interface {
	GetStakedTokensForDerivatives(ctx sdk.Context, derivatives sdk.Coins) (sdk.Coin, error)
	IsDerivativeDenom(ctx sdk.Context, denom string) bool
}

// HardKeeper defines the expected interface needed for the hard strategy.
type HardKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error

	GetSyncedDeposit(ctx sdk.Context, depositor sdk.AccAddress) (hardtypes.Deposit, bool)
}

// SavingsKeeper defines the expected interface needed for the savings strategy.
type SavingsKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error

	GetDeposit(ctx sdk.Context, depositor sdk.AccAddress) (savingstypes.Deposit, bool)
}

// EarnHooks are event hooks called when a user's deposit to a earn vault changes.
type EarnHooks interface {
	AfterVaultDepositCreated(ctx context.Context, vaultDenom string, depositor sdk.AccAddress, sharesOwned sdkmath.LegacyDec)
	BeforeVaultDepositModified(ctx context.Context, vaultDenom string, depositor sdk.AccAddress, sharesOwned sdkmath.LegacyDec)
}
