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

	// TODO: extract method
	params := k.GetParams(ctx)
	creationAllowed := false
	for _, p := range params.AllowedPools {
		if p.TokenA == amountA.Denom && p.TokenB == amountB.Denom {
			creationAllowed = true
		}
	}
	if !creationAllowed {
		return sdkerrors.Wrap(types.ErrNotAllowed, fmt.Sprintf("can not create pool '%s'", poolName))
	}

	// TODO: extract method, wrap error
	amount := sdk.NewCoins(amountA, amountB)
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, amount)
	if err != nil {
		return err
	}

	// TODO: extra method
	pool := types.NewPool(amountA, amountB)
	k.SetPool(ctx, pool)
	k.SetDepositorShares(ctx, depositor, pool.Name(), pool.TotalShares)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapDeposit,
			sdk.NewAttribute(types.AttributeKeyPoolName, pool.Name()),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}
