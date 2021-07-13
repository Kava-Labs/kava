package keeper

import (
	"fmt"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SwapExactForTokens swaps an exact coin a input for a coin b output
func (k *Keeper) SwapExactForTokens(ctx sdk.Context, requester sdk.AccAddress, exactCoinA, coinB sdk.Coin, slippageLimit sdk.Dec) error {
	poolID := types.PoolID(exactCoinA.Denom, coinB.Denom)

	poolRecord, found := k.GetPool(ctx, poolID)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidPool, "pool %s not found", poolID)
	}

	pool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
	if err != nil {
		panic(fmt.Sprintf("invalid pool %s: %s", poolID, err))
	}

	swapOutput, feePaid := pool.SwapWithExactInput(exactCoinA, k.GetSwapFee(ctx))
	if swapOutput.IsZero() {
		return sdkerrors.Wrapf(types.ErrInsufficientLiquidity, "increase input amount")
	}

	priceChange := swapOutput.Amount.ToDec().Quo(coinB.Amount.ToDec())
	slippage := sdk.OneDec().Sub(priceChange)
	if slippage.GT(slippageLimit) {
		return sdkerrors.Wrapf(types.ErrSlippageExceeded, "slippage %s > limit %s", slippage, slippageLimit)
	}

	k.SetPool(ctx, types.NewPoolRecord(pool))

	if err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, requester, types.ModuleAccountName, sdk.NewCoins(exactCoinA)); err != nil {
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
			sdk.NewAttribute(types.AttributeKeySwapInput, exactCoinA.String()),
			sdk.NewAttribute(types.AttributeKeySwapOutput, swapOutput.String()),
			sdk.NewAttribute(types.AttributeKeyFeePaid, feePaid.String()),
			sdk.NewAttribute(types.AttributeKeyExactDirection, "input"),
		),
	)

	return nil
}

// SwapExactForTokens swaps a coin a input for an exact coin b output
func (k *Keeper) SwapForExactTokens(ctx sdk.Context, requester sdk.AccAddress, coinA, exactCoinB sdk.Coin, slippageLimit sdk.Dec) error {
	poolID := types.PoolID(coinA.Denom, exactCoinB.Denom)

	poolRecord, found := k.GetPool(ctx, poolID)
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidPool, "pool %s not found", poolID)
	}

	pool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
	if err != nil {
		panic(fmt.Sprintf("invalid pool %s: %s", poolID, err))
	}

	if exactCoinB.Amount.GTE(pool.Reserves().AmountOf(exactCoinB.Denom)) {
		return sdkerrors.Wrapf(
			types.ErrInsufficientLiquidity,
			"output %s >= pool reserves %s", exactCoinB.Amount.String(), pool.Reserves().AmountOf(exactCoinB.Denom).String(),
		)
	}

	swapInput, feePaid := pool.SwapWithExactOutput(exactCoinB, k.GetSwapFee(ctx))

	priceChange := coinA.Amount.ToDec().Quo(swapInput.Sub(feePaid).Amount.ToDec())
	slippage := sdk.OneDec().Sub(priceChange)
	if slippage.GT(slippageLimit) {
		return sdkerrors.Wrapf(types.ErrSlippageExceeded, "slippage %s > limit %s", slippage, slippageLimit)
	}

	k.SetPool(ctx, types.NewPoolRecord(pool))

	if err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, requester, types.ModuleAccountName, sdk.NewCoins(swapInput)); err != nil {
		return err
	}

	if err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, requester, sdk.NewCoins(exactCoinB)); err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapTrade,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyRequester, requester.String()),
			sdk.NewAttribute(types.AttributeKeySwapInput, swapInput.String()),
			sdk.NewAttribute(types.AttributeKeySwapOutput, exactCoinB.String()),
			sdk.NewAttribute(types.AttributeKeyFeePaid, feePaid.String()),
			sdk.NewAttribute(types.AttributeKeyExactDirection, "output"),
		),
	)

	return nil
}
