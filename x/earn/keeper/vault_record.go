package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// ----------------------------------------------------------------------------
// VaultRecord -- vault total shares

// GetVaultRecord returns the vault record for a given denom.
func (k *Keeper) GetVaultRecord(
	ctx sdk.Context,
	vaultDenom string,
) (types.VaultRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)

	bz := store.Get(types.VaultKey(vaultDenom))
	if bz == nil {
		return types.VaultRecord{}, false
	}

	var record types.VaultRecord
	k.cdc.MustUnmarshal(bz, &record)

	return record, true
}

// UpdateVaultRecord updates the vault record in state for a given denom. This
// deletes it if the supply is zero and updates the state if supply is non-zero.
func (k *Keeper) UpdateVaultRecord(
	ctx sdk.Context,
	vaultRecord types.VaultRecord,
) {
	if vaultRecord.TotalShares.Amount.IsZero() {
		k.DeleteVaultRecord(ctx, vaultRecord.TotalShares.Denom)
	} else {
		k.SetVaultRecord(ctx, vaultRecord)
	}
}

// DeleteVaultRecord deletes the vault record for a given denom.
func (k *Keeper) DeleteVaultRecord(ctx sdk.Context, vaultDenom string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)
	store.Delete(types.VaultKey(vaultDenom))
}

// SetVaultRecord sets the vault record for a given denom.
func (k *Keeper) SetVaultRecord(ctx sdk.Context, record types.VaultRecord) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.VaultKey(record.TotalShares.Denom), bz)
}

// IterateVaultRecords iterates over all vault objects in the store and performs
// a callback function.
func (k Keeper) IterateVaultRecords(
	ctx sdk.Context,
	cb func(record types.VaultRecord) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.VaultRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// GetAllVaultRecords returns all vault records from the store.
func (k Keeper) GetAllVaultRecords(ctx sdk.Context) types.VaultRecords {
	var records types.VaultRecords

	k.IterateVaultRecords(ctx, func(record types.VaultRecord) bool {
		records = append(records, record)
		return false
	})

	return records
}
