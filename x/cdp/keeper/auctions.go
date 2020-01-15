package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

const (
	// factor for setting the initial value of gov tokens to sell at debt auctions -- assuming stable token is ~1 usd, this starts the auction with a price of $0.01 KAVA
	dump = 100
)

type partialDeposit struct {
	Depositor sdk.AccAddress
	Amount    sdk.Coins
	DebtShare sdk.Int
}

func newPartialDeposit(depositor sdk.AccAddress, amount sdk.Coins, ds sdk.Int) partialDeposit {
	return partialDeposit{
		Depositor: depositor,
		Amount:    amount,
		DebtShare: ds,
	}
}

type partialDeposits []partialDeposit

func (pd partialDeposits) SumCollateral() (sum sdk.Int) {
	sum = sdk.ZeroInt()
	for _, d := range pd {
		sum = sum.Add(d.Amount[0].Amount)
	}
	return
}

func (pd partialDeposits) SumDebt() (sum sdk.Int) {
	sum = sdk.ZeroInt()
	for _, d := range pd {
		sum = sum.Add(d.DebtShare)
	}
	return
}

// AuctionCollateral creates auctions from the input deposits which attempt to raise the corresponding amount of debt
func (k Keeper) AuctionCollateral(ctx sdk.Context, deposits types.Deposits, debt sdk.Int, bidDenom string) sdk.Error {
	auctionSize := k.getAuctionSize(ctx, deposits[0].Amount[0].Denom)
	partialAuctionDeposits := partialDeposits{}
	totalCollateral := deposits.SumCollateral()
	for totalCollateral.GT(sdk.ZeroInt()) {
		for i, dep := range deposits {
			// create auctions from individual deposits that are larger than the auction size
			debtChange, collateralChange, err := k.CreateAuctionsFromDeposit(ctx, dep, debt, totalCollateral, auctionSize, bidDenom)
			if err != nil {
				return err
			}
			debt = debt.Sub(debtChange)
			totalCollateral = totalCollateral.Sub(collateralChange)
			dep.Amount = sdk.NewCoins(sdk.NewCoin(dep.Amount[0].Denom, dep.Amount[0].Amount.Sub(collateralChange)))
			// if there is leftover collateral that is less than a lot
			if !dep.Amount.IsZero() {
				collateralAmount := dep.Amount[0].Amount
				collateralDenom := dep.Amount[0].Denom
				// figure out how much debt this deposit accounts for
				// (depositCollateral / totalCollateral) * totalDebtFromCDP
				debtCoveredByDeposit := (collateralAmount.Quo(totalCollateral)).Mul(debt)
				// if adding this deposit to the other partial deposits is less than a lot
				if (partialAuctionDeposits.SumCollateral().Add(collateralAmount)).LT(auctionSize) {
					// append the deposit to the partial deposits and zero out the deposit
					pd := newPartialDeposit(dep.Depositor, dep.Amount, debtCoveredByDeposit)
					partialAuctionDeposits = append(partialAuctionDeposits, pd)
					dep.Amount = sdk.NewCoins(sdk.NewCoin(collateralDenom, sdk.ZeroInt()))
				} else {
					// if the sum of partial deposits now makes a lot
					partialCollateral := sdk.NewCoins(sdk.NewCoin(collateralDenom, auctionSize.Sub(partialAuctionDeposits.SumCollateral())))
					partialAmount := partialCollateral[0].Amount
					partialDebt := (partialAmount.Quo(collateralAmount)).Mul(debtCoveredByDeposit)

					// create a partial deposit from the deposit
					partialDep := newPartialDeposit(dep.Depositor, partialCollateral, partialDebt)
					// append it to the partial deposits
					partialAuctionDeposits = append(partialAuctionDeposits, partialDep)
					// create an auction from the partial deposits
					debtChange, collateralChange, err := k.CreateAuctionFromPartialDeposits(ctx, partialAuctionDeposits, debt, totalCollateral, auctionSize, bidDenom)
					if err != nil {
						return err
					}
					debt = debt.Sub(debtChange)
					totalCollateral = totalCollateral.Sub(collateralChange)
					// reset partial deposits and update the deposit amount
					partialAuctionDeposits = partialDeposits{}
					dep.Amount = sdk.NewCoins(sdk.NewCoin(collateralDenom, collateralAmount.Sub(partialAmount)))
				}
			}
			if dep.Amount.IsZero() {
				// remove the deposit from the slice if it is empty
				deposits = append(deposits[:i], deposits[i+1:]...)
				i--
			} else {
				deposits[i] = dep
			}
			totalCollateral = deposits.SumCollateral()
		}
	}
	if partialAuctionDeposits.SumCollateral().GT(sdk.ZeroInt()) {
		_, _, err := k.CreateAuctionFromPartialDeposits(ctx, partialAuctionDeposits, debt, totalCollateral, partialAuctionDeposits.SumCollateral(), bidDenom)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateAuctionsFromDeposit creates auctions from the input deposit until there is less than auctionSize left on the deposit
func (k Keeper) CreateAuctionsFromDeposit(ctx sdk.Context, dep types.Deposit, debt sdk.Int, totalCollateral sdk.Int, auctionSize sdk.Int, principalDenom string) (debtChange sdk.Int, collateralChange sdk.Int, err sdk.Error) {
	debtChange = sdk.ZeroInt()
	collateralChange = sdk.ZeroInt()
	depositAmount := dep.Amount[0].Amount
	depositDenom := dep.Amount[0].Denom
	for depositAmount.GTE(auctionSize) {
		// figure out how much debt is covered by one lots worth of collateral
		depositDebtAmount := (sdk.NewDecFromInt(auctionSize).Quo(sdk.NewDecFromInt(totalCollateral))).Mul(sdk.NewDecFromInt(debt)).RoundInt()
		// start an auction for one lot, attempting to raise depositDebtAmount
		_, err := k.auctionKeeper.StartCollateralAuction(
			ctx, types.LiquidatorMacc, sdk.NewCoin(depositDenom, auctionSize), sdk.NewCoin(principalDenom, depositDebtAmount), []sdk.AccAddress{dep.Depositor},
			[]sdk.Int{auctionSize}, sdk.NewCoin(k.GetDebtDenom(ctx), depositDebtAmount))
		if err != nil {
			return sdk.ZeroInt(), sdk.ZeroInt(), err
		}
		depositAmount = depositAmount.Sub(auctionSize)
		totalCollateral = totalCollateral.Sub(auctionSize)
		debt = debt.Sub(depositDebtAmount)
		// subtract one lot's worth of debt from the total debt covered by this deposit
		debtChange = debtChange.Add(depositDebtAmount)
		collateralChange = collateralChange.Add(auctionSize)

	}
	return debtChange, collateralChange, nil
}

// CreateAuctionFromPartialDeposits creates an auction from the input partial deposits
func (k Keeper) CreateAuctionFromPartialDeposits(ctx sdk.Context, partialDeps partialDeposits, debt sdk.Int, collateral sdk.Int, auctionSize sdk.Int, bidDenom string) (debtChange, collateralChange sdk.Int, err sdk.Error) {

	returnAddrs := []sdk.AccAddress{}
	returnWeights := []sdk.Int{}
	for _, pd := range partialDeps {
		returnAddrs = append(returnAddrs, pd.Depositor)
		returnWeights = append(returnWeights, pd.DebtShare)
	}
	_, err = k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorMacc, sdk.NewCoin(partialDeps[0].Amount[0].Denom, auctionSize), sdk.NewCoin(bidDenom, partialDeps.SumDebt()), returnAddrs, returnWeights, sdk.NewCoin(k.GetDebtDenom(ctx), partialDeps.SumDebt()))
	if err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), err
	}
	debtChange = partialDeps.SumDebt()
	collateralChange = partialDeps.SumCollateral()
	return debtChange, collateralChange, nil
}

// NetSurplusAndDebt burns surplus and debt coins equal to the minimum of surplus and debt balances held by the liquidator module account
// for example, if there is 1000 debt and 100 surplus, 100 surplus and 100 debt are burned, netting to 900 debt
func (k Keeper) NetSurplusAndDebt(ctx sdk.Context) sdk.Error {
	totalSurplus := k.GetTotalSurplus(ctx, types.LiquidatorMacc)
	debt := k.GetTotalDebt(ctx, types.LiquidatorMacc)
	netAmount := sdk.MinInt(totalSurplus, debt)
	if netAmount.GT(sdk.ZeroInt()) {
		surplusToBurn := netAmount
		err := k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(k.GetDebtDenom(ctx), netAmount)))
		if err != nil {
			return err
		}
		for surplusToBurn.GT(sdk.ZeroInt()) {
			for _, dp := range k.GetParams(ctx).DebtParams {
				balance := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc).GetCoins().AmountOf(dp.Denom)
				if balance.LT(netAmount) {
					err = k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, balance)))
					if err != nil {
						return err
					}
					surplusToBurn = surplusToBurn.Sub(balance)
				} else {
					err = k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, surplusToBurn)))
					if err != nil {
						return err
					}
					surplusToBurn = sdk.ZeroInt()
				}
			}
		}
	}
	return nil
}

// GetTotalSurplus returns the total amount of surplus tokens held by the liquidator module account
func (k Keeper) GetTotalSurplus(ctx sdk.Context, accountName string) sdk.Int {
	acc := k.supplyKeeper.GetModuleAccount(ctx, accountName)
	totalSurplus := sdk.ZeroInt()
	for _, dp := range k.GetParams(ctx).DebtParams {
		surplus := acc.GetCoins().AmountOf(dp.Denom)
		totalSurplus = totalSurplus.Add(surplus)
	}
	return totalSurplus
}

// GetTotalDebt returns the total amount of debt tokens held by the liquidator module account
func (k Keeper) GetTotalDebt(ctx sdk.Context, accountName string) sdk.Int {
	acc := k.supplyKeeper.GetModuleAccount(ctx, accountName)
	debt := acc.GetCoins().AmountOf(k.GetDebtDenom(ctx))
	return debt
}

// RunSurplusAndDebtAuctions nets the surplus and debt balances and then creates surplus or debt auctions if the remaining balance is above the auction threshold parameter
func (k Keeper) RunSurplusAndDebtAuctions(ctx sdk.Context) sdk.Error {
	k.NetSurplusAndDebt(ctx)
	remainingDebt := k.GetTotalDebt(ctx, types.LiquidatorMacc)
	params := k.GetParams(ctx)
	if remainingDebt.GTE(params.DebtAuctionThreshold) {
		_, err := k.auctionKeeper.StartDebtAuction(ctx, types.LiquidatorMacc, sdk.NewCoin("usdx", remainingDebt), sdk.NewCoin(k.GetGovDenom(ctx), remainingDebt.Mul(sdk.NewInt(dump))), sdk.NewCoin(k.GetDebtDenom(ctx), remainingDebt))
		if err != nil {
			return err
		}
	}
	remainingSurplus := k.GetTotalSurplus(ctx, types.LiquidatorMacc)
	if remainingSurplus.GTE(params.SurplusAuctionThreshold) {
		for _, dp := range params.DebtParams {
			surplusLot := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc).GetCoins().AmountOf(dp.Denom)
			_, err := k.auctionKeeper.StartSurplusAuction(ctx, types.LiquidatorMacc, sdk.NewCoin(dp.Denom, surplusLot), k.GetGovDenom(ctx))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
