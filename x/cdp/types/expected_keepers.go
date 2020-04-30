package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// SupplyKeeper defines the expected supply keeper for module accounts  (noalias)
type SupplyKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(sdk.Context, supplyexported.ModuleAccountI)

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	GetSupply(ctx sdk.Context) (supply supplyexported.SupplyI)
}

// PricefeedKeeper defines the expected interface for the pricefeed  (noalias)
type PricefeedKeeper interface {
	GetCurrentPrice(sdk.Context, string) (pftypes.CurrentPrice, error)
	GetParams(sdk.Context) pftypes.Params
	// These are used for testing TODO replace mockApp with keeper in tests to remove these
	SetParams(sdk.Context, pftypes.Params)
	SetPrice(sdk.Context, sdk.AccAddress, string, sdk.Dec, time.Time) (pftypes.PostedPrice, error)
	SetCurrentPrices(sdk.Context, string) error
}

// AuctionKeeper expected interface for the auction keeper (noalias)
type AuctionKeeper interface {
	StartSurplusAuction(ctx sdk.Context, seller string, lot sdk.Coin, bidDenom string) (uint64, error)
	StartDebtAuction(ctx sdk.Context, buyer string, bid sdk.Coin, initialLot sdk.Coin, debt sdk.Coin) (uint64, error)
	StartCollateralAuction(ctx sdk.Context, seller string, lot sdk.Coin, maxBid sdk.Coin, lotReturnAddrs []sdk.AccAddress, lotReturnWeights []sdk.Int, debt sdk.Coin) (uint64, error)
}

// AccountKeeper expected interface for the account keeper (noalias)
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, cb func(account authexported.Account) (stop bool))
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.Account
}
