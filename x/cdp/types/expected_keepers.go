package types

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// BankKeeper defines the expected bank keeper for module accounts
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

var _ BankKeeper = (bankkeeper.Keeper)(nil)

// PricefeedKeeper defines the expected interface for the pricefeed
type PricefeedKeeper interface {
	GetCurrentPrice(context.Context, string) (pftypes.CurrentPrice, error)
	GetParams(context.Context) pftypes.Params
	// These are used for testing TODO replace mockApp with keeper in tests to remove these
	SetParams(context.Context, pftypes.Params)
	SetPrice(context.Context, sdk.AccAddress, string, sdkmath.LegacyDec, time.Time) (pftypes.PostedPrice, error)
	SetCurrentPrices(context.Context, string) error
}

// AuctionKeeper expected interface for the auction keeper
type AuctionKeeper interface {
	StartSurplusAuction(ctx context.Context, seller string, lot sdk.Coin, bidDenom string) (uint64, error)
	StartDebtAuction(ctx context.Context, buyer string, bid sdk.Coin, initialLot sdk.Coin, debt sdk.Coin) (uint64, error)
	StartCollateralAuction(ctx context.Context, seller string, lot sdk.Coin, maxBid sdk.Coin, lotReturnAddrs []sdk.AccAddress, lotReturnWeights []sdkmath.Int, debt sdk.Coin) (uint64, error)
}

// AccountKeeper expected interface for the account keeper
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(context.Context, sdk.ModuleAccountI)

	IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool))
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// CDPHooks event hooks for other keepers to run code in response to CDP modifications
type CDPHooks interface {
	AfterCDPCreated(ctx context.Context, cdp CDP)
	BeforeCDPModified(ctx context.Context, cdp CDP)
}
