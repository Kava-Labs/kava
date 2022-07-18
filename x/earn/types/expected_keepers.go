package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	SetModuleAccount(sdk.Context, types.ModuleAccountI)
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// HardKeeper defines the expected interface needed for the hard strategy.
type HardKeeper interface {
	// State modification methods
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error

	// Read only methods
	GetSyncedBorrow(ctx sdk.Context, borrower sdk.AccAddress) (hardtypes.Borrow, bool)
}

// SavingsKeeper defines the expected interface needed for the savings strategy.
type SavingsKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error

	GetDeposit(ctx sdk.Context, depositor sdk.AccAddress) (savingstypes.Deposit, bool)
}
