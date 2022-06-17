package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/kava/x/liquidstaking/types"
)

// Keeper struct for savings module
type Keeper struct {
	key           sdk.StoreKey
	cdc           codec.Codec
	paramSubspace paramtypes.Subspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	hooks         types.LiquidStakingHooks
}

// NewKeeper returns a new keeper for the liquidstaking module.
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
		hooks:         nil,
	}
}

// SetHooks adds hooks to the keeper.
func (k *Keeper) SetHooks(hooks types.MultiLiquidStakingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set liquidstaking hooks twice")
	}
	k.hooks = hooks
	return k
}

// GetDelegationHolder returns a DelegationHolder from the store for a particular validator address
func (k Keeper) GetDelegationHolder(ctx sdk.Context, validator sdk.ValAddress) (types.DelegationHolder, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegationHoldersKeyPrefix)
	bz := store.Get(validator.Bytes())
	if len(bz) == 0 {
		return types.DelegationHolder{}, false
	}
	var delegationHolder types.DelegationHolder
	k.cdc.MustUnmarshal(bz, &delegationHolder)
	return delegationHolder, true
}

// SetDelegationHolder sets the input DelegationHolder in the store
func (k Keeper) SetDelegationHolder(ctx sdk.Context, delegationHolder types.DelegationHolder) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegationHoldersKeyPrefix)
	bz := k.cdc.MustMarshal(&delegationHolder)
	store.Set(delegationHolder.Validator.Bytes(), bz)
}

// DeleteDelegationHolder deletes a DelegationHolder from the store
func (k Keeper) DeleteDelegationHolder(ctx sdk.Context, delegationHolder types.DelegationHolder) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegationHoldersKeyPrefix)
	store.Delete(delegationHolder.Validator.Bytes())
}

// IterateDelegationHolders iterates over all DelegationHolder objects in the store and performs a callback function
func (k Keeper) IterateDelegationHolders(ctx sdk.Context, cb func(delegationHolder types.DelegationHolder) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DelegationHoldersKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var delegationHolder types.DelegationHolder
		k.cdc.MustUnmarshal(iterator.Value(), &delegationHolder)
		if cb(delegationHolder) {
			break
		}
	}
}

// GetAllDelegationHolders returns all DelegationHolders from the store
func (k Keeper) GetAllDelegationHolders(ctx sdk.Context) (delegationHolders types.DelegationHolders) {
	k.IterateDelegationHolders(ctx, func(delegationHolder types.DelegationHolder) bool {
		delegationHolders = append(delegationHolders, delegationHolder)
		return false
	})
	return
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
