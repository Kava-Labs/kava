package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/kava-labs/kava/x/cdp/types"
)

// ApplySavingsRate distributes surplus that has accumulated in the liquidator account to address holding stable coins according the the savings rate
func (k Keeper) ApplySavingsRate(ctx sdk.Context, debtDenom string) sdk.Error {
	dp, found := k.GetDebtParam(ctx, debtDenom)
	if !found {
		return types.ErrDebtNotSupported(k.codespace, debtDenom)
	}
	liquidatorMacc := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc)
	totalSurplusCoins := sdk.NewCoins(sdk.NewCoin(debtDenom, liquidatorMacc.GetCoins().AmountOf(dp.Denom)))

	surplusToDistribute := sdk.NewDecFromInt(liquidatorMacc.GetCoins().AmountOf(dp.Denom)).Mul(dp.SavingsRate).RoundInt()
	if surplusToDistribute.IsZero() {
		return nil
	}
	totalSupplyLessSurplus := k.supplyKeeper.GetSupply(ctx).GetTotal().Sub(totalSurplusCoins)
	surplusDistributed := sdk.ZeroInt()
	for _, acc := range k.accountKeeper.GetAllAccounts(ctx) {
		_, ok := acc.(supplyexported.ModuleAccountI)
		if ok {
			// don't distribute rewards to module accounts
			continue
		}
		debtAmount := acc.GetCoins().AmountOf(debtDenom)
		if debtAmount.IsPositive() {
			// (balance / totalSupply) * rewardToDisribute
			// interest is the ratable fraction of rewards owed to that account, rounded using bankers rounding
			interest := (sdk.NewDecFromInt(debtAmount).Quo(sdk.NewDecFromInt(totalSupplyLessSurplus.AmountOf(debtDenom)))).Mul(sdk.NewDecFromInt(surplusToDistribute)).RoundInt()
			// sanity check, if we are going to over-distribute due to rounding, distribute only the remaining rewards that haven't been distributed.
			if interest.GT(surplusToDistribute.Sub(surplusDistributed)) {
				interest = surplusToDistribute.Sub(surplusDistributed)
			}
			// sanity check - don't send rewards if the rounded reward is zero
			if interest.IsPositive() {
				interestCoins := sdk.NewCoins(sdk.NewCoin(debtDenom, interest))
				err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.LiquidatorMacc, acc.GetAddress(), interestCoins)
				if err != nil {
					return err
				}
				surplusDistributed = surplusDistributed.Add(interest)
			}
		}
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
