package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// GetVaultTotalSupplied returns the total balance supplied to the vault. This
// may not necessarily be the current value of the vault, as it is the sum
// of the supplied denom and the value may be higher due to accumulated APYs.
func (k *Keeper) GetVaultTotalSupplied(
	ctx sdk.Context,
	denom string,
) (sdk.Coin, error) {
	vault, found := k.GetVaultRecord(ctx, denom)
	if !found {
		return sdk.Coin{}, types.ErrVaultRecordNotFound
	}

	return vault.TotalSupply, nil
}

// GetTotalValue returns the total **value** of all coins in this vault,
// i.e. the realizable total value denominated by GetDenom() if the vault
// were to liquidate its entire strategies.
//
// **Note:** This does not include the tokens held in bank by the module
// account. If it were to be included, also note that the module account is
// unblocked and can receive funds from bank sends.
func (k *Keeper) GetVaultTotalValue(
	ctx sdk.Context,
	denom string,
) (sdk.Coin, error) {
	enabledVault, found := k.GetAllowedVault(ctx, denom)
	if !found {
		return sdk.Coin{}, types.ErrVaultRecordNotFound
	}

	strategy, err := k.GetStrategy(enabledVault.VaultStrategy)
	if err != nil {
		return sdk.Coin{}, types.ErrInvalidVaultStrategy
	}

	return strategy.GetEstimatedTotalAssets(ctx, enabledVault.Denom)
}

// GetVaultAccountSupplied returns the supplied amount for a single address
// within a vault.
func (k *Keeper) GetVaultAccountSupplied(
	ctx sdk.Context,
	acc sdk.AccAddress,
) (sdk.Coins, error) {
	vaultShareRecord, found := k.GetVaultShareRecord(ctx, acc)
	if !found {
		return sdk.Coins{}, types.ErrVaultShareRecordNotFound
	}

	return vaultShareRecord.AmountSupplied, nil
}

// GetVaultAccountValue returns the value of a single address within a vault
// if the account were to withdraw their entire balance.
func (k *Keeper) GetVaultAccountValue(
	ctx sdk.Context,
	denom string,
	acc sdk.AccAddress,
) (sdk.Coin, error) {
	totalSupplied, err := k.GetVaultTotalSupplied(ctx, denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	accSupplied, err := k.GetVaultAccountSupplied(ctx, acc)
	if err != nil {
		return sdk.Coin{}, err
	}

	vaultTotalValue, err := k.GetVaultTotalValue(ctx, denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Percent of vault account ownership = accountSupply / totalSupply
	// Value of vault account ownership = percentOwned * totalValue
	vaultShare := accSupplied.AmountOf(denom).ToDec().Quo(totalSupplied.Amount.ToDec())
	shareValueDec := vaultTotalValue.Amount.ToDec().Mul(vaultShare)

	return sdk.NewCoin(denom, shareValueDec.TruncateInt()), nil
}

// ----------------------------------------------------------------------------
// VaultRecord -- vault total supplies

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
	if vaultRecord.TotalSupply.IsZero() {
		k.DeleteVaultRecord(ctx, vaultRecord.Denom)
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
	store.Set(types.VaultKey(record.Denom), bz)
}

// ----------------------------------------------------------------------------
// VaultShare -- user shares per vault

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
	if record.AmountSupplied.IsZero() {
		k.DeleteVaultShareRecord(ctx, record.Depositor)
	} else {
		k.SetVaultShareRecord(ctx, record)
	}
}

// DeleteVaultShareRecord deletes the vault share record for a given denom and
// account.
func (k *Keeper) DeleteVaultShareRecord(
	ctx sdk.Context,
	acc sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultShareRecordKeyPrefix)
	store.Delete(types.DepositorVaultSharesKey(acc))
}

// SetVaultShareRecord sets the vault share record for a given denom and account.
func (k *Keeper) SetVaultShareRecord(
	ctx sdk.Context,
	record types.VaultShareRecord,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultShareRecordKeyPrefix)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.DepositorVaultSharesKey(record.Depositor), bz)
}
