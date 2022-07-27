package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetVaultPricePerShare returns the price per share for a given vault denom.
func (k *Keeper) GetVaultPricePerShare(ctx sdk.Context, denom string) (sdk.Int, error) {
	// totalTokens := shareCount * sharePrice
	// sharePrice := totalTokens / shareCount
	totalShares, found := k.GetVaultTotalShares(ctx, denom)
	if !found {
		return sdk.ZeroInt(), fmt.Errorf("vault denom %s not found", denom)
	}

	totalValue, err := k.GetVaultTotalValue(ctx, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// totalTokens / totalShares
	sharePrice := totalValue.Amount.Quo(totalShares.Amount)

	if sharePrice.IsZero() {
		return sdk.ZeroInt(), fmt.Errorf("share price is zero")
	}

	return sharePrice, nil
}
