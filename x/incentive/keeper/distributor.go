package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
	swapkeeper "github.com/kava-labs/kava/x/swap/keeper"
)

type SharesUpdateListener interface {
	SharesUpdated(sourceID string, owner sdk.AccAddress, oldShares sdk.Dec)
	SharesCreated(sourceID string, owner sdk.AccAddress)
}

type SwapSourceGroup struct {
	swapKeeper swapkeeper.Keeper
	listener   SharesUpdateListener
}

func NewSwapSourceGroup(swapKeeper swapkeeper.Keeper) SwapSourceGroup {
	ssg := SwapSourceGroup{
		swapKeeper: swapKeeper,
	}
	swapKeeper.SetHooks(ssg) // keeper needs to be pointer value, any way to assert this?
	return ssg
}

func (ssg SwapSourceGroup) GetTotalShares(sourceID string) sdk.Dec {
	// ssg.swapKeeper.GetPoolShares(sourceID)
	return sdk.ZeroDec()
}

func (ssg SwapSourceGroup) RegisterSharesUpdateListener(listener SharesUpdateListener) {
	ssg.listener = listener
}

func (ssg SwapSourceGroup) AfterPoolDepositCreated(ctx sdk.Context, poolID string, depositor sdk.AccAddress, _ sdk.Int) {
	// ssg.hook.InitializeUser(ctx, types.RewardTypeSwap, poolID, depositor)
	ssg.listener.SharesCreated(poolID, depositor)
}

func (ssg SwapSourceGroup) BeforePoolDepositModified(ctx sdk.Context, poolID string, depositor sdk.AccAddress, sharesOwned sdk.Int) {
	// ssg.hook.SynchronizeReward(ctx, types.RewardTypeSwap, poolID, depositor, sharesOwned)
	ssg.listener.SharesUpdated(poolID, depositor, sharesOwned.ToDec())
}

// --------------

type Store struct {
	Keeper // temp to keep type check working
} // has all the keeper Get/Set/Iterate methods

type SwapStore struct {
	Store // embed to inherit methods?
}

func (ss SwapStore) GetClaim(sourceID string, owner sdk.AccAddress) types.Claim {
	// c := ss.GetSwapClaim(sourceID, owner)
	// return c.ToClaim()
	return types.Claim{}
}

// --------------

type SourceGroup interface {
	GetTotalShares(sourceID string) sdk.Dec
	GetShares(sourceID string, owner sdk.AccAddress) sdk.Dec
	RegisterSharesUpdateListener(SharesUpdateListener)
}

type DistributorStore interface {
	GetClaim(sourceID string, owner sdk.AccAddress) (types.Claim, bool)
	SetClaim(claim types.Claim)
	GetGlobalIndexes(sourceID string) types.RewardIndexes
	SetGlobalIndexes(indexes types.RewardIndexes)
}

type Distributor struct {
	store       DistributorStore
	sourceGroup SourceGroup
}

func NewDistributor(store DistributorStore, sourceGroup SourceGroup) Distributor {
	d := Distributor{
		store:       store,
		sourceGroup: sourceGroup,
	}
	d.sourceGroup.RegisterSharesUpdateListener(d)
	return d
}

func (d Distributor) Distribute(sourceID string, amt sdk.Coins)                                {}
func (d Distributor) GetUserBalance(sourceID string, owner sdk.AccAddress) sdk.Coins           { return nil } // returns synced balance
func (d Distributor) WithdrawUserBalance(sourceID string, owner sdk.AccAddress, amt sdk.Coins) {}

// these could be sectioned off onto own private struct to avoid exposing them on public interface (as they're kind of dangerous)
// Or make these private, and pass funcs to RegisterListener methods.
func (d Distributor) SharesUpdated(sourceID string, owner sdk.AccAddress, oldShares sdk.Dec) {}
func (d Distributor) SharesCreated(sourceID string, owner sdk.AccAddress)                    {}

/*
Don't want to change Claim types (I think it would work, but non standard unmarshaling a proto type with a different type to what is was marshalled with).

But can add a new generic Claim type for Earn rewards (maybe don't call it Claim). Generic store method can translate into that type (shouldn't be interface, a distributor only knows about one claim thing).

Source is actually a group of sources of the same type. Can't easily generate a single source distributor when a param is added(?).
*/

// What about claiming?
type Claimer interface {
	Claim(ctx sdk.Context, sourceID string, opts types.Selections)
}

type SimplerMultiplierClaimer struct {
	rewardsWithdrawer rewardsWithdrawer // pass in distributor
	// needs access to params
	// needs service to send locked funds
}

type rewardsWithdrawer interface {
	GetUserBalance(sourceID string, owner sdk.AccAddress) sdk.Coins
	WithdrawUserBalance(sourceID string, owner sdk.AccAddress, amt sdk.Coins)
}

func (smc SimplerMultiplierClaimer) Claim(ctx sdk.Context, sourceID string, opts types.Selections) {
	// do normal claim stuff:
	// get user balance
	// verify selection
	// withdraw rewards from distributor

}

/*
Above creates "services" that run business logic. Better than what we have because they're smaller (easier to understand) and decoupled (easier to test and reuse).

Similar approach taken by swap pools - create "types" that don't use interfaces, but just perform calculations on internal state. Then we set them up by loading stuff from state, running, and writing back to state.
If the calculations are small then not worth doing.
*/
