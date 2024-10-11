package types

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type ParamKeeper interface {
	GetSubspace(string) (paramstypes.Subspace, bool)
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetSupply(ctx context.Context, denom string) sdk.Coin
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
