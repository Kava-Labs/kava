package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/swap/types"
)

// Deposit creates a new pool or adds liquidity to an existing pool.  For a pool to be created, a pool
// for the coin denominations must not exist yet, and it must be allowed by the swap module parameters.
//
// When adding liquidity to an existing pool, the provided coins are considered to be the desired deposit
// amount, and the actual deposited coins may be less than or equal to the provided coins.  A deposit
// will never be exceed the coinA and coinB amounts.
//
// The slippage is calculated using both the price and inverse price of the provided coinA and coinB.
// Since adding liquidity is not directional, like a swap would be, using both the price (coinB/coinA),
// and the inverse price (coinA/coinB), protects the depositor from a large deviation in their deposit.
//
// The amount deposited may only change by B' < B or A' < A -- either B depreciates, or A depreciates.
// Therefore, slippage can be written as a function of this depreciation d.  Where the new price is
// B*(1-d)/A or A*(1-d)/B, and the inverse of each, and is A/(B*(1-d)) and B/(A*(1-d))
// respectively.
//
// Since 1/(1-d) >= (1-d) for d <= 1, the maximum slippage is always in the appreciating price
// A/(B*(1-d)) and B/(A*(1-d)).  In other words, when the price of an asset depreciates, the
// inverse price -- or the price of the other pool asset, appreciates by a larger amount.
// It's this percent change we calculate and compare to the slippage limit provided.
//
// For example, if we have a pool with 100e6 ukava and 400e6 usdx.  The ukava price is 4 usdx and the
// usdx price is 0.25 ukava.  If a depositor adds liquidity of 4e6 ukava and 14e6 usdx, a kava price of
// 3.50 usdx and a usdx price of 0.29 ukava.  This is a -12.5% slippage is the ukava price, and a 14.3%
// slippage in the usdx price.
//
// These slippages can be calculated by S_B = ((A/B')/(A/B) - 1) and S_A ((B/A')/(B/A) - 1), simplifying to
// S_B = (A/A' - 1), and S_B = (B/B' - 1).  An error is returned when max(S_A, S_B) > slippageLimit.
func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, coinA sdk.Coin, coinB sdk.Coin, slippageLimit sdk.Dec) error {
	desiredAmount := sdk.NewCoins(coinA, coinB)

	poolID := types.PoolIDFromCoins(desiredAmount)
	poolRecord, found := k.GetPool(ctx, poolID)

	var (
		pool          *types.DenominatedPool
		depositAmount sdk.Coins
		shares        sdk.Int
		err           error
	)
	if found {
		pool, depositAmount, shares, err = k.addLiquidityToPool(ctx, poolRecord, depositor, desiredAmount)
	} else {
		pool, depositAmount, shares, err = k.initializePool(ctx, poolID, depositor, desiredAmount)
	}
	if err != nil {
		return err
	}

	if depositAmount.AmountOf(coinA.Denom).IsZero() || depositAmount.AmountOf(coinB.Denom).IsZero() {
		return sdkerrors.Wrap(types.ErrInsufficientLiquidity, "deposit must be increased")
	}

	if shares.IsZero() {
		return sdkerrors.Wrap(types.ErrInsufficientLiquidity, "deposit must be increased")
	}

	maxPercentPriceChange := sdk.MaxDec(
		desiredAmount.AmountOf(coinA.Denom).ToDec().Quo(depositAmount.AmountOf(coinA.Denom).ToDec()),
		desiredAmount.AmountOf(coinB.Denom).ToDec().Quo(depositAmount.AmountOf(coinB.Denom).ToDec()),
	)
	slippage := maxPercentPriceChange.Sub(sdk.OneDec())

	if slippage.GT(slippageLimit) {
		return sdkerrors.Wrapf(types.ErrSlippageExceeded, "slippage %s > limit %s", slippage, slippageLimit)
	}

	k.updatePool(ctx, poolID, pool)
	if shareRecord, hasExistingShares := k.GetDepositorShares(ctx, depositor, poolID); hasExistingShares {
		k.BeforePoolDepositModified(ctx, poolID, depositor, shareRecord.SharesOwned)
		k.updateShares(ctx, depositor, poolID, shareRecord.SharesOwned.Add(shares))
	} else {
		k.updateShares(ctx, depositor, poolID, shares)
		k.AfterPoolDepositCreated(ctx, poolID, depositor, shares)
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

func (k Keeper) initializePool(ctx sdk.Context, poolID string, depositor sdk.AccAddress, reserves sdk.Coins) (*types.DenominatedPool, sdk.Coins, sdk.Int, error) {
	if allowed := k.depositAllowed(ctx, poolID); !allowed {
		return nil, sdk.Coins{}, sdk.ZeroInt(), sdkerrors.Wrap(types.ErrNotAllowed, fmt.Sprintf("can not create pool '%s'", poolID))
	}

	pool, err := types.NewDenominatedPool(reserves)
	if err != nil {
		return nil, sdk.Coins{}, sdk.ZeroInt(), err
	}

	return pool, pool.Reserves(), pool.TotalShares(), nil
}

func (k Keeper) addLiquidityToPool(ctx sdk.Context, record types.PoolRecord, depositor sdk.AccAddress, desiredAmount sdk.Coins) (*types.DenominatedPool, sdk.Coins, sdk.Int, error) {
	pool, err := types.NewDenominatedPoolWithExistingShares(record.Reserves(), record.TotalShares)
	if err != nil {
		return nil, sdk.Coins{}, sdk.ZeroInt(), err
	}

	depositAmount, shares := pool.AddLiquidity(desiredAmount)

	return pool, depositAmount, shares, nil
}
