package keeper

import (
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers the swap module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "pool-records", PoolRecordsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "share-records", ShareRecordsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "pool-reserves", PoolReservesInvariant(k))
	ir.RegisterRoute(types.ModuleName, "pool-shares", PoolSharesInvariant(k))
}

// AllInvariants runs all invariants of the swap module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if res, stop := PoolRecordsInvariant(k)(ctx); stop {
			return res, stop
		}

		if res, stop := ShareRecordsInvariant(k)(ctx); stop {
			return res, stop
		}

		if res, stop := PoolReservesInvariant(k)(ctx); stop {
			return res, stop
		}

		res, stop := PoolSharesInvariant(k)(ctx)
		return res, stop
	}
}

// PoolRecordsInvariant iterates all pool records and asserts that they are valid
func PoolRecordsInvariant(k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "validate pool records broken", "pool record invalid")

	return func(ctx sdk.Context) (string, bool) {
		k.IteratePools(ctx, func(record types.PoolRecord) bool {
			if err := record.Validate(); err != nil {
				broken = true
				return true
			}
			return false
		})

		return message, broken
	}
}

// ShareRecordsInvariant iterates all pool records and asserts that they are valid
func ShareRecordsInvariant(k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "validate share records broken", "share record invalid")

	return func(ctx sdk.Context) (string, bool) {
		k.IterateDepositorShares(ctx, func(record types.ShareRecord) bool {
			if err := record.Validate(); err != nil {
				broken = true
				return true
			}
			return false
		})

		return message, broken
	}
}

// PoolReservesInvariant iterates all pools and ensures the total reserves matches the module account coins
func PoolReservesInvariant(k Keeper) sdk.Invariant {
	message := sdk.FormatInvariant(types.ModuleName, "pool reserves broken", "pool reserves do not match module account")

	return func(ctx sdk.Context) (string, bool) {
		mAcc := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)

		reserves := sdk.Coins{}
		k.IteratePools(ctx, func(record types.PoolRecord) bool {
			for _, coin := range record.Reserves() {
				reserves = reserves.Add(coin)
			}
			return false
		})

		broken := !reserves.IsEqual(mAcc.GetCoins())
		return message, broken
	}
}

type poolShares struct {
	totalShares      sdk.Int
	totalSharesOwned sdk.Int
}

// PoolSharesInvariant iterates all pools and shares and ensures the total pool shares match the sum of depositor shares
func PoolSharesInvariant(k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "pool shares broken", "pool shares do not match depositor shares")

	return func(ctx sdk.Context) (string, bool) {
		totalShares := make(map[string]poolShares)

		k.IteratePools(ctx, func(pr types.PoolRecord) bool {
			totalShares[pr.PoolID] = poolShares{
				totalShares:      pr.TotalShares,
				totalSharesOwned: sdk.ZeroInt(),
			}

			return false
		})

		k.IterateDepositorShares(ctx, func(sr types.ShareRecord) bool {
			if shares, found := totalShares[sr.PoolID]; found {
				shares.totalSharesOwned = shares.totalSharesOwned.Add(sr.SharesOwned)
				totalShares[sr.PoolID] = shares
			} else {
				totalShares[sr.PoolID] = poolShares{
					totalShares:      sdk.ZeroInt(),
					totalSharesOwned: sr.SharesOwned,
				}
			}

			return false
		})

		for _, ps := range totalShares {
			if !ps.totalShares.Equal(ps.totalSharesOwned) {
				broken = true
				break
			}
		}

		return message, broken
	}
}
