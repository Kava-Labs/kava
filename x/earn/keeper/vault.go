package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

type ViewVaultKeeper interface {
	// GetTotalSupplied returns the total balance supplied to the. This
	// may not necessarily be the current value of the vault, as it is the sum
	// of the supplied denom.
	GetVaultTotalSupplied(ctx sdk.Context, denom string) (sdk.Coin, error)

	// GetTotalValue returns the total **value** of all coins in this vault,
	// i.e. the realizable total value denominated by GetDenom() if the vault
	// were to liquidate its entire strategies.
	GetVaultTotalValue(ctx sdk.Context, denom string) (sdk.Coin, error)

	// GetAccountSupplied returns the supplied amount for a single address
	// within the vault.
	GetVaultAccountSupplied(ctx sdk.Context, denom string, acc sdk.AccAddress) (sdk.Coin, error)

	// GetAccountValue returns the value of a single address within the vault
	// if the account were to withdraw their entire balance.
	GetVaultAccountValue(ctx sdk.Context, denom string, acc sdk.AccAddress) (sdk.Coin, error)

	// GetStrategy returns the strategy the vault is associated with.
	GetVaultStrategy(ctx sdk.Context, denom string) Strategy

	// Deposit a coin into a vault. The coin denom determines the vault.
	Deposit(ctx sdk.Context, acc sdk.AccAddress, amount sdk.Coin) error

	// Withdraw a coin from a vault. The coin denom determines the vault.
	Withdraw(ctx sdk.Context, acc sdk.AccAddress, amount sdk.Coin) error
}

// ----------------------------------------------------------------------------
// VaultRecord

func (k *Keeper) GetVaultRecord(ctx sdk.Context, vaultDenom string) (types.VaultRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)

	bz := store.Get(types.VaultKey(vaultDenom))
	if bz == nil {
		return types.VaultRecord{}, false
	}

	var record types.VaultRecord
	k.cdc.MustUnmarshal(bz, &record)

	return record, true
}

func (k *Keeper) SetVaultRecord(ctx sdk.Context, record types.VaultRecord) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.VaultKey(record.Denom), bz)
}

// ----------------------------------------------------------------------------
// VaultShare

func (k *Keeper) GetVaultShareRecord(
	ctx sdk.Context,
	vaultDenom string,
	acc sdk.AccAddress,
) (types.VaultShareRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)

	bz := store.Get(types.DepositorVaultSharesKey(acc, vaultDenom))
	if bz == nil {
		return types.VaultShareRecord{}, false
	}

	var record types.VaultShareRecord
	k.cdc.MustUnmarshal(bz, &record)

	return record, true
}

func (k *Keeper) SetVaultShareRecord(
	ctx sdk.Context,
	record types.VaultShareRecord,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.VaultRecordKeyPrefix)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.DepositorVaultSharesKey(record.Depositor, record.AmountSupplied.Denom), bz)
}
