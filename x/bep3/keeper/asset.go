package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// TODO: Consider gov proposal to remove asset - must clear out asset supply amounts?

// IncrementCurrentAssetSupply increments an asset's supply by the coin
func (k Keeper) IncrementCurrentAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	supply, found := k.GetAssetSupply(ctx, []byte(coin.Denom))
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}

	if !supply.Limit.IsGTE(supply.CurrentSupply.Add(coin)) {
		return types.ErrExceedsSupplyLimit(k.codespace, coin, supply.CurrentSupply, supply.Limit)
	}

	supply.CurrentSupply = supply.CurrentSupply.Add(coin)
	k.SetAssetSupply(ctx, supply, []byte(coin.Denom))
	return nil
}

// DecrementCurrentAssetSupply decrement an asset's supply by the coin
func (k Keeper) DecrementCurrentAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	supply, found := k.GetAssetSupply(ctx, []byte(coin.Denom))
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}

	// Use sdk.Int instead of sdk.Coin to prevent panic if true
	if supply.CurrentSupply.Amount.Sub(coin.Amount).IsNegative() {
		return types.ErrInvalidCurrentSupply(k.codespace, coin, supply.CurrentSupply)
	}

	supply.CurrentSupply = supply.CurrentSupply.Sub(coin)
	k.SetAssetSupply(ctx, supply, []byte(coin.Denom))
	return nil
}

// IncrementIncomingAssetSupply increments an asset's incoming supply
func (k Keeper) IncrementIncomingAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	supply, found := k.GetAssetSupply(ctx, []byte(coin.Denom))
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}

	totalSupply := supply.CurrentSupply.Add(supply.IncomingSupply)
	if !supply.Limit.IsGTE(totalSupply.Add(coin)) {
		return types.ErrExceedsSupplyLimit(k.codespace, coin, totalSupply, supply.Limit)
	}

	supply.IncomingSupply = supply.IncomingSupply.Add(coin)
	k.SetAssetSupply(ctx, supply, []byte(coin.Denom))
	return nil
}

// DecrementIncomingAssetSupply decrements an asset's incoming supply
func (k Keeper) DecrementIncomingAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	supply, found := k.GetAssetSupply(ctx, []byte(coin.Denom))
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}

	// Use sdk.Int instead of sdk.Coin to prevent panic if true
	if supply.IncomingSupply.Amount.Sub(coin.Amount).IsNegative() {
		return types.ErrInvalidIncomingSupply(k.codespace, coin, supply.IncomingSupply)
	}

	supply.IncomingSupply = supply.IncomingSupply.Sub(coin)
	k.SetAssetSupply(ctx, supply, []byte(coin.Denom))
	return nil
}

// IncrementOutgoingAssetSupply increments an asset's outoing supply
func (k Keeper) IncrementOutgoingAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	supply, found := k.GetAssetSupply(ctx, []byte(coin.Denom))
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}

	if supply.CurrentSupply.IsLT(supply.OutgoingSupply.Add(coin)) {
		return types.ErrExceedsAvailableSupply(k.codespace, coin,
			supply.CurrentSupply.Amount.Sub(supply.OutgoingSupply.Amount))
	}

	supply.OutgoingSupply = supply.OutgoingSupply.Add(coin)
	k.SetAssetSupply(ctx, supply, []byte(coin.Denom))
	return nil
}

// DecrementOutgoingAssetSupply decrements an asset's outoing supply
func (k Keeper) DecrementOutgoingAssetSupply(ctx sdk.Context, coin sdk.Coin) sdk.Error {
	supply, found := k.GetAssetSupply(ctx, []byte(coin.Denom))
	if !found {
		return types.ErrAssetNotSupported(k.codespace, coin.Denom)
	}

	// Use sdk.Int instead of sdk.Coin to prevent panic if true
	if supply.OutgoingSupply.Amount.Sub(coin.Amount).IsNegative() {
		return types.ErrInvalidOutgoingSupply(k.codespace, coin, supply.OutgoingSupply)
	}

	supply.OutgoingSupply = supply.OutgoingSupply.Sub(coin)
	k.SetAssetSupply(ctx, supply, []byte(coin.Denom))
	return nil
}
