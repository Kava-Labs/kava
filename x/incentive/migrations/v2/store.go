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

	// Migrate USDX minting claims
	if err := migrateUSDXMintingClaims(store, cdc); err != nil {
		return err
	}

	return nil
}

func migrateUSDXMintingClaims(store sdk.KVStore, cdc codec.BinaryCodec) error {
	newStore := prefix.NewStore(store, types.GetClaimKeyPrefix(types.CLAIM_TYPE_USDX_MINTING))

	iterator := sdk.KVStorePrefixIterator(store, USDXMintingClaimKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c v1types.USDXMintingClaim
		cdc.MustUnmarshal(iterator.Value(), &c)

		if err := c.Validate(); err != nil {
			return err
		}

		// Convert to the new Claim type
		newClaim := types.NewClaim(
			types.CLAIM_TYPE_USDX_MINTING,
			c.Owner,
			sdk.NewCoins(c.Reward),
			usdxRewardIndexToMultiRewardIndexes(c.RewardIndexes),
		)

		// Set in the **newStore** for the new store prefix
		newStore.Set(c.Owner, cdc.MustMarshal(&newClaim))
	}

	return nil
}

func usdxRewardIndexToMultiRewardIndexes(index v1types.RewardIndexes) types.MultiRewardIndexes {
	var multi types.MultiRewardIndexes
	for _, i := range index {
		// Each collateral type has "ukava" reward index
		multiRewardIndex := types.NewMultiRewardIndex(i.CollateralType, types.RewardIndexes{
			types.NewRewardIndex(types.USDXMintingRewardDenom, i.RewardFactor),
		})

		multi = append(multi, multiRewardIndex)
	}

	return multi
}
