package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/savings/types"
)

// Keeper struct for savings module
type Keeper struct {
	key           sdk.StoreKey
	cdc           codec.Codec
	paramSubspace paramtypes.Subspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper returns a new keeper for the savings module.
func NewKeeper(
	cdc codec.Codec, key sdk.StoreKey, paramstore paramtypes.Subspace,
	ak types.AccountKeeper, bk types.BankKeeper,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		key:           key,
		paramSubspace: paramstore,
		accountKeeper: ak,
		bankKeeper:    bk,
	}
}

// GetDeposit returns a deposit from the store for a particular depositor address, deposit denom
func (k Keeper) GetDeposit(ctx sdk.Context, depositor sdk.AccAddress) (types.Deposit, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	bz := store.Get(depositor.Bytes())
	if len(bz) == 0 {
		return types.Deposit{}, false
	}
	var deposit types.Deposit
	k.cdc.MustUnmarshal(bz, &deposit)
	return deposit, true
}

// SetDeposit sets the input deposit in the store
func (k Keeper) SetDeposit(ctx sdk.Context, deposit types.Deposit) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	bz := k.cdc.MustMarshal(&deposit)
	store.Set(deposit.Depositor.Bytes(), bz)
}

// DeleteDeposit deletes a deposit from the store
func (k Keeper) DeleteDeposit(ctx sdk.Context, deposit types.Deposit) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	store.Delete(deposit.Depositor.Bytes())
}

// IterateDeposits iterates over all deposit objects in the store and performs a callback function
func (k Keeper) IterateDeposits(ctx sdk.Context, cb func(deposit types.Deposit) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositsKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		if cb(deposit) {
			break
		}
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
