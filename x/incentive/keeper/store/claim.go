package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/kava-labs/kava/x/incentive/types"
)

// GetClaim returns the claim in the store corresponding the the owner and
// claimType, and a boolean for if the claim was found
func (k IncentiveStore) GetClaim(
	ctx sdk.Context,
	claimType types.ClaimType,
	addr sdk.AccAddress,
) (types.Claim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetClaimKeyPrefix(claimType))
	bz := store.Get(addr)
	if bz == nil {
		return types.Claim{}, false
	}
	var c types.Claim
	k.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetClaim sets the claim in the store corresponding to the owner and claimType
func (k IncentiveStore) SetClaim(
	ctx sdk.Context,
	c types.Claim,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetClaimKeyPrefix(c.Type))
	bz := k.cdc.MustMarshal(&c)
	store.Set(c.Owner, bz)
}

// DeleteClaim deletes the claim in the store corresponding to the owner and claimType
func (k IncentiveStore) DeleteClaim(
	ctx sdk.Context,
	claimType types.ClaimType,
	owner sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.GetClaimKeyPrefix(claimType))
	store.Delete(owner)
}

// IterateClaimsByClaimType iterates over all claim objects in the store of a given
// claimType and preforms a callback function
func (k IncentiveStore) IterateClaimsByClaimType(
	ctx sdk.Context,
	claimType types.ClaimType,
	cb func(c types.Claim) (stop bool),
) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.GetClaimKeyPrefix(claimType))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.Claim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetClaimsOfClaimType returns all Claim objects in the store of the given claimType
func (k IncentiveStore) GetClaimsOfClaimType(
	ctx sdk.Context,
	claimType types.ClaimType,
) types.Claims {
	var cs types.Claims
	k.IterateClaimsByClaimType(ctx, claimType, func(c types.Claim) (stop bool) {
		cs = append(cs, c)
		return false
	})

	return cs
}

// GetClaims returns all Claim objects in the store of a given claimType
func (k IncentiveStore) GetClaims(
	ctx sdk.Context,
	claimType types.ClaimType,
) types.Claims {
	var cs types.Claims
	k.IterateClaimsByClaimType(ctx, claimType, func(c types.Claim) (stop bool) {
		cs = append(cs, c)
		return false
	})

	return cs
}

// IterateClaims iterates over all claim objects of any claimType in the
// store and preforms a callback function
func (k IncentiveStore) IterateClaims(
	ctx sdk.Context,
	cb func(c types.Claim) (stop bool),
) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.ClaimKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.Claim
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		if cb(c) {
			break
		}
	}
}

// GetAllClaims returns all Claim objects in the store of any claimType
func (k IncentiveStore) GetAllClaims(ctx sdk.Context) types.Claims {
	var cs types.Claims
	k.IterateClaims(ctx, func(c types.Claim) (stop bool) {
		cs = append(cs, c)
		return false
	})

	return cs
}
