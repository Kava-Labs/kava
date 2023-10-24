package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetModuleAddress(name string) sdk.AccAddress
}

// BankKeeper defines the contract needed to be fulfilled for banking dependencies.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	GetSupply(ctx sdk.Context, denom string) sdk.Coin
}

// CdpKeeper defines the contract needed to be fulfilled for cdp dependencies.
type CdpKeeper interface {
	RepayPrincipal(ctx sdk.Context, owner sdk.AccAddress, collateralType string, payment sdk.Coin) error
	WithdrawCollateral(ctx sdk.Context, owner, depositor sdk.AccAddress, collateral sdk.Coin, collateralType string) error
}

// HardKeeper defines the contract needed to be fulfilled for Kava Lend dependencies.
type HardKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
}

// DistributionKeeper defines the contract needed to be fulfilled for distribution dependencies.
type DistributionKeeper interface {
	DistributeFromFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
	GetCommunityTax(ctx sdk.Context) sdk.Dec
}

type MintKeeper interface {
	GetMinter(ctx sdk.Context) (minter minttypes.Minter)
}

// StakingKeeper expected interface for the staking keeper
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	TotalBondedTokens(ctx sdk.Context) sdkmath.Int
}
