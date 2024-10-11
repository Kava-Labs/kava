package types

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank keeper (noalias)
type BankKeeper interface {
	GetSupply(ctx context.Context, denom string) sdk.Coin
}
