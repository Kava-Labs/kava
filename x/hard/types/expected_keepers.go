package types // noalias

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected keeper interface for interacting with account
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
}

// StakingKeeper defines the expected keeper interface for the staking keeper
type StakingKeeper interface {
	IterateLastValidators(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	IterateValidators(sdk.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	IterateAllDelegations(ctx sdk.Context, cb func(delegation stakingtypes.Delegation) (stop bool))
	GetBondedPool(ctx sdk.Context) (bondedPool authtypes.ModuleAccountI)
	BondDenom(ctx sdk.Context) (res string)
}

// PricefeedKeeper defines the expected interface for the pricefeed
type PricefeedKeeper interface {
	GetCurrentPrice(sdk.Context, string) (pftypes.CurrentPrice, error)
}

// AuctionKeeper expected interface for the auction keeper (noalias)
type AuctionKeeper interface {
	StartCollateralAuction(ctx sdk.Context, seller string, lot sdk.Coin, maxBid sdk.Coin, lotReturnAddrs []sdk.AccAddress, lotReturnWeights []sdk.Int, debt sdk.Coin) (uint64, error)
}

// HARDHooks event hooks for other keepers to run code in response to HARD modifications
type HARDHooks interface {
	AfterDepositCreated(ctx sdk.Context, deposit Deposit)
	BeforeDepositModified(ctx sdk.Context, deposit Deposit, newDepositDenoms []string)
	AfterDepositModified(ctx sdk.Context, deposit Deposit)
	AfterBorrowCreated(ctx sdk.Context, borrow Borrow)
	BeforeBorrowModified(ctx sdk.Context, borrow Borrow, newBorrowDenoms []string)
	AfterBorrowModified(ctx sdk.Context, borrow Borrow)
}
