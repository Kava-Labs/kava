package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

const (
	// factor for setting the initial value of gov tokens to sell at debt auctions -- assuming stable token is ~1 usd, this starts the auction with a price of $0.01 KAVA
	dump = 100
)

// AuctionCollateral creates auctions from the input deposits which attempt to raise the corresponding amount of debt
func (k Keeper) AuctionCollateral(ctx sdk.Context, deposits types.Deposits, debt sdk.Int, bidDenom string) error {

	auctionSize := k.getAuctionSize(ctx, deposits[0].Amount.Denom)
	totalCollateral := deposits.SumCollateral()
	for _, deposit := range deposits {

		debtCoveredByDeposit := (sdk.NewDecFromInt(deposit.Amount.Amount).Quo(sdk.NewDecFromInt(totalCollateral))).Mul(sdk.NewDecFromInt(debt)).RoundInt()
		err := k.CreateAuctionsFromDeposit(ctx, deposit.Amount, deposit.Depositor, debtCoveredByDeposit, auctionSize, bidDenom)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateAuctionsFromDeposit creates auctions from the input deposit
func (k Keeper) CreateAuctionsFromDeposit(
	ctx sdk.Context, collateral sdk.Coin, returnAddr sdk.AccAddress, debt, auctionSize sdk.Int,
	principalDenom string) (err error) {

	amountToAuction := collateral.Amount
	totalCollateralAmount := collateral.Amount
	remainingDebt := debt
	if !amountToAuction.IsPositive() {
		return nil
	}
	for amountToAuction.GT(auctionSize) {
		debtCoveredByAuction := (sdk.NewDecFromInt(auctionSize).Quo(sdk.NewDecFromInt(totalCollateralAmount))).Mul(sdk.NewDecFromInt(debt)).RoundInt()
		penalty := k.ApplyLiquidationPenalty(ctx, collateral.Denom, debtCoveredByAuction)
		_, err := k.auctionKeeper.StartCollateralAuction(
			ctx, types.LiquidatorMacc, sdk.NewCoin(collateral.Denom, auctionSize), sdk.NewCoin(principalDenom, debtCoveredByAuction.Add(penalty)), []sdk.AccAddress{returnAddr},
			[]sdk.Int{auctionSize}, sdk.NewCoin(k.GetDebtDenom(ctx), debtCoveredByAuction))
		if err != nil {
			return err
		}
		amountToAuction = amountToAuction.Sub(auctionSize)
		remainingDebt = remainingDebt.Sub(debtCoveredByAuction)
	}
	penalty := k.ApplyLiquidationPenalty(ctx, collateral.Denom, remainingDebt)
	_, err = k.auctionKeeper.StartCollateralAuction(
		ctx, types.LiquidatorMacc, sdk.NewCoin(collateral.Denom, amountToAuction), sdk.NewCoin(principalDenom, remainingDebt.Add(penalty)), []sdk.AccAddress{returnAddr},
		[]sdk.Int{amountToAuction}, sdk.NewCoin(k.GetDebtDenom(ctx), remainingDebt))
	if err != nil {
		return err
	}

	return nil
}

// NetSurplusAndDebt burns surplus and debt coins equal to the minimum of surplus and debt balances held by the liquidator module account
// for example, if there is 1000 debt and 100 surplus, 100 surplus and 100 debt are burned, netting to 900 debt
func (k Keeper) NetSurplusAndDebt(ctx sdk.Context) error {
	totalSurplus := k.GetTotalSurplus(ctx, types.LiquidatorMacc)
	debt := k.GetTotalDebt(ctx, types.LiquidatorMacc)
	netAmount := sdk.MinInt(totalSurplus, debt)
	if netAmount.IsZero() {
		return nil
	}

	// burn debt coins equal to netAmount
	err := k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(k.GetDebtDenom(ctx), netAmount)))
	if err != nil {
		return err
	}

	// burn stable coins equal to min(balance, netAmount)
	dp := k.GetParams(ctx).DebtParam
	balance := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc).GetCoins().AmountOf(dp.Denom)
	burnAmount := sdk.MinInt(balance, netAmount)
	return k.supplyKeeper.BurnCoins(ctx, types.LiquidatorMacc, sdk.NewCoins(sdk.NewCoin(dp.Denom, burnAmount)))
}

// GetTotalSurplus returns the total amount of surplus tokens held by the liquidator module account
func (k Keeper) GetTotalSurplus(ctx sdk.Context, accountName string) sdk.Int {
	acc := k.supplyKeeper.GetModuleAccount(ctx, accountName)
	totalSurplus := sdk.ZeroInt()
	dp := k.GetParams(ctx).DebtParam
	surplus := acc.GetCoins().AmountOf(dp.Denom)
	totalSurplus = totalSurplus.Add(surplus)
	return totalSurplus
}

// GetTotalDebt returns the total amount of debt tokens held by the liquidator module account
func (k Keeper) GetTotalDebt(ctx sdk.Context, accountName string) sdk.Int {
	acc := k.supplyKeeper.GetModuleAccount(ctx, accountName)
	debt := acc.GetCoins().AmountOf(k.GetDebtDenom(ctx))
	return debt
}

// RunSurplusAndDebtAuctions nets the surplus and debt balances and then creates surplus or debt auctions if the remaining balance is above the auction threshold parameter
func (k Keeper) RunSurplusAndDebtAuctions(ctx sdk.Context) error {
	if err := k.NetSurplusAndDebt(ctx); err != nil {
		return err
	}
	remainingDebt := k.GetTotalDebt(ctx, types.LiquidatorMacc)
	params := k.GetParams(ctx)
	if remainingDebt.GTE(params.DebtAuctionThreshold) {
		debtLot := sdk.NewCoin(k.GetDebtDenom(ctx), params.DebtAuctionLot)
		// debtAuctionLot * remainingDebt / remainingDebt
		debtCoveredByLot := (sdk.NewDecFromInt(params.DebtAuctionLot).Quo(sdk.NewDecFromInt(remainingDebt))).Mul(sdk.NewDecFromInt(remainingDebt)).RoundInt()
		bidCoin := sdk.NewCoin(params.DebtParam.Denom, debtCoveredByLot)
		_, err := k.auctionKeeper.StartDebtAuction(
			ctx, types.LiquidatorMacc, bidCoin, sdk.NewCoin(k.GetGovDenom(ctx), debtCoveredByLot.Mul(sdk.NewInt(dump))), debtLot)
		if err != nil {
			return err
		}
	}

	surplus := k.supplyKeeper.GetModuleAccount(ctx, types.LiquidatorMacc).GetCoins().AmountOf(params.DebtParam.Denom)
	if !surplus.GTE(params.SurplusAuctionThreshold) {
		return nil
	}
	surplusLot := sdk.NewCoin(params.DebtParam.Denom, sdk.MinInt(params.SurplusAuctionLot, surplus))
	_, err := k.auctionKeeper.StartSurplusAuction(ctx, types.LiquidatorMacc, surplusLot, k.GetGovDenom(ctx))
	return err
}
