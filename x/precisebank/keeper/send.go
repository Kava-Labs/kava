package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	panic("unimplemented")
}

func (k Keeper) SendCoins(
	ctx sdk.Context,
	from, to sdk.AccAddress,
	amt sdk.Coins,
) error {
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
