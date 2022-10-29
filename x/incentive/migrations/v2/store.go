package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1types "github.com/kava-labs/kava/x/incentive/migrations/v2/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// MigrateStore performs in-place migrations from incentive ConsensusVersion 1 to 2.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	// Migrate earn claims
	if err := migrateEarnClaims(store, cdc); err != nil {
		return err
	}

	return nil
}

func migrateEarnClaims(store sdk.KVStore, cdc codec.BinaryCodec) error {
	newStore := prefix.NewStore(store, types.GetClaimKeyPrefix(types.CLAIM_TYPE_EARN))

	iterator := sdk.KVStorePrefixIterator(store, EarnClaimKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c v1types.EarnClaim
		cdc.MustUnmarshal(iterator.Value(), &c)

		if err := c.Validate(); err != nil {
			return err
		}

		// Convert to the new Claim type
		newClaim := types.NewClaim(
			types.CLAIM_TYPE_EARN,
			c.Owner,
			c.Reward,
			MultiRewardIndexesV1ToV2(c.RewardIndexes),
		)

		// Set in the **newStore** for the new store prefix
		newStore.Set(c.Owner, cdc.MustMarshal(&newClaim))
	}

	return nil
}

func RewardIndexesV1ToV2(v1 v1types.RewardIndexes) types.RewardIndexes {
	v2 := make(types.RewardIndexes, len(v1))
	for i, r := range v1 {
		v2[i] = types.NewRewardIndex(r.CollateralType, r.RewardFactor)
	}

	return v2
}

func MultiRewardIndexesV1ToV2(v1 v1types.MultiRewardIndexes) types.MultiRewardIndexes {
	v2 := make(types.MultiRewardIndexes, len(v1))
	for i, r := range v1 {
		v2[i] = types.NewMultiRewardIndex(r.CollateralType, RewardIndexesV1ToV2(r.RewardIndexes))
	}

	return v2
}
