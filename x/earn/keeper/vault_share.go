package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// ConvertToShares converts a given amount of tokens to shares.
func (k *Keeper) ConvertToShares(ctx sdk.Context, assets sdk.Coin) (types.VaultShare, error) {
	totalShares, found := k.GetVaultTotalShares(ctx, assets.Denom)
	if !found {
		// No shares issued yet, so shares are issued 1:1
		return types.NewVaultShare(assets.Denom, assets.Amount.ToDec()), nil
	}

	totalValue, err := k.GetVaultTotalValue(ctx, assets.Denom)
	if err != nil {
		return types.VaultShare{}, err
	}

	if totalValue.Amount.IsZero() {
		return types.VaultShare{}, fmt.Errorf("total value of vault is zero")
	}

	// sharePrice := totalTokens / shareCount
	// issuedShares = assetAmount / sharePrice
	// issuedShares := assetAmount / (totalTokens / shareCount)
	//               = assetAmount * (shareCount / totalTokens)
	//
	// multiply by reciprocal  of sharePrice to avoid two divisions and limit
	// rounding to one time

	// Per share is not used here as it loses decimal values and can cause a 0
	// share count.
	// Division is done at the last step as decimals are truncated then.
	// For example:
	// 100 * 100 / 101   == 10000 / 101 == 99
	// 100 * (100 / 101) == 100 * 0     == 0
	shareCount := assets.Amount.ToDec().Mul(totalShares.Amount).QuoTruncate(totalValue.Amount.ToDec())

	if shareCount.IsZero() {
		return types.VaultShare{}, fmt.Errorf("share count is zero")
	}

	return types.NewVaultShare(assets.Denom, shareCount), nil
}

// ConvertToAssets converts a given amount of shares to tokens.
func (k *Keeper) ConvertToAssets(ctx sdk.Context, share types.VaultShare) (sdk.Coin, error) {
	totalVaultShares, found := k.GetVaultTotalShares(ctx, share.Denom)
	if !found {
		return sdk.Coin{}, fmt.Errorf("vault for %s not found", share.Denom)
	}

	totalValue, err := k.GetVaultTotalValue(ctx, share.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// percentOwnership := accShares / totalVaultShares
	// accValue := totalValue * percentOwnership
	// accValue := totalValue * accShares / totalVaultShares
	// Division must be last to avoid rounding errors and properly truncate.
	value := totalValue.Amount.ToDec().Mul(share.Amount).QuoTruncate(totalVaultShares.Amount)

	return sdk.NewCoin(share.Denom, value.TruncateInt()), nil
}

// ShareIsDust returns true if the share value is less than 1 coin
func (k *Keeper) ShareIsDust(ctx sdk.Context, share types.VaultShare) (bool, error) {
	coin, err := k.ConvertToAssets(ctx, share)
	if err != nil {
		return false, err
	}

	// Truncated int, becomes zero if < 1
	return coin.IsZero(), nil
}
