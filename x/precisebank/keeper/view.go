package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	panic("unimplemented")
}

func (k Keeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	panic("unimplemented")
}

func (k Keeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	panic("unimplemented")
}
