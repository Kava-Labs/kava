package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/issuance/types"
)

// CreateNewAssetSupply creates a new AssetSupply in the store for the input denom
func (k Keeper) CreateNewAssetSupply(ctx sdk.Context, denom string) types.AssetSupply {
	supply := types.NewAssetSupply(
		sdk.NewCoin(denom, sdk.ZeroInt()), time.Duration(0))
	k.SetAssetSupply(ctx, supply, denom)
	return supply
}

// IncrementCurrentAssetSupply increments an asset's supply by the coin
func (k Keeper) IncrementCurrentAssetSupply(ctx sdk.Context, coin sdk.Coin) error {
	supply, found := k.GetAssetSupply(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotFound, coin.Denom)
	}

	limit, err := k.GetRateLimit(ctx, coin.Denom)
	if err != nil {
		return err
	}

	if limit.Active {
		supplyLimit := sdk.NewCoin(coin.Denom, limit.Limit)
		// Resulting current supply must be under asset's limit
		if supplyLimit.IsLT(supply.CurrentSupply.Add(coin)) {
			return sdkerrors.Wrapf(types.ErrExceedsSupplyLimit, "increase %s, asset supply %s, limit %s", coin, supply.CurrentSupply, supplyLimit)
		}
		supply.CurrentSupply = supply.CurrentSupply.Add(coin)
		k.SetAssetSupply(ctx, supply, coin.Denom)
	}
	return nil
}

// UpdateTimeBasedSupplyLimits updates the time based supply for each asset, resetting it if the current time window has elapsed.
func (k Keeper) UpdateTimeBasedSupplyLimits(ctx sdk.Context) {
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
		k.SetPreviousBlockTime(ctx, previousBlockTime)
	}
	timeElapsed := ctx.BlockTime().Sub(previousBlockTime)
	for _, asset := range params.Assets {
		supply, found := k.GetAssetSupply(ctx, asset.Denom)
		// if a new asset has been added by governance, create a new asset supply for it in the store
		if !found {
			supply = k.CreateNewAssetSupply(ctx, asset.Denom)
		}
		if !asset.RateLimit.Active {
			// rate limiting is not active, reset supply
			supply.CurrentSupply = sdk.NewCoin(asset.Denom, sdk.ZeroInt())
			supply.TimeElapsed = time.Duration(0)
			k.SetAssetSupply(ctx, supply, asset.Denom)
			continue
		}
		if asset.RateLimit.TimePeriod > supply.TimeElapsed+timeElapsed {
			// rate limiting is active, the rate-limiting period has not expired
			supply.TimeElapsed = supply.TimeElapsed + timeElapsed
			k.SetAssetSupply(ctx, supply, asset.Denom)
			continue
		}
		// rate limiting is active, the rate-limiting period has expired, and is now reset
		supply.TimeElapsed = time.Duration(0)
		supply.CurrentSupply = sdk.NewCoin(asset.Denom, sdk.ZeroInt())
		k.SetAssetSupply(ctx, supply, asset.Denom)
	}
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
