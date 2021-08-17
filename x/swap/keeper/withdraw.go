package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/swap/types"
)

// Withdraw removes liquidity from an existing pool from an owners deposit, converting the provided shares for
// the returned pool liquidity.
//
// If 100% of the owners shares are removed, then the deposit is deleted.  In addition, if all the pool shares
// are removed then the pool is deleted.
//
// The number of shares must be large enough to result in at least 1 unit of the smallest reserve in the pool.
// If the share input is below the minimum required for positive liquidity to be remove from both reserves, a
// insufficient error is returned.
//
// In addition, if the withdrawn liquidity for each reserve is below the provided minimum, a slippage exceeded
// error is returned.
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
		return sdkerrors.Wrap(types.ErrPoolNotFound, poolID)
	}

	pool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidPool, "%s: %s", poolID, err)
	}

	withdrawnAmount := pool.RemoveLiquidity(shares)
	if withdrawnAmount.AmountOf(minCoinA.Denom).IsZero() || withdrawnAmount.AmountOf(minCoinB.Denom).IsZero() {
		return sdkerrors.Wrap(types.ErrInsufficientLiquidity, "shares must be increased")
	}
	if withdrawnAmount.AmountOf(minCoinA.Denom).LT(minCoinA.Amount) || withdrawnAmount.AmountOf(minCoinB.Denom).LT(minCoinB.Amount) {
		return sdkerrors.Wrap(types.ErrSlippageExceeded, "minimum withdraw not met")
	}

	k.updatePool(ctx, poolID, pool)
	k.BeforePoolDepositModified(ctx, poolID, owner, shareRecord.SharesOwned)
	k.updateDepositorShares(ctx, owner, poolID, shareRecord.SharesOwned.Sub(shares))

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, owner, withdrawnAmount)
	if err != nil {
		return err
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
