package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/swap/types"
)

func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, amountA sdk.Coin, amountB sdk.Coin) error {
	poolName := types.PoolName(amountA.Denom, amountB.Denom)

	_, found := k.GetPool(ctx, poolName)
	if found {
		return sdkerrors.Wrap(types.ErrNotImplemented, fmt.Sprintf("can not deposit into existing pool '%s'", poolName))
	}

	err := k.DepositAllowed(ctx, poolName, amountA, amountB)
	if err != nil {
		return err
	}

	// TODO: extract method, wrap error
	amount := sdk.NewCoins(amountA, amountB)
	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, amount)
	if err != nil {
		return err
	}

	k.InitializePool(ctx, depositor, amountA, amountB)

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

func (k Keeper) InitializePool(ctx sdk.Context, depositor sdk.AccAddress, amountA, amountB sdk.Coin) error {
	pool, err := types.NewPool(amountA, amountB)
	if err != nil {
		return err
	}
	k.SetPool(ctx, pool)
	k.SetDepositorShares(ctx, depositor, pool.Name(), pool.TotalShares)
	return nil
}

func (k Keeper) DepositAllowed(ctx sdk.Context, poolName string, amountA, amountB sdk.Coin) error {
	params := k.GetParams(ctx)
	for _, p := range params.AllowedPools {
		if p.TokenA == amountA.Denom && p.TokenB == amountB.Denom {
			return nil
		}
	}
	return sdkerrors.Wrap(types.ErrNotAllowed, fmt.Sprintf("can not create pool '%s'", poolName))
}
