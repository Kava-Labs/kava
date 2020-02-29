package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/kava-labs/kava/x/cdp/types"
)

// ApplySavingsRate distributes surplus that has accumulated in the liquidator account to address holding stable coins according the the savings rate
func (k Keeper) ApplySavingsRate(ctx sdk.Context, debtDenom string) sdk.Error {
	dp, found := k.GetDebtParam(ctx, debtDenom)
	if !found {
		return types.ErrDebtNotSupported(k.codespace, debtDenom)
	}
	savingsRateMacc := k.supplyKeeper.GetModuleAccount(ctx, types.SavingsRateMacc)

	surplusToDistribute := savingsRateMacc.GetCoins().AmountOf(dp.Denom)
	if surplusToDistribute.IsZero() {
		return nil
	}

	totalSurplusCoins := sdk.NewCoins(sdk.NewCoin(debtDenom, savingsRateMacc.GetCoins().AmountOf(dp.Denom)))
	totalSupplyLessSurplus := k.supplyKeeper.GetSupply(ctx).GetTotal().Sub(totalSurplusCoins)
	surplusDistributed := sdk.ZeroInt()
	var iterationErr sdk.Error
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
		interest := (sdk.NewDecFromInt(debtAmount).Mul(sdk.NewDecFromInt(surplusToDistribute))).Quo(sdk.NewDecFromInt(totalSupplyLessSurplus.AmountOf(debtDenom))).RoundInt()
		// sanity check, if we are going to over-distribute due to rounding, distribute only the remaining savings rate that hasn't been distributed.
		if interest.GT(surplusToDistribute.Sub(surplusDistributed)) {
			interest = surplusToDistribute.Sub(surplusDistributed)
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
		surplusDistributed = surplusDistributed.Add(interest)
		return false
	})
	if iterationErr != nil {
		return iterationErr
	}
	return nil
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
