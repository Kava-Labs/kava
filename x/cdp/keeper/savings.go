package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	"github.com/kava-labs/kava/x/cdp/types"
)

// DistributeSavingsRate distributes surplus that has accumulated in the liquidator account to address holding stable coins according the the savings rate
func (k Keeper) DistributeSavingsRate(ctx sdk.Context, debtDenom string) error {
	dp, found := k.GetDebtParam(ctx, debtDenom)
	if !found {
		return sdkerrors.Wrap(types.ErrDebtNotSupported, debtDenom)
	}
	savingsRateMacc := k.supplyKeeper.GetModuleAccount(ctx, types.SavingsRateMacc)
	surplusToDistribute := savingsRateMacc.GetCoins().AmountOf(dp.Denom)
	if surplusToDistribute.IsZero() {
		return nil
	}

	modAccountCoins := k.getModuleAccountCoins(ctx, dp.Denom)
	totalSupplyLessModAccounts := k.supplyKeeper.GetSupply(ctx).GetTotal().Sub(modAccountCoins)
	// values to use in interest calculation
	totalSurplus := sdk.NewDecFromInt(surplusToDistribute)
	totalSupply := sdk.NewDecFromInt(totalSupplyLessModAccounts.AmountOf(debtDenom))

	var iterationErr error
	k.accountKeeper.IterateAccounts(ctx, func(acc authexported.Account) (stop bool) {
		_, ok := acc.(supplyexported.ModuleAccountI)
		if ok {
			// don't distribute savings rate to module accounts
			return false
		}
		debtAmount := acc.GetCoins().AmountOf(debtDenom)
		if !debtAmount.IsPositive() {
			return false
		}
		// (balance * rewardToDisribute) /  totalSupply
		// interest is the ratable fraction of savings rate owed to that account, rounded using bankers rounding
		interest := (sdk.NewDecFromInt(debtAmount).Mul(totalSurplus)).Quo(totalSupply).RoundInt()
		// sanity check, if we are going to over-distribute due to rounding, distribute only the remaining savings rate that hasn't been distributed.
		if interest.GT(surplusToDistribute) {
			interest = surplusToDistribute
		}
		// sanity check - don't send saving rate if the rounded amount is zero
		if !interest.IsPositive() {
			return false
		}
		interestCoins := sdk.NewCoins(sdk.NewCoin(debtDenom, interest))
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.SavingsRateMacc, acc.GetAddress(), interestCoins)
		if err != nil {
			iterationErr = err
			return true
		}
		surplusToDistribute = surplusToDistribute.Sub(interest)
		return false
	})
	return iterationErr
}

// GetPreviousSavingsDistribution get the time of the previous savings rate distribution
func (k Keeper) GetPreviousSavingsDistribution(ctx sdk.Context) (distTime time.Time, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDistributionTimeKey)
	b := store.Get([]byte{})
	if b == nil {
		return time.Time{}, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &distTime)
	return distTime, true
}

// SetPreviousSavingsDistribution set the time of the previous block
func (k Keeper) SetPreviousSavingsDistribution(ctx sdk.Context, distTime time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PreviousDistributionTimeKey)
	store.Set([]byte{}, k.cdc.MustMarshalBinaryLengthPrefixed(distTime))
}

// getModuleAccountCoins gets the total coin balance of this coin currently held by module accounts
func (k Keeper) getModuleAccountCoins(ctx sdk.Context, denom string) sdk.Coins {
	totalModCoinBalance := sdk.NewCoins(sdk.NewCoin(denom, sdk.ZeroInt()))
	for macc := range k.maccPerms {
		modCoinBalance := k.supplyKeeper.GetModuleAccount(ctx, macc).GetCoins().AmountOf(denom)
		if modCoinBalance.IsPositive() {
			totalModCoinBalance = totalModCoinBalance.Add(sdk.NewCoin(denom, modCoinBalance))
		}
	}
	return totalModCoinBalance
}
