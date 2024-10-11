package types // noalias

import (
	"context"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
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

// StakingKeeper defines the expected keeper interface for the staking keeper
type StakingKeeper interface {
	IterateLastValidators(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	IterateValidators(context.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	IterateAllDelegations(ctx context.Context, cb func(delegation stakingtypes.Delegation) (stop bool))
	GetBondedPool(ctx context.Context) (bondedPool sdk.ModuleAccountI)
	BondDenom(ctx context.Context) (res string, err error)
}

// PricefeedKeeper defines the expected interface for the pricefeed
type PricefeedKeeper interface {
	GetCurrentPrice(sdk.Context, string) (pftypes.CurrentPrice, error)
}

// AuctionKeeper expected interface for the auction keeper (noalias)
type AuctionKeeper interface {
	StartCollateralAuction(ctx sdk.Context, seller string, lot sdk.Coin, maxBid sdk.Coin, lotReturnAddrs []sdk.AccAddress, lotReturnWeights []sdkmath.Int, debt sdk.Coin) (uint64, error)
}

// HARDHooks event hooks for other keepers to run code in response to HARD modifications
type HARDHooks interface {
	AfterDepositCreated(ctx sdk.Context, deposit Deposit)
	BeforeDepositModified(ctx sdk.Context, deposit Deposit)
	AfterDepositModified(ctx sdk.Context, deposit Deposit)
	AfterBorrowCreated(ctx sdk.Context, borrow Borrow)
	BeforeBorrowModified(ctx sdk.Context, borrow Borrow)
	AfterBorrowModified(ctx sdk.Context, borrow Borrow)
}
