package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/evmutil/types"
)

// MigrateStore performs in-place store migrations for consensus version 3.
// V3 moves all account balances to x/precisebank and deletes it from x/evmutil.
func MigrateStore(
	ctx sdk.Context,
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	precisebankkeeper types.PreciseBankKeeper,
) error {
	store := ctx.KVStore(storeKey)
	iterator := sdk.KVStorePrefixIterator(store, AccountStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		// TODO: Change types package to v3types once moved to v3/types and
		// removed from evmutil/types
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
