package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/issuance/types"
)

// IssueTokens mints new tokens and sends them to the receiver address
func (k Keeper) IssueTokens(ctx sdk.Context, tokens sdk.Coin, owner, receiver sdk.AccAddress) error {
	asset, found := k.GetAsset(ctx, tokens.Denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrAssetNotFound, "denom: %s", tokens.Denom)
	}
	if !owner.Equals(asset.Owner) {
		return sdkerrors.Wrapf(types.ErrNotAuthorized, "owner: %s, address: %s", asset.Owner, owner)
	}
	if asset.Paused {
		return sdkerrors.Wrapf(types.ErrAssetPaused, "denom: %s", tokens.Denom)
	}
	blocked, _ := k.checkBlockedAddress(ctx, asset, receiver)
	if blocked {
		return sdkerrors.Wrapf(types.ErrAccountBlocked, "address: %s", receiver)
	}

	// mint new tokens
	err := k.supplyKeeper.MintCoins(ctx, types.ModuleAccountName, sdk.NewCoins(tokens))
	if err != nil {
		return err
	}
	// send to receiver
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, receiver, sdk.NewCoins(tokens))
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIssue,
			sdk.NewAttribute(types.AttributeKeyIssueAmount, fmt.Sprintf("%s", tokens)),
		),
	)
	return nil
}

// RedeemTokens sends tokens from the owner address  to the module account and burns them
func (k Keeper) RedeemTokens(ctx sdk.Context, tokens sdk.Coin, owner sdk.AccAddress) error {
	asset, found := k.GetAsset(ctx, tokens.Denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrAssetNotFound, "denom: %s", tokens.Denom)
	}
	if !owner.Equals(asset.Owner) {
		return sdkerrors.Wrapf(types.ErrNotAuthorized, "owner: %s, address: %s", asset.Owner, owner)
	}
	if asset.Paused {
		return sdkerrors.Wrapf(types.ErrAssetPaused, "denom: %s", tokens.Denom)
	}
	coins := sdk.NewCoins(tokens)
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleAccountName, coins)
	if err != nil {
		return err
	}
	err = k.supplyKeeper.BurnCoins(ctx, types.ModuleAccountName, coins)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRedeem,
			sdk.NewAttribute(types.AttributeKeyRedeemAmount, fmt.Sprintf("%s", tokens)),
		),
	)
	return nil
}

// BlockAddress adds an address to the blocked list
func (k Keeper) BlockAddress(ctx sdk.Context, denom string, owner, blockedAddress sdk.AccAddress) error {
	asset, found := k.GetAsset(ctx, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrAssetNotFound, "denom: %s", denom)
	}
	if !owner.Equals(asset.Owner) {
		return sdkerrors.Wrapf(types.ErrNotAuthorized, "owner: %s, address: %s", asset.Owner, owner)
	}
	blocked, _ := k.checkBlockedAddress(ctx, asset, blockedAddress)
	if blocked {
		return nil
	}
	asset.BlockedAddresses = append(asset.BlockedAddresses, blockedAddress)
	k.SetAsset(ctx, asset)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBlock,
			sdk.NewAttribute(types.AttributeKeyBlock, fmt.Sprintf("%s", blockedAddress)),
			sdk.NewAttribute(types.AttributeKeyDenom, asset.Denom),
		),
	)
	return nil
}

// UnblockAddress adds an address to the blocked list
func (k Keeper) UnblockAddress(ctx sdk.Context, denom string, owner, addr sdk.AccAddress) error {
	asset, found := k.GetAsset(ctx, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrAssetNotFound, "denom: %s", denom)
	}
	if !owner.Equals(asset.Owner) {
		return sdkerrors.Wrapf(types.ErrNotAuthorized, "owner: %s, address: %s", asset.Owner, owner)
	}
	blocked, i := k.checkBlockedAddress(ctx, asset, addr)
	if !blocked {
		return nil
	}

	blockedAddrs := k.removeBlockedAddress(ctx, asset.BlockedAddresses, i)
	asset.BlockedAddresses = blockedAddrs
	k.SetAsset(ctx, asset)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnblock,
			sdk.NewAttribute(types.AttributeKeyUnblock, fmt.Sprintf("%s", addr)),
			sdk.NewAttribute(types.AttributeKeyDenom, asset.Denom),
		),
	)
	return nil
}

// ChangePauseStatus pauses/un-pauses an asset
func (k Keeper) ChangePauseStatus(ctx sdk.Context, owner sdk.AccAddress, denom string, status bool) error {
	asset, found := k.GetAsset(ctx, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrAssetNotFound, "denom: %s", denom)
	}
	if !owner.Equals(asset.Owner) {
		return sdkerrors.Wrapf(types.ErrNotAuthorized, "owner: %s, address: %s", asset.Owner, owner)
	}
	if asset.Paused == status {
		return nil
	}
	asset.Paused = status
	k.SetAsset(ctx, asset)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePause,
			sdk.NewAttribute(types.AttributeKeyPauseStatus, fmt.Sprintf("%t", status)),
			sdk.NewAttribute(types.AttributeKeyDenom, asset.Denom),
		),
	)
	return nil
}

// SeizeCoinsFromBlockedAddresses checks blocked addresses for coins of the input denom and transfers them to the owner account
func (k Keeper) SeizeCoinsFromBlockedAddresses(ctx sdk.Context, denom string) error {
	asset, found := k.GetAsset(ctx, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrAssetNotFound, "denom: %s", denom)
	}
	for _, address := range asset.BlockedAddresses {
		account := k.accountKeeper.GetAccount(ctx, address)
		coinsAmount := account.GetCoins().AmountOf(denom)
		if coinsAmount.IsPositive() {
			coins := sdk.NewCoins(sdk.NewCoin(denom, coinsAmount))
			err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, address, types.ModuleAccountName, coins)
			if err != nil {
				return err
			}
			err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, asset.Owner, coins)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k Keeper) checkBlockedAddress(ctx sdk.Context, asset types.Asset, checkAddress sdk.AccAddress) (bool, int) {
	for i, address := range asset.BlockedAddresses {
		if address.Equals(checkAddress) {
			return true, i
		}
	}
	return false, 0
}

func (k Keeper) removeBlockedAddress(ctx sdk.Context, blockedAddrs []sdk.AccAddress, i int) []sdk.AccAddress {
	blockedAddrs[len(blockedAddrs)-1], blockedAddrs[i] = blockedAddrs[i], blockedAddrs[len(blockedAddrs)-1]
	return blockedAddrs[:len(blockedAddrs)-1]
}
