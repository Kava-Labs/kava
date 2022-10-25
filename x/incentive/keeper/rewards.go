package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

type RewardType int

const (
	RewardTypeEarn RewardType = iota + 1
	RewardTypeSavings
	RewardTypeHardSupply
	RewardTypeHardBorrow
)

type NewParams struct {
	// module -> reward period
	RewardPeriods map[string]types.MultiRewardPeriod
	ClaimEnd      time.Time
}

func rewardTypeFromString(rewardType string) RewardType {
	switch rewardType {
	case "hard_borrow":
		return RewardTypeHardBorrow
	case "hard_supply":
		return RewardTypeHardSupply
	case "savings":
		return RewardTypeSavings
	case "earn":
		return RewardTypeEarn
	default:
		panic("invalid reward type")
	}
}

func (k Keeper) getAccumulatorFromRewardType(rewardType RewardType) RewardTypeStrategy {
	switch rewardType {
	case RewardTypeEarn:
		return NewTestEarnAccumulator(k)
	default:
		panic("invalid scope")
	}
}

func (k Keeper) AccumulateAllRewards(ctx sdk.Context, params NewParams) error {
	for rewardTypeStr, rp := range params.RewardPeriods {
		rewardType := rewardTypeFromString(rewardTypeStr)
		strategy := k.getAccumulatorFromRewardType(rewardType)
		k.AccumulateRewards(ctx, strategy, rp)
	}

	return nil
}

func (k Keeper) AccumulateRewards(
	ctx sdk.Context,
	strategy RewardTypeStrategy,
	rewardPeriod types.MultiRewardPeriod,
) {
	storeKey := strategy.getStoreKey()

	previousAccrualTime, found := k.GetRewardAccrualTime(ctx, storeKey, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetRewardIndexes(ctx, storeKey, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSourceShares := strategy.getTotalSourceShares(ctx, rewardPeriod.CollateralType)
	acc.Accumulate(rewardPeriod, totalSourceShares, ctx.BlockTime())

	// Additional rewards, e.g. bkava
	additionalRewardIndexes := strategy.getAdditionalRewardIndexes(ctx, rewardPeriod.CollateralType)
	acc.Indexes = acc.Indexes.Add(additionalRewardIndexes)

	k.SetRewardAccrualTime(ctx, storeKey, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)

	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetRewardIndexes(ctx, storeKey, rewardPeriod.CollateralType, acc.Indexes)
	}
}

func (k *Keeper) SynchronizeClaimReward(
	ctx sdk.Context,
	claim types.Claim,
	collateralType string,
	shares sdk.Dec,
) types.Claim {
	globalRewardIndexes, found := k.GetRewardIndexes(ctx, claim.GetPrefix(), collateralType)
	if !found {
		// The global factor is only not found if
		// - the vault has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded vaults.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return claim
	}

	userRewardIndexes, found := claim.GetRewardIndexes().Get(collateralType)
	if !found {
		// Normally the reward indexes should always be found.
		// But if a vault was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that vault.
		// So given the reward period was just added, assume the starting value for any global reward indexes, which is an empty slice.
		userRewardIndexes = types.RewardIndexes{}
	}

	newRewards, err := k.CalculateRewards(userRewardIndexes, globalRewardIndexes, shares)
	if err != nil {
		// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
		// This panics if a global reward factor decreases or disappears between the old and new indexes.
		panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
	}

	claim.SetReward(claim.GetReward().Add(newRewards...))
	claim.SetRewardIndexes(claim.GetRewardIndexes().With(collateralType, globalRewardIndexes))

	return claim
}

func (k Keeper) InitializeClaim(
	ctx sdk.Context,
	rewardType RewardType,
	indexDenom string,
	owner sdk.AccAddress,
) types.Claim {
	rewardTypeStrategy := k.getAccumulatorFromRewardType(rewardType)
	prefix := rewardTypeStrategy.getStoreKey()

	claim, found := k.GetClaim(ctx, prefix, owner)
	if !found {
		claim = rewardTypeStrategy.newClaim(ctx, owner)
	}

	globalRewardIndexes, found := k.GetRewardIndexes(ctx, prefix, indexDenom)
	if !found {
		globalRewardIndexes = types.RewardIndexes{}
	}

	newRewardIndexes := claim.GetRewardIndexes().With(indexDenom, globalRewardIndexes)
	claim.SetRewardIndexes(newRewardIndexes)

	return claim
}

type RewardTypeStrategy interface {
	getStoreKey() []byte
	getTotalSourceShares(ctx sdk.Context, key string) sdk.Dec
	getAccountSourceShares(ctx sdk.Context, key string, account sdk.AccAddress) sdk.Dec

	getAdditionalRewardIndexes(ctx sdk.Context, collateralType string) types.RewardIndexes

	newClaim(ctx sdk.Context, owner sdk.AccAddress) types.Claim
}

type TestEarnAccumulator struct {
	keeper Keeper
}

var _ RewardTypeStrategy = (*TestEarnAccumulator)(nil)

func NewTestEarnAccumulator(k Keeper) TestEarnAccumulator {
	return TestEarnAccumulator{k}
}

func (k TestEarnAccumulator) getStoreKey() []byte {
	return types.EarnClaimKeyPrefix
}

func (k TestEarnAccumulator) getTotalSourceShares(ctx sdk.Context, key string) sdk.Dec {
	totalShares, found := k.keeper.earnKeeper.GetVaultTotalShares(ctx, key)
	if !found {
		return sdk.ZeroDec()
	}

	return totalShares.Amount
}

func (k TestEarnAccumulator) getAccountSourceShares(ctx sdk.Context, key string, account sdk.AccAddress) sdk.Dec {
	shares, found := k.keeper.earnKeeper.GetVaultAccountShares(ctx, account)
	if !found {
		return sdk.ZeroDec()
	}

	return shares.AmountOf(key)
}

func (k TestEarnAccumulator) getAdditionalRewardIndexes(
	ctx sdk.Context,
	collateralType string,
) types.RewardIndexes {
	if collateralType != "bkava" {
		return nil
	}

	// Collect staking rewards

	return nil
}

func (k TestEarnAccumulator) newClaim(ctx sdk.Context, owner sdk.AccAddress) types.Claim {
	claim := types.NewEarnClaim(owner, sdk.NewCoins(), nil)
	return &claim
}

// ----------------------------
// state methods

// GetClaim returns the claim in the store corresponding the the input address.
func (k Keeper) GetClaim(ctx sdk.Context, claimPrefix []byte, addr sdk.AccAddress) (types.Claim, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), append(types.ClaimKeyPrefix, claimPrefix...))
	bz := store.Get(addr)
	if bz == nil {
		return nil, false
	}
	var c types.Claim
	k.cdc.UnmarshalInterface(bz, &c)
	return c, true
}

// SetClaim sets the claim in the store corresponding to the input claim.
func (k Keeper) SetClaim(ctx sdk.Context, c types.Claim) {
	store := prefix.NewStore(ctx.KVStore(k.key), append(types.ClaimKeyPrefix, c.GetPrefix()...))
	bz, err := k.cdc.MarshalInterface(c)
	if err != nil {
		panic(err)
	}

	store.Set(c.GetKey(), bz)
}

func (k Keeper) GetRewardAccrualTime(ctx sdk.Context, rewardPrefix []byte, key string) (blockTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), append(types.PreviousRewardAccrualTimeKeyPrefix, rewardPrefix...))
	b := store.Get([]byte(key))
	if b == nil {
		return time.Time{}, false
	}
	if err := blockTime.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return blockTime, true
}

func (k Keeper) SetRewardAccrualTime(ctx sdk.Context, rewardPrefix []byte, key string, blockTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), append(types.PreviousRewardAccrualTimeKeyPrefix, rewardPrefix...))
	bz, err := blockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(key), bz)
}

// SetSwapRewardIndexes stores the global reward indexes that track total rewards to a swap pool.
func (k Keeper) SetRewardIndexes(ctx sdk.Context, rewardPrefix []byte, key string, indexes types.RewardIndexes) {
	store := prefix.NewStore(ctx.KVStore(k.key), append(types.RewardIndexesKeyPrefix, rewardPrefix...))
	bz := k.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set([]byte(key), bz)
}

// GetSwapRewardIndexes fetches the global reward indexes that track total rewards to a swap pool.
func (k Keeper) GetRewardIndexes(ctx sdk.Context, rewardPrefix []byte, key string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), append(types.RewardIndexesKeyPrefix, rewardPrefix...))
	bz := store.Get([]byte(key))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	k.cdc.MustUnmarshal(bz, &proto)
	return proto.RewardIndexes, true
}
