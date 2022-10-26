package distributor

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/incentive/types"
)

type SharesUpdateListener interface {
	SharesUpdated(ctx sdk.Context, sourceID string, owner sdk.AccAddress, oldShares sdk.Dec)
	SharesCreated(ctx sdk.Context, sourceID string, owner sdk.AccAddress)
}

type SourceAdapter interface {
	// returns sum of all user's source shares, eg swap pool shares, or savings deposits
	GetTotalShares(ctx sdk.Context, sourceID string) sdk.Dec
	// returns a user's share balance, eg swap pool shares, or savings deposit for a particular denom
	GetShares(ctx sdk.Context, sourceID string, owner sdk.AccAddress) sdk.Dec
	// RegisterSharesUpdateListener registers hooks to be called when a user's share balance changes
	RegisterSharesUpdateListener(SharesUpdateListener)
}

// store used by a Distributor
type DistributorStore interface {
	SetClaim(ctx sdk.Context, claim types.Claim)
	GetClaim(ctx sdk.Context, owner sdk.AccAddress) (types.Claim, bool)
	GetGlobalIndexes(ctx sdk.Context, sourceID string) (types.RewardIndexes, bool)
	SetGlobalIndexes(ctx sdk.Context, sourceID string, indexes types.RewardIndexes)
	IterateGlobalIndexes(ctx sdk.Context, callback func(sourceID string, indexes types.RewardIndexes) bool)
}

type Distributor struct {
	store  DistributorStore
	source SourceAdapter
}

func New(store DistributorStore, source SourceAdapter) *Distributor {
	d := &Distributor{
		store:  store,
		source: source,
	}
	d.source.RegisterSharesUpdateListener(shareUpdateListener{d})
	return d
}

// Distribute allocates amt coins rateably between all source share owners.
// Internally it fetches the total source shares and updates the global indexes, but callers shouldn't need to know about this.
func (d *Distributor) Distribute(ctx sdk.Context, sourceID string, amt sdk.DecCoins) {
	indexes, found := d.store.GetGlobalIndexes(ctx, sourceID)
	if !found {
		indexes = types.RewardIndexes{}
	}
	totalShares := d.source.GetTotalShares(ctx, sourceID)

	increment := types.NewRewardIndexesFromCoins(amt).Quo(totalShares)
	indexes = indexes.Add(increment)

	if len(indexes) > 0 {
		// the store panics when setting empty or nil indexes // TODO does it?
		d.store.SetGlobalIndexes(ctx, sourceID, indexes)
	}
}

// GetUserBalance returns the up to date (ie synced) rewards balance of a user.
func (d *Distributor) GetUserBalance(ctx sdk.Context, owner sdk.AccAddress) sdk.Coins {

	// same as current keeper.GetSynchronizedClaim, but only returns the balance

	claim, found := d.store.GetClaim(ctx, owner)
	if !found {
		return sdk.NewCoins()
	}

	d.store.IterateGlobalIndexes(ctx, func(sourceID string, _ types.RewardIndexes) bool {

		shares := d.source.GetShares(ctx, sourceID, owner)

		claim = d.synchronizeReward(ctx, claim, sourceID, owner, shares)

		return false
	})

	return claim.Reward
}

// WithdrawUserBalance removes some rewards from the user's reward balance.
func (d *Distributor) WithdrawUserBalance(ctx sdk.Context, owner sdk.AccAddress, amt sdk.Coins) error {
	balance := d.GetUserBalance(ctx, owner)

	balance = balance.Sub(amt)

	claim, found := d.store.GetClaim(ctx, owner)
	if !found {
		panic("TODO")
	}
	claim.Reward = balance
	d.store.SetClaim(ctx, claim)
	return nil
}

// sharesUpdateListener holds the Sync and Init methods to update claims when source shares change.
// it mainly exists to avoid exposing these methods on the public Distributor interface as they can be dangerous if called incorrectly.
type shareUpdateListener struct {
	*Distributor
}

// SharesUpdated should be called whenever a user's source share amount changes.
func (d shareUpdateListener) SharesUpdated(ctx sdk.Context, sourceID string, owner sdk.AccAddress, oldShares sdk.Dec) {

	// same as current keeper.SynchronizeReward

	claim, found := d.store.GetClaim(ctx, owner)
	if !found {
		return
	}
	claim = d.synchronizeReward(ctx, claim, sourceID, owner, oldShares)

	d.store.SetClaim(ctx, claim)
}

// SharesCreated should be called whenever a user's source share amount is instantiated.
// This is not strictly necessary, SharesUpdated could do everything.
func (d shareUpdateListener) SharesCreated(ctx sdk.Context, sourceID string, owner sdk.AccAddress) {

	// same as current keeper.InitializeReward

	claim, found := d.store.GetClaim(ctx, owner)
	if !found {
		claim = types.NewClaim(owner, sdk.Coins{}, nil)
	}

	globalRewardIndexes, found := d.store.GetGlobalIndexes(ctx, sourceID)
	if !found {
		globalRewardIndexes = types.RewardIndexes{}
	}
	claim.RewardIndexes = claim.RewardIndexes.With(sourceID, globalRewardIndexes)

	d.store.SetClaim(ctx, claim)
}

// ----- These methods copied in from current keeper, there might be a better way to do this -----

// synchronizeSwapReward updates the reward and indexes in a swap claim for one pool.
func (d *Distributor) synchronizeReward(ctx sdk.Context, claim types.Claim, poolID string, owner sdk.AccAddress, shares sdk.Dec) types.Claim {
	globalRewardIndexes, found := d.store.GetGlobalIndexes(ctx, poolID)
	if !found {
		// The global factor is only not found if
		// - the pool has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded pools.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return claim
	}

	userRewardIndexes, found := claim.RewardIndexes.Get(poolID)
	if !found {
		// Normally the reward indexes should always be found.
		// But if a pool was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that pool.
		// So given the reward period was just added, assume the starting value for any global reward indexes, which is an empty slice.
		userRewardIndexes = types.RewardIndexes{}
	}

	newRewards, err := calculateRewards(userRewardIndexes, globalRewardIndexes, shares)
	if err != nil {
		// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
		// This panics if a global reward factor decreases or disappears between the old and new indexes.
		panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
	}

	claim.Reward = claim.Reward.Add(newRewards...)
	claim.RewardIndexes = claim.RewardIndexes.With(poolID, globalRewardIndexes)

	return claim
}

// CalculateRewards computes how much rewards should have accrued to a reward source (eg a user's hard borrowed btc amount)
// between two index values.
//
// oldIndex is normally the index stored on a claim, newIndex the current global value, and sourceShares a hard borrowed/supplied amount.
//
// It returns an error if newIndexes does not contain all CollateralTypes from oldIndexes, or if any value of oldIndex.RewardFactor > newIndex.RewardFactor.
// This should never happen, as it would mean that a global reward index has decreased in value, or that a global reward index has been deleted from state.
func calculateRewards(oldIndexes, newIndexes types.RewardIndexes, sourceShares sdk.Dec) (sdk.Coins, error) {
	// check for missing CollateralType's
	for _, oldIndex := range oldIndexes {
		if newIndex, found := newIndexes.Get(oldIndex.CollateralType); !found {
			return nil, sdkerrors.Wrapf(types.ErrDecreasingRewardFactor, "old: %v, new: %v", oldIndex, newIndex)
		}
	}
	var reward sdk.Coins
	for _, newIndex := range newIndexes {
		oldFactor, found := oldIndexes.Get(newIndex.CollateralType)
		if !found {
			oldFactor = sdk.ZeroDec()
		}

		rewardAmount, err := calculateSingleReward(oldFactor, newIndex.RewardFactor, sourceShares)
		if err != nil {
			return nil, err
		}

		reward = reward.Add(
			sdk.NewCoin(newIndex.CollateralType, rewardAmount),
		)
	}
	return reward, nil
}

// CalculateSingleReward computes how much rewards should have accrued to a reward source (eg a user's btcb-a cdp principal)
// between two index values.
//
// oldIndex is normally the index stored on a claim, newIndex the current global value, and sourceShares a cdp principal amount.
//
// Returns an error if oldIndex > newIndex. This should never happen, as it would mean that a global reward index has decreased in value,
// or that a global reward index has been deleted from state.
func calculateSingleReward(oldIndex, newIndex, sourceShares sdk.Dec) (sdk.Int, error) {
	increase := newIndex.Sub(oldIndex)
	if increase.IsNegative() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrDecreasingRewardFactor, "old: %v, new: %v", oldIndex, newIndex)
	}
	reward := increase.Mul(sourceShares).RoundInt()
	return reward, nil
}
