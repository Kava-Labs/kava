package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/kava-labs/kava/x/cdp/types"
)

// Keeper keeper for the cdp module
type Keeper struct {
	key             sdk.StoreKey
	cdc             *codec.Codec
	paramSubspace   params.Subspace
	pricefeedKeeper types.PricefeedKeeper
	bankKeeper      types.BankKeeper
	auctionKeeper   types.AuctionKeeper
	accountKeeper   types.AccountKeeper
	maccPerms       map[string][]string
}

// NewKeeper creates a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace, pfk types.PricefeedKeeper,
	ak types.AuctionKeeper, bk types.BankKeeper, ack types.AccountKeeper, maccs map[string][]string) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:             key,
		cdc:             cdc,
		paramSubspace:   paramstore,
		pricefeedKeeper: pfk,
		auctionKeeper:   ak,
		bankKeeper:      bk,
		accountKeeper:   ack,
		maccPerms:       maccs,
	}
}

// CdpDenomIndexIterator returns an sdk.Iterator for all cdps with matching collateral denom
func (k Keeper) CdpDenomIndexIterator(ctx sdk.Context, denom string) sdk.Iterator {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	db, found := k.GetDenomPrefix(ctx, denom)
	if !found {
		panic(fmt.Sprintf("denom %s prefix not found", denom))
	}
	return sdk.KVStorePrefixIterator(store, types.DenomIterKey(db))
}

// CdpCollateralRatioIndexIterator returns an sdk.Iterator for all cdps that have collateral denom
// matching denom and collateral:debt ratio LESS THAN targetRatio
func (k Keeper) CdpCollateralRatioIndexIterator(ctx sdk.Context, denom string, targetRatio sdk.Dec) sdk.Iterator {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CollateralRatioIndexPrefix)
	db, found := k.GetDenomPrefix(ctx, denom)
	if !found {
		panic(fmt.Sprintf("denom %s prefix not found", denom))
	}
	return store.Iterator(types.CollateralRatioIterKey(db, sdk.ZeroDec()), types.CollateralRatioIterKey(db, targetRatio))
}

// IterateAllCdps iterates over all cdps and performs a callback function
func (k Keeper) IterateAllCdps(ctx sdk.Context, cb func(cdp types.CDP) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.CdpKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cdp types.CDP
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &cdp)

		if cb(cdp) {
			break
		}
	}
}

// IterateCdpsByDenom iterates over cdps with matching denom and performs a callback function
func (k Keeper) IterateCdpsByDenom(ctx sdk.Context, denom string, cb func(cdp types.CDP) (stop bool)) {
	iterator := k.CdpDenomIndexIterator(ctx, denom)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cdp types.CDP
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &cdp)
		if cb(cdp) {
			break
		}
	}
}

// IterateCdpsByCollateralRatio iterate over cdps with collateral denom equal to denom and
// collateral:debt ratio LESS THAN targetRatio and performs a callback function.
func (k Keeper) IterateCdpsByCollateralRatio(ctx sdk.Context, denom string, targetRatio sdk.Dec, cb func(cdp types.CDP) (stop bool)) {
	iterator := k.CdpCollateralRatioIndexIterator(ctx, denom, targetRatio)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		db, id, _ := types.SplitCollateralRatioKey(iterator.Key())
		d := k.getDenomFromByte(ctx, db)
		cdp, found := k.GetCDP(ctx, d, id)
		if !found {
			panic(fmt.Sprintf("cdp %d does not exist", id))
		}
		if cb(cdp) {
			break
		}

	}
}
