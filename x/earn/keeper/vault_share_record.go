package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// ----------------------------------------------------------------------------
// VaultShareRecords -- user shares per vault

// GetVaultShareRecord returns the vault share record for a given denom and
// account.
func (k *Keeper) GetVaultShareRecord(
	ctx sdk.Context,
	acc sdk.AccAddress,
) (types.VaultShareRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultShareRecordKeyPrefix)

	bz := store.Get(types.DepositorVaultSharesKey(acc))
	if bz == nil {
		return types.VaultShareRecord{}, false
	}

	var record types.VaultShareRecord
	k.cdc.MustUnmarshal(bz, &record)

	return record, true
}

// UpdateVaultShareRecord updates the vault share record in state for a given
// denom and account. This deletes it if the supply is zero and updates the
// state if supply is non-zero.
func (k *Keeper) UpdateVaultShareRecord(
	ctx sdk.Context,
	record types.VaultShareRecord,
) {
	if record.Shares.IsZero() {
		k.DeleteVaultShareRecord(ctx, record.Depositor)
	} else {
		k.SetVaultShareRecord(ctx, record)
	}
}

// DeleteVaultShareRecord deletes the vault share record for a given account.
func (k *Keeper) DeleteVaultShareRecord(
	ctx sdk.Context,
	acc sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultShareRecordKeyPrefix)
	store.Delete(types.DepositorVaultSharesKey(acc))
}

// SetVaultShareRecord sets the vault share record for a given account.
func (k *Keeper) SetVaultShareRecord(
	ctx sdk.Context,
	record types.VaultShareRecord,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultShareRecordKeyPrefix)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.DepositorVaultSharesKey(record.Depositor), bz)
}

// IterateVaultShareRecords iterates over all vault share objects in the store
// and performs a callback function.
func (k Keeper) IterateVaultShareRecords(
	ctx sdk.Context,
	cb func(record types.VaultShareRecord) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultShareRecordKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.VaultShareRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// GetAllVaultShareRecords returns all vault share records from the store.
func (k Keeper) GetAllVaultShareRecords(ctx sdk.Context) types.VaultShareRecords {
	var records types.VaultShareRecords

	k.IterateVaultShareRecords(ctx, func(record types.VaultShareRecord) bool {
		records = append(records, record)
		return false
	})

	return records
}
