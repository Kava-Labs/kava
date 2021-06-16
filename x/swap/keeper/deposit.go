package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/swap/types"
)

func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, coinA sdk.Coin, coinB sdk.Coin) error {
	depositAmount := sdk.NewCoins(coinA, coinB)
	poolID := types.PoolIDFromCoins(depositAmount)

	_, found := k.GetPool(ctx, poolID)
	if found {
		//		//depositAmount, shares := pool.AddLiquidty(depositAmount)
		//		//if depositAmount.IsZero() || shares.IsZero() {
		return sdkerrors.Wrap(types.ErrNotImplemented, fmt.Sprintf("can not deposit into existing pool '%s'", poolID))
		//		//}
		//
		//		//desiredPrice := types.PriceFromPair(coinA, coinB)
		//		//actualPrice := types.PriceFromCoins(coinA.Denom, depositAmount)
		//
		//		//if actualPrice.Sub(desiredPrice).Quo(desiredPrice).Abs().GT(slippage) {
		//		//	// TODO: slippage error!
		//		//}
		//
		//		//k.SetPool(pool)
		//		//k.SetDepositorShares(meow)
	} else {
		if allowed := k.depositAllowed(ctx, poolID); !allowed {
			return sdkerrors.Wrap(types.ErrNotAllowed, fmt.Sprintf("can not create pool '%s'", poolID))
		}

		k.initializePool(ctx, depositor, depositAmount)
	}

	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, depositAmount)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapDeposit,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
		),
	)

	return nil
}

func (k Keeper) depositAllowed(ctx sdk.Context, poolID string) bool {
	params := k.GetParams(ctx)
	for _, p := range params.AllowedPools {
		if poolID == types.PoolID(p.TokenA, p.TokenB) {
			return true
		}
	}
	return false
}

func (k Keeper) initializePool(ctx sdk.Context, depositor sdk.AccAddress, reserves sdk.Coins) error {
	pool, err := types.NewDenominatedPool(reserves)
	if err != nil {
		return err
	}

	poolRecord := types.NewPoolRecord(pool)
	shareRecord := types.NewShareRecord(depositor, poolRecord.PoolID, pool.TotalShares())

	k.SetPool(ctx, poolRecord)
	k.SetDepositorShares(ctx, shareRecord)

	return nil
}
