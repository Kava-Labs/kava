package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
)

type BankKeeper interface {
	GetCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	HasCoins(sdk.Context, sdk.AccAddress, sdk.Coins) bool
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
}

type PricefeedKeeper interface {
	GetCurrentPrice(sdk.Context, string) pftypes.CurrentPrice
	GetAssetParams(sdk.Context) pftypes.AssetParams
	// These are used for testing TODO replace mockApp with keeper in tests to remove these
	SetAssetParams(sdk.Context, pftypes.AssetParams)
	SetPrice(sdk.Context, sdk.AccAddress, string, sdk.Dec, sdk.Int) (pftypes.PostedPrice, sdk.Error)
	SetCurrentPrices(sdk.Context) sdk.Error
}
