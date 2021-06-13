package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/swap/types"
)

func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, coinA sdk.Coin, coinB sdk.Coin) error {
	poolName := types.PoolName(coinA.Denom, coinB.Denom)

	_, found := k.GetPool(ctx, poolName)
	if found {
		return sdkerrors.Wrap(types.ErrNotImplemented, fmt.Sprintf("can not deposit into existing pool '%s'", poolName))
	}

	err := k.depositAllowed(ctx, poolName, coinA, coinB)
	if err != nil {
		return err
	}

	// TODO: extract method, wrap error
	amount := sdk.NewCoins(coinA, coinB)
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, amount)
	if err != nil {
		return err
	}

	k.initializePool(ctx, depositor, coinA, coinB)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapDeposit,
			sdk.NewAttribute(types.AttributeKeyPoolName, poolName),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}

func (k Keeper) depositAllowed(ctx sdk.Context, poolName string, coinA, coinB sdk.Coin) error {
	params := k.GetParams(ctx)
	for _, p := range params.AllowedPools {
		if p.TokenA == coinA.Denom && p.TokenB == coinB.Denom {
			return nil
		}
	}
	return sdkerrors.Wrap(types.ErrNotAllowed, fmt.Sprintf("can not create pool '%s'", poolName))
}

func (k Keeper) initializePool(ctx sdk.Context, depositor sdk.AccAddress, coinA, coinB sdk.Coin) error {
	pool, err := types.NewPool(coinA, coinB)
	if err != nil {
		return err
	}
	k.SetPool(ctx, pool)
	k.SetDepositorShares(ctx, depositor, pool.Name(), pool.TotalShares)
	return nil
}
