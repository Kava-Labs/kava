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

// GetDerivative returns a derivative from the store for a particular validator address
func (k Keeper) GetDerivative(ctx sdk.Context, validator sdk.ValAddress) (types.Derivative, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DerivativesKeyPrefix)
	bz := store.Get(validator.Bytes())
	if len(bz) == 0 {
		return types.Derivative{}, false
	}
	var derivative types.Derivative
	k.cdc.MustUnmarshal(bz, &derivative)
	return derivative, true
}

// SetDerivative sets the input derivative in the store
func (k Keeper) SetDerivative(ctx sdk.Context, derivative types.Derivative) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DerivativesKeyPrefix)
	bz := k.cdc.MustMarshal(&derivative)
	store.Set(derivative.Validator.Bytes(), bz)
}

// DeleteDerivative deletes a derivative from the store
func (k Keeper) DeleteDerivative(ctx sdk.Context, derivitive types.Derivative) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DerivativesKeyPrefix)
	store.Delete(derivitive.Validator.Bytes())
}

// IterateDerivatives iterates over all derivative objects in the store and performs a callback function
func (k Keeper) IterateDerivatives(ctx sdk.Context, cb func(derivative types.Derivative) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DerivativesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var derivative types.Derivative
		k.cdc.MustUnmarshal(iterator.Value(), &derivative)
		if cb(derivative) {
			break
		}
	}
}

// GetAllDerivatives returns all derivatives from the store
func (k Keeper) GetAllDerivatives(ctx sdk.Context) (derivatives types.Derivatives) {
	k.IterateDerivatives(ctx, func(derivative types.Derivative) bool {
		derivatives = append(derivatives, derivative)
		return false
	})
	return
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
