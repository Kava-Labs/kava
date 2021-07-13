package keeper

import (
	"fmt"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SwapExactForTokens swaps an exact coin a input for a coin b output
func (k *Keeper) SwapExactForTokens(ctx sdk.Context, requester sdk.AccAddress, exactCoinA, coinB sdk.Coin, slippageLimit sdk.Dec) error {
	poolID, pool, err := k.loadPool(ctx, exactCoinA.Denom, coinB.Denom)
	if err != nil {
		return err
	}

	swapOutput, feePaid := pool.SwapWithExactInput(exactCoinA, k.GetSwapFee(ctx))
	if swapOutput.IsZero() {
		return sdkerrors.Wrapf(types.ErrInsufficientLiquidity, "swap output rounds to zero, increase input amount")
	}

	priceChange := swapOutput.Amount.ToDec().Quo(coinB.Amount.ToDec())
	if err := k.assertSlippageWithinLimit(priceChange, slippageLimit); err != nil {
		return err
	}

	if err := k.commitSwap(ctx, poolID, pool, requester, exactCoinA, swapOutput, feePaid, "input"); err != nil {
		return err
	}

	return nil
}

// SwapForExactTokens swaps a coin a input for an exact coin b output
func (k *Keeper) SwapForExactTokens(ctx sdk.Context, requester sdk.AccAddress, coinA, exactCoinB sdk.Coin, slippageLimit sdk.Dec) error {
	poolID, pool, err := k.loadPool(ctx, coinA.Denom, exactCoinB.Denom)
	if err != nil {
		return err
	}

	if exactCoinB.Amount.GTE(pool.Reserves().AmountOf(exactCoinB.Denom)) {
		return sdkerrors.Wrapf(
			types.ErrInsufficientLiquidity,
			"output %s >= pool reserves %s", exactCoinB.Amount.String(), pool.Reserves().AmountOf(exactCoinB.Denom).String(),
		)
	}

	swapInput, feePaid := pool.SwapWithExactOutput(exactCoinB, k.GetSwapFee(ctx))

	priceChange := coinA.Amount.ToDec().Quo(swapInput.Sub(feePaid).Amount.ToDec())
	if err := k.assertSlippageWithinLimit(priceChange, slippageLimit); err != nil {
		return err
	}

	if err := k.commitSwap(ctx, poolID, pool, requester, swapInput, exactCoinB, feePaid, "output"); err != nil {
		return err
	}

	return nil
}

func (k Keeper) loadPool(ctx sdk.Context, denomA string, denomB string) (string, *types.DenominatedPool, error) {
	poolID := types.PoolID(denomA, denomB)

	poolRecord, found := k.GetPool(ctx, poolID)
	if !found {
		return poolID, nil, sdkerrors.Wrapf(types.ErrInvalidPool, "pool %s not found", poolID)
	}

	pool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
	if err != nil {
		panic(fmt.Sprintf("invalid pool %s: %s", poolID, err))
	}

	return poolID, pool, nil
}

func (k Keeper) assertSlippageWithinLimit(priceChange sdk.Dec, slippageLimit sdk.Dec) error {
	slippage := sdk.OneDec().Sub(priceChange)
	if slippage.GT(slippageLimit) {
		return sdkerrors.Wrapf(types.ErrSlippageExceeded, "slippage %s > limit %s", slippage, slippageLimit)
	}

	return nil
}

func (k Keeper) commitSwap(
	ctx sdk.Context,
	poolID string,
	pool *types.DenominatedPool,
	requester sdk.AccAddress,
	swapInput sdk.Coin,
	swapOutput sdk.Coin,
	feePaid sdk.Coin,
	exactDirection string,
) error {
	k.SetPool(ctx, types.NewPoolRecord(pool))

	if err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, requester, types.ModuleAccountName, sdk.NewCoins(swapInput)); err != nil {
		return err
	}

	if err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, requester, sdk.NewCoins(swapOutput)); err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapTrade,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyRequester, requester.String()),
			sdk.NewAttribute(types.AttributeKeySwapInput, swapInput.String()),
			sdk.NewAttribute(types.AttributeKeySwapOutput, swapOutput.String()),
			sdk.NewAttribute(types.AttributeKeyFeePaid, feePaid.String()),
			sdk.NewAttribute(types.AttributeKeyExactDirection, exactDirection),
		),
	)

	return nil
}
