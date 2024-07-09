package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/evmutil/types"
)

// MigrateStore
func MigrateStore(
	ctx sdk.Context,
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	precisebankkeeper types.PreciseBankKeeper,
) error {
	store := ctx.KVStore(storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AccountStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var acc types.Account
		if err := cdc.Unmarshal(iterator.Value(), &acc); err != nil {
			panic(err)
		}

		// Panics if the balance is invalid.
		precisebankkeeper.SetFractionalBalance(ctx, acc.Address, acc.Balance)

		// Delete after transferring balance to x/precisebank.
		store.Delete(iterator.Key())
	}

	return nil
}
