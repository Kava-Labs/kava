package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/swap/types"
)

func (k Keeper) Withdraw(ctx sdk.Context, owner sdk.AccAddress, poolID string, withdrawShares sdk.Int) error {
	// Confirm that the depositor owns the requested shares to withdraw
	depositorShareRecord, found := k.GetDepositorShares(ctx, owner, poolID)
	if !found {
		return sdkerrors.Wrapf(types.ErrShareRecordNotFound, "share record of %s for pool %s not found", owner, poolID)
	}

	if withdrawShares.GT(depositorShareRecord.SharesOwned) {
		return sdkerrors.Wrapf(types.ErrInvalidShares,
			"requested shares to withdraw %s is greater than total amount of shares owned by requester %s",
			withdrawShares,
			depositorShareRecord.SharesOwned,
		)
	}

	denominatedPool, err := k.loadDenominatedPool(ctx, poolID)
	if err != nil {
		return err
	}
	withdrawCoins := denominatedPool.RemoveLiquidity(withdrawShares)

	// Send withdrawn tokens to owner
	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, owner, withdrawCoins)
	if err != nil {
		return err
	}

	// Update pool record
	if denominatedPool.IsEmpty() {
		k.DeletePool(ctx, poolID)
	} else {
		poolRecord := types.NewPoolRecord(denominatedPool)
		k.SetPool(ctx, poolRecord)
	}

	// Update depositor's share record
	depositorShareRecord.SharesOwned = depositorShareRecord.SharesOwned.Sub(withdrawShares)
	k.SetDepositorShares(ctx, depositorShareRecord)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapWithdraw,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, withdrawCoins.String()),
			sdk.NewAttribute(types.AttributeKeyShares, withdrawShares.String()),
		),
	)

	return nil
}

func (k Keeper) loadDenominatedPool(ctx sdk.Context, poolID string) (*types.DenominatedPool, error) {
	pool, found := k.GetPool(ctx, poolID)
	if !found {
		return &types.DenominatedPool{}, types.ErrInvalidPool
	}
	denominatedPool, err := types.NewDenominatedPool(pool.Reserves())
	if err != nil {
		return &types.DenominatedPool{}, types.ErrInvalidPool
	}
	return denominatedPool, nil
}
