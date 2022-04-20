package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// AccumulateSavingsRewards calculates new rewards to distribute this block and updates the global indexes
func (k Keeper) AccumulateSavingsRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {

	previousAccrualTime, found := k.GetSavingsRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetSavingsRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	savingsMacc := k.accountKeeper.GetModuleAccount(ctx, savingstypes.ModuleName)
	maccCoins := k.bankKeeper.GetAllBalances(ctx, savingsMacc.GetAddress())
	denomBalance := maccCoins.AmountOf(rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, denomBalance.ToDec(), ctx.BlockTime())

	k.SetSavingsRewardAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)

	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetSavingsRewardIndexes(ctx, rewardPeriod.CollateralType, acc.Indexes)
	}
}

func (k Keeper) SynchronizeSavingsClaim(ctx sdk.Context, owner sdk.AccAddress) {
	deposit, found := k.savingsKeeper.GetDeposit(ctx, owner)
	if !found {
		return
	}
	k.SynchronizeSavingsReward(ctx, deposit)
}

// SynchronizeSavingsReward updates the claim object by adding any accumulated rewards
// and updating the reward index value
func (k Keeper) SynchronizeSavingsReward(ctx sdk.Context, deposit savingstypes.Deposit) {
	claim, found := k.GetSavingsClaim(ctx, deposit.Depositor)
	if !found {
		return
	}

	// Source shares for savings deposits is the deposit amount
	for _, coin := range deposit.Amount {
		claim = k.synchronizeSingleSavingsReward(ctx, claim, coin.Denom, coin.Amount.ToDec())
	}
	k.SetSavingsClaim(ctx, claim)
}

// synchronizeSingleSavingsReward synchronizes a single rewarded savings denom in a s claim.
// It returns the claim without setting in the store.
// The public methods for accessing and modifying claims are preferred over this one. Direct modification of claims is easy to get wrong.
func (k Keeper) synchronizeSingleSavingsReward(ctx sdk.Context, claim types.SavingsClaim, denom string, sourceShares sdk.Dec) types.SavingsClaim {
	globalRewardIndexes, found := k.GetSavingsRewardIndexes(ctx, denom)
	if !found {
		// The global factor is only not found if
		// - the savings denom has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded denoms.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return claim
	}

	userRewardIndexes, found := claim.RewardIndexes.Get(denom)
	if !found {
		// Normally the reward indexes should always be found.
		// But if a denom was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that supplied denom.
		// So given the reward period was just added, assume the starting value for any global reward indexes, which is an empty slice.
		userRewardIndexes = types.RewardIndexes{}
	}

	newRewards, err := k.CalculateRewards(userRewardIndexes, globalRewardIndexes, sourceShares)
	if err != nil {
		// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
		// This panics if a global reward factor decreases or disappears between the old and new indexes.
		panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
	}

	claim.Reward = claim.Reward.Add(newRewards...)
	claim.RewardIndexes = claim.RewardIndexes.With(denom, globalRewardIndexes)

	return claim
}

// GetSynchronizedSavingsClaim fetches a savings claim from the store and syncs rewards for all rewarded pools.
func (k Keeper) GetSynchronizedSavingsClaim(ctx sdk.Context, owner sdk.AccAddress) (types.SavingsClaim, bool) {
	claim, found := k.GetSavingsClaim(ctx, owner)
	if !found {
		return types.SavingsClaim{}, false
	}

	deposit, found := k.savingsKeeper.GetDeposit(ctx, owner)
	if !found {
		return types.SavingsClaim{}, false
	}

	for _, coin := range deposit.Amount {
		claim = k.synchronizeSingleSavingsReward(ctx, claim, coin.Denom, coin.Amount.ToDec())
	}

	return claim, true
}
