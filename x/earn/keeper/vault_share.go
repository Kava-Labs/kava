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
		return types.NewVaultShare(assets.Denom, sdk.NewDecFromInt(assets.Amount)), nil
	}

	totalValue, err := k.GetVaultTotalValue(ctx, assets.Denom)
	if err != nil {
		return types.VaultShare{}, err
	}

	if totalValue.Amount.IsZero() {
		return types.VaultShare{}, fmt.Errorf("total value of vault is zero")
	}

	// sharePrice   = totalValue / totalShares
	// issuedShares = assets / sharePrice
	// issuedShares = assets / (totalValue / totalShares)
	//              = assets * (totalShares / totalValue)
	//              = (assets * totalShares) / totalValue
	//
	// Multiply by reciprocal of sharePrice to avoid two divisions and limit
	// rounding to one time. Per-share price is also not used as there is a loss
	// of precision.

	// Division is done at the last step as there is a slight amount that is
	// rounded down.
	// For example:
	// 100 * 100 / 105   == 10000 / 105                == 95.238095238095238095
	// 100 * (100 / 105) == 100 * 0.952380952380952380 == 95.238095238095238000
	//                    rounded down and truncated ^    loss of precision ^
	issuedShares := sdk.NewDecFromInt(assets.Amount).Mul(totalShares.Amount).QuoTruncate(sdk.NewDecFromInt(totalValue.Amount))

	if issuedShares.IsZero() {
		return types.VaultShare{}, fmt.Errorf("share count is zero")
	}

	return types.NewVaultShare(assets.Denom, issuedShares), nil
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
	value := sdk.NewDecFromInt(totalValue.Amount).Mul(share.Amount).QuoTruncate(totalVaultShares.Amount)

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
