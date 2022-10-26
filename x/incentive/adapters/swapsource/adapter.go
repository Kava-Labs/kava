package swapsource

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/distributor"
	"github.com/kava-labs/kava/x/incentive/types"
)

var _ distributor.SourceAdapter = &SwapAdapter{}

type SwapAdapter struct {
	swapKeeper types.SwapKeeper
	listener   distributor.SharesUpdateListener
}

func New(swapKeeper types.SwapKeeper) *SwapAdapter {
	sa := &SwapAdapter{
		swapKeeper: swapKeeper,
	}
	// keeper needs to be pointer, otherwise msgs won't trigger hooks
	// Also this wouldn't work for staking hooks, as it has multiple listeners registered. It would need an `AppendHooks` method.
	swapKeeper.SetHooks(sa)
	return sa
}

func (sa *SwapAdapter) GetTotalShares(ctx sdk.Context, sourceID string) sdk.Dec {
	shares, found := sa.swapKeeper.GetPoolShares(ctx, sourceID)
	if !found {
		shares = sdk.ZeroInt()
	}
	return shares.ToDec()
}

func (sa *SwapAdapter) GetShares(ctx sdk.Context, sourceID string, owner sdk.AccAddress) sdk.Dec {
	shares, found := sa.swapKeeper.GetDepositorSharesAmount(ctx, owner, sourceID)
	if !found {
		shares = sdk.ZeroInt()
	}
	return shares.ToDec()
}

func (sa *SwapAdapter) RegisterSharesUpdateListener(listener distributor.SharesUpdateListener) {
	sa.listener = listener
}

// ----- Implement SwapKeeper hooks -----

func (sa *SwapAdapter) AfterPoolDepositCreated(ctx sdk.Context, poolID string, depositor sdk.AccAddress, _ sdk.Int) {
	sa.listener.SharesCreated(ctx, poolID, depositor)
}

func (sa *SwapAdapter) BeforePoolDepositModified(ctx sdk.Context, poolID string, depositor sdk.AccAddress, sharesOwned sdk.Int) {
	sa.listener.SharesUpdated(ctx, poolID, depositor, sharesOwned.ToDec())
}
