package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsSendEnabledCoins uses the parent x/bank keeper to check the coins provided
// and returns an ErrSendDisabled if any of the coins are not configured for
// sending. Returns nil if sending is enabled for all provided coin
func (k Keeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	panic("unimplemented")
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
	panic("unimplemented")
}
