package keeper

import (
	//"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/swap/types"
)

// Deposit creates a new pool or adds liquidity to an existing pool.  For a pool to be created, a pool
// for the coin denominations must not exist yet, and it must be allowed by the swap module parameters.
//
// When adding liquidity to an existing pool, the provided coins are consided to be the desired deposit
// amount, and the actual deposited coins may be less than or equal to the provided coins.  A deposit
// will never be exceed the coinA and coinB amounts.
func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, coinA sdk.Coin, coinB sdk.Coin) error {
	desiredAmount := sdk.NewCoins(coinA, coinB)

	poolID := types.PoolIDFromCoins(desiredAmount)
	poolRecord, found := k.GetPool(ctx, poolID)

	var (
		depositAmount sdk.Coins
		shares        sdk.Int
		err           error
	)
	if found {
		depositAmount, shares, err = k.addLiquidityToPool(ctx, poolRecord, depositor, desiredAmount)
	} else {
		depositAmount, shares, err = k.initializePool(ctx, poolID, depositor, desiredAmount)
	}
	if err != nil {
		return err
	}

	err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleAccountName, depositAmount)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapDeposit,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
			sdk.NewAttribute(types.AttributeKeyShares, shares.String()),
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

func (k Keeper) initializePool(ctx sdk.Context, poolID string, depositor sdk.AccAddress, reserves sdk.Coins) (sdk.Coins, sdk.Int, error) {
	if allowed := k.depositAllowed(ctx, poolID); !allowed {
		return sdk.Coins{}, sdk.ZeroInt(), sdkerrors.Wrap(types.ErrNotAllowed, fmt.Sprintf("can not create pool '%s'", poolID))
	}

	pool, err := types.NewDenominatedPool(reserves)
	if err != nil {
		return sdk.Coins{}, sdk.ZeroInt(), err
	}

	poolRecord := types.NewPoolRecord(pool)
	shareRecord := types.NewShareRecord(depositor, poolRecord.PoolID, pool.TotalShares())

	k.SetPool(ctx, poolRecord)
	k.SetDepositorShares(ctx, shareRecord)

	return pool.Reserves(), pool.TotalShares(), nil
}

func (k Keeper) addLiquidityToPool(ctx sdk.Context, record types.PoolRecord, depositor sdk.AccAddress, desiredAmount sdk.Coins) (sdk.Coins, sdk.Int, error) {
	pool, err := types.NewDenominatedPoolWithExistingShares(record.Reserves(), record.TotalShares)
	if err != nil {
		return sdk.Coins{}, sdk.ZeroInt(), err
	}

	depositAmount, shares := pool.AddLiquidity(desiredAmount)

	poolRecord := types.NewPoolRecord(pool)
	shareRecord := types.NewShareRecord(depositor, poolRecord.PoolID, shares)

	k.SetPool(ctx, poolRecord)
	k.SetDepositorShares(ctx, shareRecord)

	return depositAmount, shares, nil
}
