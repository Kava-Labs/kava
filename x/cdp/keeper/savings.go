package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	auctiontypes "github.com/kava-labs/kava/x/auction/types"
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
	surplusDistributed := sdk.ZeroInt()
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
		interest := (sdk.NewDecFromInt(debtAmount).Mul(sdk.NewDecFromInt(surplusToDistribute))).Quo(sdk.NewDecFromInt(totalSupplyLessModAccounts.AmountOf(debtDenom))).RoundInt()
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

func (k Keeper) getModuleAccountCoins(ctx sdk.Context, denom string) sdk.Coins {
	// NOTE: these are the module accounts that could end up holding stable denoms at some point.
	// Since there are currently no api methods to 'GetAllModuleAccounts', this function will need to be updated if a
	// new module account is added which can hold stable denoms.
	savingsRateMaccCoinAmount := k.supplyKeeper.GetModuleAccount(ctx, types.SavingsRateMacc).GetCoins().AmountOf(denom)
	cdpMaccCoinAmount := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName).GetCoins().AmountOf(denom)
	auctionMaccCoinAmount := k.supplyKeeper.GetModuleAccount(ctx, auctiontypes.ModuleName).GetCoins().AmountOf(denom)
	liquidatorMaccCoinAmount := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc).GetCoins().AmountOf(denom)
	feeMaccCoinAmount := k.supplyKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName).GetCoins().AmountOf(denom)
	totalModAccountAmount := savingsRateMaccCoinAmount.Add(cdpMaccCoinAmount).Add(auctionMaccCoinAmount).Add(liquidatorMaccCoinAmount).Add(feeMaccCoinAmount)
	return sdk.NewCoins(sdk.NewCoin(denom, totalModAccountAmount))
}
