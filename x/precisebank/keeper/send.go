package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IsSendEnabledCoins uses the parent x/bank keeper to check the coins provided
// and returns an ErrSendDisabled if any of the coins are not configured for
// sending. Returns nil if sending is enabled for all provided coin
func (k Keeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	// TODO: This does not actually seem to be used by x/evm, so it should be
	// removed from the expected_interface in x/evm.
	return k.bk.IsSendEnabledCoins(ctx, coins...)
}

func (k Keeper) SendCoins(
	ctx sdk.Context,
	from, to sdk.AccAddress,
	amt sdk.Coins,
) error {
	// IsSendEnabledCoins() is only used in x/bank in msg server, not in keeper,
	// so we should also not use it here to align with x/bank behavior.

	panic("unimplemented")
}

func (k Keeper) SendCoinsFromAccountToModule(
	ctx sdk.Context,
	senderAddr sdk.AccAddress,
	recipientModule string,
	amt sdk.Coins,
) error {
	panic("unimplemented")
}

func (k Keeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context,
	senderModule string,
	recipientAddr sdk.AccAddress,
	amt sdk.Coins,
) error {
	// Identical panics to x/bank
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	// Uses x/bank BlockedAddr, no need to modify
	if k.bk.BlockedAddr(recipientAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
	}

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}
