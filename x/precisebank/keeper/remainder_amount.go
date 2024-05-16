package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank/types"
)

// GetRemainderAmount returns the internal remainder amount.
func (k *Keeper) GetRemainderAmount(
	ctx sdk.Context,
) sdkmath.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.RemainderBalanceKey)
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	var bal sdkmath.Int
	if err := bal.Unmarshal(bz); err != nil {
		panic(fmt.Errorf("failed to unmarshal remainder amount: %w", err))
	}

	return bal
}

// SetRemainderAmount sets the internal remainder amount.
func (k *Keeper) SetRemainderAmount(
	ctx sdk.Context,
	amount sdkmath.Int,
) {
	// Prevent storing zero amounts. In practice, the remainder amount should
	// only be non-zero during transactions as mint and burns should net zero
	// due to only being used for EVM transfers.
	if amount.IsZero() {
		k.DeleteRemainderAmount(ctx)
		return
	}

	// Ensure the remainder is valid before setting it. Follows the same
	// validation as FractionalBalance with the same value range.
	if err := types.NewFractionalAmountFromInt(amount).Validate(); err != nil {
		panic(fmt.Errorf("remainder amount is invalid: %w", err))
	}

	store := ctx.KVStore(k.storeKey)

	amountBytes, err := amount.Marshal()
	if err != nil {
		panic(fmt.Errorf("failed to marshal remainder amount: %w", err))
	}

	store.Set(types.RemainderBalanceKey, amountBytes)
}

// DeleteRemainderAmount deletes the internal remainder amount.
func (k *Keeper) DeleteRemainderAmount(
	ctx sdk.Context,
) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.RemainderBalanceKey)
}
