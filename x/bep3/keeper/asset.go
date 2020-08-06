package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/bep3/types"
)

// IncrementCurrentAssetSupply increments an asset's supply by the coin
func (k Keeper) IncrementCurrentAssetSupply(ctx sdk.Context, coin sdk.Coin) error {
	supply, found := k.GetAssetSupply(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotSupported, coin.Denom)
	}

	limit, err := k.GetSupplyLimit(ctx, coin.Denom)
	if err != nil {
		return err
	}
	supplyLimit := sdk.NewCoin(coin.Denom, limit)

	// Resulting current supply must be under asset's limit
	if supplyLimit.IsLT(supply.CurrentSupply.Add(coin)) {
		return sdkerrors.Wrapf(types.ErrExceedsSupplyLimit, "increase %s, asset supply %s, limit %s", coin, supply.CurrentSupply, supplyLimit)
	}

	supply.CurrentSupply = supply.CurrentSupply.Add(coin)
	k.SetAssetSupply(ctx, supply, coin.Denom)
	return nil
}

// DecrementCurrentAssetSupply decrement an asset's supply by the coin
func (k Keeper) DecrementCurrentAssetSupply(ctx sdk.Context, coin sdk.Coin) error {
	supply, found := k.GetAssetSupply(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotSupported, coin.Denom)
	}

	// Resulting current supply must be greater than or equal to 0
	// Use sdk.Int instead of sdk.Coin to prevent panic if true
	if supply.CurrentSupply.Amount.Sub(coin.Amount).IsNegative() {
		return sdkerrors.Wrapf(types.ErrInvalidCurrentSupply, "decrease %s, asset supply %s", coin, supply.CurrentSupply)
	}

	supply.CurrentSupply = supply.CurrentSupply.Sub(coin)
	k.SetAssetSupply(ctx, supply, coin.Denom)
	return nil
}

// IncrementIncomingAssetSupply increments an asset's incoming supply
func (k Keeper) IncrementIncomingAssetSupply(ctx sdk.Context, coin sdk.Coin) error {
	supply, found := k.GetAssetSupply(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotSupported, coin.Denom)
	}

	// 	Result of (current + incoming + amount) must be under asset's limit
	totalSupply := supply.CurrentSupply.Add(supply.IncomingSupply)

	limit, err := k.GetSupplyLimit(ctx, coin.Denom)
	if err != nil {
		return err
	}
	supplyLimit := sdk.NewCoin(coin.Denom, limit)
	if supplyLimit.IsLT(totalSupply.Add(coin)) {
		return sdkerrors.Wrapf(types.ErrExceedsSupplyLimit, "increase %s, asset supply %s, limit %s", coin, totalSupply, supplyLimit)
	}

	supply.IncomingSupply = supply.IncomingSupply.Add(coin)
	k.SetAssetSupply(ctx, supply, coin.Denom)
	return nil
}

// DecrementIncomingAssetSupply decrements an asset's incoming supply
func (k Keeper) DecrementIncomingAssetSupply(ctx sdk.Context, coin sdk.Coin) error {
	supply, found := k.GetAssetSupply(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotSupported, coin.Denom)
	}

	// Resulting incoming supply must be greater than or equal to 0
	// Use sdk.Int instead of sdk.Coin to prevent panic if true
	if supply.IncomingSupply.Amount.Sub(coin.Amount).IsNegative() {
		return sdkerrors.Wrapf(types.ErrInvalidIncomingSupply, "decrease %s, incoming supply %s", coin, supply.IncomingSupply)
	}

	supply.IncomingSupply = supply.IncomingSupply.Sub(coin)
	k.SetAssetSupply(ctx, supply, coin.Denom)
	return nil
}

// IncrementOutgoingAssetSupply increments an asset's outgoing supply
func (k Keeper) IncrementOutgoingAssetSupply(ctx sdk.Context, coin sdk.Coin) error {
	supply, found := k.GetAssetSupply(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotSupported, coin.Denom)
	}

	// Result of (outgoing + amount) must be less than current supply
	if supply.CurrentSupply.IsLT(supply.OutgoingSupply.Add(coin)) {
		return sdkerrors.Wrapf(types.ErrExceedsAvailableSupply, "swap amount %s, available supply %s", coin,
			supply.CurrentSupply.Amount.Sub(supply.OutgoingSupply.Amount))
	}

	supply.OutgoingSupply = supply.OutgoingSupply.Add(coin)
	k.SetAssetSupply(ctx, supply, coin.Denom)
	return nil
}

// DecrementOutgoingAssetSupply decrements an asset's outgoing supply
func (k Keeper) DecrementOutgoingAssetSupply(ctx sdk.Context, coin sdk.Coin) error {
	supply, found := k.GetAssetSupply(ctx, coin.Denom)
	if !found {
		return sdkerrors.Wrap(types.ErrAssetNotSupported, coin.Denom)
	}

	// Resulting outgoing supply must be greater than or equal to 0
	// Use sdk.Int instead of sdk.Coin to prevent panic if true
	if supply.OutgoingSupply.Amount.Sub(coin.Amount).IsNegative() {
		return sdkerrors.Wrapf(types.ErrInvalidOutgoingSupply, "decrease %s, outgoing supply %s", coin, supply.OutgoingSupply)
	}

	supply.OutgoingSupply = supply.OutgoingSupply.Sub(coin)
	k.SetAssetSupply(ctx, supply, coin.Denom)
	return nil
}
