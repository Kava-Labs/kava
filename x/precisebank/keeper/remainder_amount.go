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
) (sdkmath.Int, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.RemainderBalanceKey)
	if bz == nil {
		return sdkmath.ZeroInt(), false
	}

	var bal sdkmath.Int
	if err := bal.Unmarshal(bz); err != nil {
		panic(fmt.Errorf("failed to unmarshal fractional balance: %w", err))
	}

	return bal, true
}

// SetRemainderAmount sets the internal remainder amount.
func (k *Keeper) SetRemainderAmount(
	ctx sdk.Context,
	amount sdkmath.Int,
) {
	store := ctx.KVStore(k.storeKey)

	amountBytes, err := amount.Marshal()
	if err != nil {
		panic(fmt.Errorf("failed to marshal fractional balance: %w", err))
	}

	store.Set(types.RemainderBalanceKey, amountBytes)
}
