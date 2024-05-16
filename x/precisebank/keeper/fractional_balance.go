package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank/types"
)

// GetFractionalBalance returns the fractional balance for an address.
func (k *Keeper) GetFractionalBalance(
	ctx sdk.Context,
	address sdk.AccAddress,
) (sdkmath.Int, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FractionalBalancePrefix)

	bz := store.Get(types.FractionalBalanceKey(address))
	if bz == nil {
		return sdkmath.ZeroInt(), false
	}

	var bal sdkmath.Int
	if err := bal.Unmarshal(bz); err != nil {
		panic(fmt.Errorf("failed to unmarshal fractional balance: %w", err))
	}

	return bal, true
}

// SetFractionalBalance sets the fractional balance for an address.
func (k *Keeper) SetFractionalBalance(
	ctx sdk.Context,
	address sdk.AccAddress,
	amount sdkmath.Int,
) {
	if address.Empty() {
		panic(errors.New("address cannot be empty"))
	}

	if amount.IsZero() {
		k.DeleteFractionalBalance(ctx, address)
		return
	}

	// Ensure the fractional balance is valid before setting it. Use the
	// NewFractionalAmountFromInt wrapper to use its Validate() method.
	if err := types.NewFractionalAmountFromInt(amount).Validate(); err != nil {
		panic(fmt.Errorf("amount is invalid: %w", err))
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FractionalBalancePrefix)

	amountBytes, err := amount.Marshal()
	if err != nil {
		panic(fmt.Errorf("failed to marshal fractional balance: %w", err))
	}

	store.Set(types.FractionalBalanceKey(address), amountBytes)
}

// DeleteFractionalBalance deletes the fractional balance for an address.
func (k *Keeper) DeleteFractionalBalance(
	ctx sdk.Context,
	address sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FractionalBalancePrefix)
	store.Delete(types.FractionalBalanceKey(address))
}

// IterateFractionalBalances iterates over all fractional balances in the store
// and performs a callback function.
func (k *Keeper) IterateFractionalBalances(
	ctx sdk.Context,
	cb func(address sdk.AccAddress, amount sdkmath.Int) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FractionalBalancePrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address := sdk.AccAddress(iterator.Key())

		var amount sdkmath.Int
		if err := amount.Unmarshal(iterator.Value()); err != nil {
			panic(fmt.Errorf("failed to unmarshal fractional balance: %w", err))
		}

		if cb(address, amount) {
			break
		}
	}
}

// GetTotalSumFractionalBalances returns the sum of all fractional balances.
func (k *Keeper) GetTotalSumFractionalBalances(ctx sdk.Context) sdkmath.Int {
	sum := sdkmath.ZeroInt()

	k.IterateFractionalBalances(ctx, func(_ sdk.AccAddress, amount sdkmath.Int) bool {
		sum = sum.Add(amount)
		return false
	})

	return sum
}
