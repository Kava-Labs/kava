package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// CdpKeeper expected interface for the cdp keeper
type CdpKeeper interface {
	GetCDP(sdk.Context, sdk.AccAddress, string) (cdptypes.CDP, bool)
	PartialSeizeCDP(sdk.Context, sdk.AccAddress, string, sdk.Int, sdk.Int) sdk.Error
	ReduceGlobalDebt(sdk.Context, sdk.Int) sdk.Error
	GetGovDenom(sdk.Context) string
	GetLiquidatorAccountAddress() sdk.AccAddress // This won't need to exist once the module account is defined in this module (instead of in the cdp module)
}

// BankKeeper expected interface for the bank keeper
type BankKeeper interface {
	GetCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
}

// AuctionKeeper expected interface for the auction keeper
type AuctionKeeper interface {
	StartForwardAuction(sdk.Context, sdk.AccAddress, sdk.Coin, sdk.Coin) (auction.ID, sdk.Error)
	StartReverseAuction(sdk.Context, sdk.AccAddress, sdk.Coin, sdk.Coin) (auction.ID, sdk.Error)
	StartForwardReverseAuction(sdk.Context, sdk.AccAddress, sdk.Coin, sdk.Coin, sdk.AccAddress) (auction.ID, sdk.Error)
}
