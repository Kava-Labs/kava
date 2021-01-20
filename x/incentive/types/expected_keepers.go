package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
)

// SupplyKeeper defines the expected supply keeper for module accounts
type SupplyKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// CdpKeeper defines the expected cdp keeper for interacting with cdps
type CdpKeeper interface {
	IterateCdpsByCollateralType(ctx sdk.Context, collateralType string, cb func(cdp cdptypes.CDP) (stop bool))
	GetTotalPrincipal(ctx sdk.Context, collateralType string, principalDenom string) (total sdk.Int)
	GetCdpByOwnerAndCollateralType(ctx sdk.Context, owner sdk.AccAddress, collateralType string) (cdptypes.CDP, bool)
}

// HardKeeper defines the expected hard keeper for interacting with Hard protocol
type HardKeeper interface {
	GetInterestFactor(ctx sdk.Context, denom string) (sdk.Dec, bool)
	GetBorrowedCoins(ctx sdk.Context) (coins sdk.Coins, found bool)
	GetSuppliedCoins(ctx sdk.Context) (coins sdk.Coins, found bool)
}

// AccountKeeper defines the expected keeper interface for interacting with account
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.Account
	SetAccount(ctx sdk.Context, acc authexported.Account)
}

// CDPHooks event hooks for other keepers to run code in response to CDP modifications
type CDPHooks interface {
	AfterCDPCreated(ctx sdk.Context, cdp cdptypes.CDP)
	BeforeCDPModified(ctx sdk.Context, cdp cdptypes.CDP)
}

// HARDHooks event hooks for other keepers to run code in response to HARD modifications
type HARDHooks interface {
	AfterDepositCreated(ctx sdk.Context, deposit hardtypes.Deposit)
	BeforeDepositModified(ctx sdk.Context, deposit hardtypes.Deposit)
	AfterDepositModified(ctx sdk.Context, deposit hardtypes.Deposit)
	AfterBorrowCreated(ctx sdk.Context, borrow hardtypes.Borrow)
	BeforeBorrowModified(ctx sdk.Context, borrow hardtypes.Borrow)
	AfterBorrowModified(ctx sdk.Context, deposit hardtypes.Deposit)
}
