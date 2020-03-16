package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Keeper keeper for the incentive module
type Keeper struct {
	key           sdk.StoreKey
	cdc           *codec.Codec
	paramSubspace subspace.Subspace
	supplyKeeper  types.SupplyKeeper
	cdpKeeper     types.CdpKeeper
	accountKeeper types.AccountKeeper
	codespace     sdk.CodespaceType
}

// NewKeeper creates a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, sk types.SupplyKeeper, cdpk types.CdpKeeper, ak types.AccountKeeper, codespace sdk.CodespaceType) Keeper {

	return Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
		supplyKeeper:  sk,
		cdpKeeper:     cdpk,
		accountKeeper: ak,
		codespace:     codespace,
	}
}

// GetPreviousBlockTime get the blocktime for the previous block
func (k Keeper) GetPreviousBlockTime(ctx sdk.Context) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	b := store.Get([]byte{})
	if b == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &blockTime)
	return blockTime, true
}

// SetPreviousBlockTime set the time of the previous block
func (k Keeper) SetPreviousBlockTime(ctx sdk.Context, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousBlockTimeKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryLengthPrefixed(blockTime))
}
