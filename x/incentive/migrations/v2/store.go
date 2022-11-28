package v2

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// MigrateStore performs in-place migrations from incentive ConsensusVersion 1 to 2.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	// Migrate earn claims
	if err := MigrateEarnClaims(store, cdc); err != nil {
		return err
	}

	// Migrate accrual times
	if err := MigrateAccrualTimes(store, cdc, types.CLAIM_TYPE_EARN); err != nil {
		return err
	}

	return nil
}

// MigrateEarnClaims migrates earn claims from v1 to v2
func MigrateEarnClaims(store sdk.KVStore, cdc codec.BinaryCodec) error {
	newStore := prefix.NewStore(store, types.GetClaimKeyPrefix(types.CLAIM_TYPE_EARN))

	iterator := sdk.KVStorePrefixIterator(store, EarnClaimKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.EarnClaim
		cdc.MustUnmarshal(iterator.Value(), &c)

		if err := c.Validate(); err != nil {
			return fmt.Errorf("invalid v1 EarnClaim: %w", err)
		}

		// Convert to the new Claim type
		newClaim := types.NewClaim(
			types.CLAIM_TYPE_EARN,
			c.Owner,
			c.Reward,
			c.RewardIndexes,
		)

		if err := newClaim.Validate(); err != nil {
			return fmt.Errorf("invalid v2 EarnClaim: %w", err)
		}

		// Set in the **newStore** for the new store prefix
		newStore.Set(c.Owner, cdc.MustMarshal(&newClaim))
	}

	return nil
}

func LegacyAccrualTimeKeyFromClaimType(claimType types.ClaimType) []byte {
	switch claimType {
	case types.CLAIM_TYPE_HARD_BORROW:
		panic("todo")
	case types.CLAIM_TYPE_HARD_SUPPLY:
		panic("todo")
	case types.CLAIM_TYPE_EARN:
		return PreviousEarnRewardAccrualTimeKeyPrefix
	case types.CLAIM_TYPE_SAVINGS:
		panic("todo")
	case types.CLAIM_TYPE_SWAP:
		panic("todo")
	case types.CLAIM_TYPE_USDX_MINTING:
		panic("todo")
	default:
		panic(fmt.Sprintf("unrecognized claim type: %s", claimType))
	}
}

// MigrateAccrualTimes migrates accrual times from v1 to v2
func MigrateAccrualTimes(
	store sdk.KVStore,
	cdc codec.BinaryCodec,
	claimType types.ClaimType,
) error {
	newStore := prefix.NewStore(store, types.GetPreviousRewardAccrualTimeKeyPrefix(claimType))

	// Need prefix.NewStore instead of using it directly in the iterator, as
	// there would be an extra space in the key
	legacyPrefix := LegacyAccrualTimeKeyFromClaimType(claimType)
	oldStore := prefix.NewStore(store, legacyPrefix)
	iterator := sdk.KVStorePrefixIterator(oldStore, []byte{})

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var blockTime time.Time
		if err := blockTime.UnmarshalBinary(iterator.Value()); err != nil {
			panic(err)
		}

		sourceID := string(iterator.Key())
		fmt.Printf("iterator key '%b'", iterator.Key())

		fmt.Printf("migrating accrual time for claim type %s, source id %v: %s", claimType, sourceID, blockTime)
		at := types.NewAccrualTime(claimType, sourceID, blockTime)
		if err := at.Validate(); err != nil {
			return fmt.Errorf("invalid v2 accrual time for claim type %s: %w", claimType, err)
		}

		// Set in the **newStore** for the new store prefix
		bz := cdc.MustMarshal(&at)
		newStore.Set(types.GetKeyFromSourceID(sourceID), bz)
	}

	return nil
}
