package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
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
func (k Keeper) AuctionCollateral(ctx sdk.Context, deposits types.Deposits, debt sdk.Int, bidDenom string) {
	auctionSize := k.getAuctionSize(ctx, deposits[0].Amount[0].Denom)
	partialAuctionDeposits := partialDeposits{}
	totalCollateral := deposits.SumCollateral()
	for deposits.SumCollateral().GTE(auctionSize) {
		for i, dep := range deposits {
			// create auctions from individual deposits that are larger than the auction size
			k.CreateAuctionsFromDeposit(ctx, &dep, &debt, &totalCollateral, auctionSize, bidDenom)
			if !dep.Amount[0].Amount.IsZero() {
				debtCoveredByDeposit := (dep.Amount[0].Amount.Quo(totalCollateral)).Mul(debt)
				if (partialAuctionDeposits.SumCollateral().Add(dep.Amount[0].Amount)).LT(auctionSize) {
					pd := newPartialDeposit(dep.Depositor, dep.Amount, debtCoveredByDeposit)
					partialAuctionDeposits = append(partialAuctionDeposits, pd)
					dep.Amount = sdk.NewCoins(sdk.NewCoin(dep.Amount[0].Denom, sdk.ZeroInt()))
				} else {
					partialCollateral := sdk.NewCoins(sdk.NewCoin(dep.Amount[0].Denom, auctionSize.Sub(partialAuctionDeposits.SumCollateral())))
					partialDebt := (partialCollateral[0].Amount.Quo(dep.Amount[0].Amount)).Mul(debtCoveredByDeposit)

					partialDep := newPartialDeposit(dep.Depositor, partialCollateral, partialDebt)
					partialAuctionDeposits = append(partialAuctionDeposits, partialDep)
					k.CreateAuctionFromPartialDeposits(ctx, partialAuctionDeposits, &debt, &totalCollateral, auctionSize, bidDenom)
					partialAuctionDeposits = partialDeposits{}
					dep.Amount = sdk.NewCoins(sdk.NewCoin(dep.Amount[0].Denom, dep.Amount[0].Amount.Sub(partialCollateral[0].Amount)))
				}
			}
			if dep.Amount.IsZero() {
				// remove the deposit from the slice if it is empty
				deposits = append(deposits[:i], deposits[i+1:]...)
				i--
			} else {
				deposits[i] = dep
			}
		}
	}
	if partialAuctionDeposits.SumCollateral().GT(sdk.ZeroInt()) {
		k.CreateAuctionFromPartialDeposits(ctx, partialAuctionDeposits, &debt, &totalCollateral, partialAuctionDeposits.SumCollateral(), bidDenom)
	}
}

// CreateAuctionsFromDeposit creates auctions from the input deposit until there is less than auctionSize left on the deposit
func (k Keeper) CreateAuctionsFromDeposit(ctx sdk.Context, dep *types.Deposit, debt *sdk.Int, totalCollateral *sdk.Int, auctionSize sdk.Int, principalDenom string) {
	for dep.Amount[0].Amount.GTE(auctionSize) {
		// figure out how much debt is covered by one lots worth of collateral
		depositDebtAmount := (sdk.NewDecFromInt(auctionSize).Quo(sdk.NewDecFromInt(*totalCollateral))).Mul(sdk.NewDecFromInt(*debt)).RoundInt()
		// start an auction for one lot, attempting to raise depositDebtAmount
		_, err := k.auctionKeeper.StartCollateralAuction(
			ctx, types.LiquidatorMacc, sdk.NewCoin(dep.Amount[0].Denom, auctionSize), sdk.NewCoin(principalDenom, depositDebtAmount), []sdk.AccAddress{dep.Depositor},
			[]sdk.Int{auctionSize}, sdk.NewCoin(k.GetDebtDenom(ctx), depositDebtAmount))
		if err != nil {
			panic(err)
		}
		// subtract one lot's worth of debt from the total debt covered by this deposit
		*debt = debt.Sub(depositDebtAmount)
		// update the deposits collateral amount
		dep.Amount[0].Amount = dep.Amount[0].Amount.Sub(auctionSize)
		// update the total collateral
		*totalCollateral = totalCollateral.Sub(auctionSize)
	}
}

// CreateAuctionFromPartialDeposits creates an auction from the input partial deposits
func (k Keeper) CreateAuctionFromPartialDeposits(ctx sdk.Context, partialDeps partialDeposits, debt *sdk.Int, collateral *sdk.Int, auctionSize sdk.Int, bidDenom string) {
	returnAddrs := []sdk.AccAddress{}
	returnWeights := []sdk.Int{}
	for _, pd := range partialDeps {
		returnAddrs = append(returnAddrs, pd.Depositor)
		returnWeights = append(returnWeights, pd.DebtShare)
	}
	_, err := k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorMacc, sdk.NewCoin(partialDeps[0].Amount[0].Denom, auctionSize), sdk.NewCoin(bidDenom, partialDeps.SumDebt()), returnAddrs, returnWeights, sdk.NewCoin(k.GetDebtDenom(ctx), partialDeps.SumDebt()))
	if err != nil {
		panic(err)
	}
	*debt = debt.Sub(partialDeps.SumDebt())
	*collateral = collateral.Sub(partialDeps.SumCollateral())
}

// NetSurplusAndDebt burns surplus and debt coins equal to the minimum of surplus and debt balances held by the liquidator module account
// for example, if there is 1000 debt and 100 surplus, 100 surplus and 100 debt are burned, netting to 900 debt
func (k Keeper) NetSurplusAndDebt(ctx sdk.Context) {
	totalSurplus := k.GetTotalSurplus(ctx, types.LiquidatorMacc)
	debt := k.GetTotalDebt(ctx, types.LiquidatorMacc)
	netAmount := sdk.MinInt(totalSurplus, debt)
	if netAmount.GT(sdk.ZeroInt()) {
		surplusToBurn := netAmount
		err := k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(k.GetDebtDenom(ctx), netAmount)))
		if err != nil {
			panic(err)
		}
		for _, dp := range k.GetParams(ctx).DebtParams {
			for surplusToBurn.GT(sdk.ZeroInt()) {
				balance := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc).GetCoins().AmountOf(dp.Denom)
				if balance.LT(netAmount) {
					err = k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, balance)))
					if err != nil {
						panic(err)
					}
					surplusToBurn = surplusToBurn.Sub(balance)
				} else {
					err = k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, surplusToBurn)))
					if err != nil {
						panic(err)
					}
					surplusToBurn = sdk.ZeroInt()
				}
			}
		}
	}
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

// HandleSurplusAndDebtAuctions nets the surplus and debt balances and then creates surplus or debt auctions if the remaining balance is above the auction threshold parameter
func (k Keeper) HandleSurplusAndDebtAuctions(ctx sdk.Context) {
	k.NetSurplusAndDebt(ctx)
	remainingDebt := k.GetTotalDebt(ctx, types.LiquidatorMacc)
	params := k.GetParams(ctx)
	if remainingDebt.GTE(params.DebtAuctionThreshold) {
		_, err := k.auctionKeeper.StartDebtAuction(ctx, types.LiquidatorMacc, sdk.NewCoin("usdx", sdk.NewInt(1)), sdk.NewCoin(k.GetGovDenom(ctx), remainingDebt.Mul(sdk.NewInt(100))), sdk.NewCoin(k.GetDebtDenom(ctx), remainingDebt))
		if err != nil {
			panic(err)
		}
	}
	remainingSurplus := k.GetTotalSurplus(ctx, types.LiquidatorMacc)
	if remainingSurplus.GTE(params.SurplusAuctionThreshold) {
		for _, dp := range params.DebtParams {
			surplusLot := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc).GetCoins().AmountOf(dp.Denom)
			_, err := k.auctionKeeper.StartSurplusAuction(ctx, types.LiquidatorMacc, sdk.NewCoin(dp.Denom, surplusLot), k.GetGovDenom(ctx))
			if err != nil {
				panic(err)
			}
		}
	}
}
