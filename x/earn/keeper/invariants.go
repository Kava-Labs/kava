package keeper

import (
	"github.com/kava-labs/kava/x/earn/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers the swap module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "vault-records", VaultRecordsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "share-records", ShareRecordsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "vault-shares", VaultSharesInvariant(k))
}

// AllInvariants runs all invariants of the swap module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if res, stop := VaultRecordsInvariant(k)(ctx); stop {
			return res, stop
		}

		if res, stop := ShareRecordsInvariant(k)(ctx); stop {
			return res, stop
		}

		res, stop := VaultSharesInvariant(k)(ctx)
		return res, stop
	}
}

// VaultRecordsInvariant iterates all vault records and asserts that they are valid
func VaultRecordsInvariant(k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "validate vault records broken", "vault record invalid")

	return func(ctx sdk.Context) (string, bool) {
		k.IterateVaultRecords(ctx, func(record types.VaultRecord) bool {
			if err := record.Validate(); err != nil {
				broken = true
				return true
			}
			return false
		})

		return message, broken
	}
}

// ShareRecordsInvariant iterates all share records and asserts that they are valid
func ShareRecordsInvariant(k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "validate share records broken", "share record invalid")

	return func(ctx sdk.Context) (string, bool) {
		k.IterateVaultShareRecords(ctx, func(record types.VaultShareRecord) bool {
			if err := record.Validate(); err != nil {
				broken = true
				return true
			}
			return false
		})

		return message, broken
	}
}

type vaultShares struct {
	totalShares      types.VaultShare
	totalSharesOwned types.VaultShare
}

// VaultSharesInvariant iterates all vaults and shares and ensures the total vault shares match the sum of depositor shares
func VaultSharesInvariant(k Keeper) sdk.Invariant {
	broken := false
	message := sdk.FormatInvariant(types.ModuleName, "vault shares broken", "vault shares do not match depositor shares")

	return func(ctx sdk.Context) (string, bool) {
		totalShares := make(map[string]vaultShares)

		k.IterateVaultRecords(ctx, func(record types.VaultRecord) bool {
			totalShares[record.TotalShares.Denom] = vaultShares{
				totalShares:      record.TotalShares,
				totalSharesOwned: types.NewVaultShare(record.TotalShares.Denom, sdk.ZeroDec()),
			}

			return false
		})

		k.IterateVaultShareRecords(ctx, func(sr types.VaultShareRecord) bool {
			for _, share := range sr.Shares {
				if shares, found := totalShares[share.Denom]; found {
					shares.totalSharesOwned = shares.totalSharesOwned.Add(share)
					totalShares[share.Denom] = shares
				} else {
					totalShares[share.Denom] = vaultShares{
						totalShares:      types.NewVaultShare(share.Denom, sdk.ZeroDec()),
						totalSharesOwned: share,
					}
				}
			}

			return false
		})

		for _, share := range totalShares {
			if !share.totalShares.Amount.Equal(share.totalSharesOwned.Amount) {
				broken = true
				break
			}
		}

		return message, broken
	}
}
