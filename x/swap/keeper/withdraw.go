package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/swap/types"
)

// Withdraw removes liquidity from an existing pool, converting the provided shares for
// the returned pool liquidity.
func (k Keeper) Withdraw(ctx sdk.Context, owner sdk.AccAddress, shares sdk.Int, minCoinA, minCoinB sdk.Coin) error {
	poolID := types.PoolID(minCoinA.Denom, minCoinB.Denom)

	shareRecord, found := k.GetDepositorShares(ctx, owner, poolID)
	if !found {
		return sdkerrors.Wrapf(types.ErrDepositNotFound, "no deposit for account %s and pool %s", owner, poolID)
	}

	if shares.GT(shareRecord.SharesOwned) {
		return sdkerrors.Wrapf(types.ErrInvalidShares, "withdraw of %s shares greater than %s shares owned", shares, shareRecord.SharesOwned)
	}

	poolRecord, found := k.GetPool(ctx, poolID)
	if !found {
		panic(fmt.Sprintf("pool %s not found", poolID))
	}

	pool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
	if err != nil {
		panic(fmt.Sprintf("invalid pool %s: %s", poolID, err))
	}

	withdrawnAmount := pool.RemoveLiquidity(shares)
	if withdrawnAmount.AmountOf(minCoinA.Denom).IsZero() || withdrawnAmount.AmountOf(minCoinB.Denom).IsZero() {
		return sdkerrors.Wrap(types.ErrInsufficientLiquidity, "shares must be increased")
	}
	if withdrawnAmount.AmountOf(minCoinA.Denom).LT(minCoinA.Amount) || withdrawnAmount.AmountOf(minCoinB.Denom).LT(minCoinB.Amount) {
		return sdkerrors.Wrap(types.ErrSlippageExceeded, "minimum withdraw not met")
	}

	if pool.IsEmpty() {
		k.DeletePool(ctx, poolID)
	} else {
		poolRecord = types.NewPoolRecord(pool)
		k.SetPool(ctx, poolRecord)
	}

	shareRecord.SharesOwned = shareRecord.SharesOwned.Sub(shares)
	if shareRecord.SharesOwned.IsZero() {
		k.DeleteDepositorShares(ctx, owner, poolID)
	} else {
		k.SetDepositorShares(ctx, shareRecord)
	}

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, owner, withdrawnAmount)
	if err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapWithdraw,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, withdrawnAmount.String()),
			sdk.NewAttribute(types.AttributeKeyShares, shares.String()),
		),
	)

	return nil
}
