package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
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

// ConvertToShares converts a given amount of tokens to shares.
func (k *Keeper) ConvertToShares(ctx sdk.Context, assets sdk.Coin) (types.VaultShare, error) {
	// sharePrice := totalTokens / shareCount
	// issuedShares := amount * sharePrice
	// issuedShares := amount * (totalTokens / shareCount)
	totalShares, found := k.GetVaultTotalShares(ctx, assets.Denom)
	if !found {
		return types.VaultShare{}, fmt.Errorf("vault denom %s not found", assets.Denom)
	}

	totalValue, err := k.GetVaultTotalValue(ctx, assets.Denom)
	if err != nil {
		return types.VaultShare{}, err
	}

	// amount * totalTokens / shareCount
	shareCount := assets.Amount.Mul(totalShares.Amount).Quo(totalValue.Amount)

	if shareCount.IsZero() {
		return types.VaultShare{}, fmt.Errorf("share count is zero")
	}

	return types.NewVaultShare(assets.Denom, shareCount), nil
}
