package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	banktypes "github.com/kava-labs/kava/x/bank"
	"github.com/kava-labs/kava/x/cdp/types"
)

// AccumulateUSDXSavings updates the global savings rate factor based on the amount of USDX savings accumulated
func (k Keeper) AccumulateUSDXSavings(ctx sdk.Context, amount sdk.Int) error {

	previousSavingsFactor, found := k.GetSavingsRateFactor(ctx)
	if !found {
		k.SetSavingsRateFactor(ctx, sdk.ZeroDec())
		return nil
	}
	if amount.IsZero() {
		return nil
	}
	usdxSupplyAmount := k.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(types.DefaultStableDenom)
	savingsMaccBalance := k.supplyKeeper.GetModuleAccount(ctx, types.SavingsRateMacc).GetCoins().AmountOf(types.DefaultStableDenom)
	usdxSupplyAmount = usdxSupplyAmount.Sub(savingsMaccBalance)
	if usdxSupplyAmount.IsZero() {
		return nil
	}

	currentFactor := amount.ToDec().Quo(usdxSupplyAmount.ToDec())
	newFactor := previousSavingsFactor.Add(currentFactor)
	k.SetSavingsRateFactor(ctx, newFactor)
	return nil

}

// SyncUSDXSavingsRateBeforeTransfer syncs the usdx savings rate for the sender and receiver accounts
func (k Keeper) SyncUSDXSavingsRateBeforeTransfer(ctx sdk.Context, sender, receiver sdk.AccAddress, amount sdk.Coins) error {
	shouldSync, shouldExit := k.validateSavingsRateSender(ctx, sender, amount)
	if shouldExit {
		return nil
	}
	if shouldSync {
		err := k.SyncSavingsRate(ctx, sender, amount)
		if err != nil {
			return err
		}
	}
	shouldSync = k.validateSavingsRateReceiver(ctx, receiver, amount)
	if shouldSync {
		err := k.SyncSavingsRate(ctx, receiver, amount)
		if err != nil {
			return err
		}
	}
	return nil
}

// SyncSavingsRate syncs the input address with the global savings rate, transfering any usdx that has accumulated
func (k Keeper) SyncSavingsRate(ctx sdk.Context, addr sdk.AccAddress, amount sdk.Coins) error {
	globalSavingsFactor, found := k.GetSavingsRateFactor(ctx)
	if !found {
		globalSavingsFactor = sdk.ZeroDec()
	}
	claim, found := k.GetSavingsRateClaim(ctx, addr)
	if !found {
		claim = types.NewUSDXSavingsRateClaim(addr, globalSavingsFactor)
		k.SetSavingRateClaim(ctx, claim)
		return nil
	}
	userFactor := globalSavingsFactor.Sub(claim.Factor)
	if userFactor.IsZero() {
		return nil
	}
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil
	}
	savingsAccumulated := acc.GetCoins().AmountOf(types.DefaultStableDenom).ToDec().Mul(userFactor)
	if savingsAccumulated.IsPositive() {
		savingsAccumulatedCoin := sdk.NewCoin(types.DefaultStableDenom, savingsAccumulated.RoundInt())
		savingsMacc := k.supplyKeeper.GetModuleAccount(ctx, types.SavingsRateMacc)
		err := k.ValidateBalance(ctx, savingsAccumulatedCoin, savingsMacc.GetAddress())
		if err != nil {
			return err
		}
		err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.SavingsRateMacc, addr, sdk.NewCoins(savingsAccumulatedCoin))
		if err != nil {
			return err
		}
	}
	claim.Factor = globalSavingsFactor
	k.SetSavingRateClaim(ctx, claim)
	return nil
}

func (k Keeper) validateSavingsRateSender(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Coins) (bool, bool) {
	acc := k.accountKeeper.GetAccount(ctx, sender)
	if acc == nil {
		return false, false
	}
	// don't accumulate USDX savings to module accounts
	mAcc, ok := acc.(supplyexported.ModuleAccountI)
	if ok {
		// exit if the sender is the savings rate account
		if mAcc.GetName() == types.SavingsRateMacc {
			return false, true
		}
		return false, false
	}
	// only sync savings rate if transaction involves usdx
	if amount.AmountOf(types.DefaultStableDenom).GT(sdk.ZeroInt()) {
		return true, false
	}
	return false, false
}

func (k Keeper) validateSavingsRateReceiver(ctx sdk.Context, receiver sdk.AccAddress, amount sdk.Coins) bool {
	acc := k.accountKeeper.GetAccount(ctx, receiver)
	if acc == nil {
		return false
	}
	_, ok := acc.(supplyexported.ModuleAccountI)
	if ok {
		return false
	}
	if amount.AmountOf(types.DefaultStableDenom).GT(sdk.ZeroInt()) {
		return true
	}
	return false
}

// SyncUSDXSavingsRateMultiSend syncs the usdx savings rate for the input (sending) and output (receiving) accounts
func (k Keeper) SyncUSDXSavingsRateMultiSend(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	return nil
}

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
	// TODO: avoid iterating over all the accounts by keeping the stored stable coin
	// holders' addresses separately.
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
		interest = sdk.MinInt(interest, surplusToDistribute)

		// sanity check - don't send saving rate if the rounded amount is zero
		if !interest.IsPositive() {
			return false
		}

		// update total savings rate distributed by surplus to distribute
		previousSavingsDistributed := k.GetSavingsRateDistributed(ctx)
		newTotalDistributed := previousSavingsDistributed.Add(interest)
		k.SetSavingsRateDistributed(ctx, newTotalDistributed)

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
