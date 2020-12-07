package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/harvest/types"
)

// AttemptIndexLiquidations attempts to liquidate the lowest LTV borrows
func (k Keeper) AttemptIndexLiquidations(ctx sdk.Context) error {
	// use moneyMarketCache := map[string]types.MoneyMarket{}

	// Iterate over index
	//		Get borrower's address
	//		Use borrower's address to fetch borrow object
	//		Calculate outstanding interest and add to borrow balances
	//		Use current asset prices from pricefeed to calculate current LTV for each asset
	//		If LTV of any asset is over the max, liquidate it by
	//			Sending coins to auction module
	//			(?) Removing borrow from the store
	//			(?) Removing borrow LTV from LTV index

	return nil
}

// AttemptKeeperLiquidation enables a keeper to liquidate an individual borrower's position
func (k Keeper) AttemptKeeperLiquidation(ctx sdk.Context, keeper sdk.AccAddress, borrower sdk.AccAddress) error {
	// Calculate outstanding interest and add to borrow balances
	borrowBalances := k.GetBorrowBalance(ctx, borrower)

	// Load a list of user's deposit coin denoms, storing them in an sdk.Coins object
	deposits := k.GetDepositsByUser(ctx, borrower)
	if len(deposits) == 0 {
		return sdkerrors.Wrapf(types.ErrDepositsNotFound, "no deposits found for %s", borrower)
	}
	depositDenoms := sdk.Coins{}
	for _, deposit := range deposits {
		depositDenoms = append(depositDenoms, sdk.NewCoin(deposit.Amount.Denom, sdk.OneInt()))
	}

	// Build map of {denom} -> {liquidation data}
	type liqData struct {
		price            sdk.Dec
		ltv              sdk.Dec
		conversionFactor sdk.Int
	}
	liqMap := make(map[string]liqData)

	// Load required liquidation data for every deposit/borrow denom
	for _, coin := range borrowBalances.Add(depositDenoms...) {
		mm, found := k.GetMoneyMarket(ctx, coin.Denom)
		if !found {
			return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", coin.Denom)
		}

		priceData, err := k.pricefeedKeeper.GetCurrentPrice(ctx, mm.SpotMarketID)
		if err != nil {
			return err
		}

		liqMap[coin.Denom] = liqData{priceData.Price, mm.BorrowLimit.LoanToValue, mm.ConversionFactor}
	}

	totalBorrowableUSDAmount := sdk.ZeroDec()
	for _, deposit := range deposits {
		lData := liqMap[deposit.Amount.Denom]
		depositUSDValue := sdk.NewDecFromInt(deposit.Amount.Amount).Quo(sdk.NewDecFromInt(lData.conversionFactor)).Mul(lData.price)
		borrowableUSDAmountForDeposit := depositUSDValue.Mul(lData.ltv)
		totalBorrowableUSDAmount = totalBorrowableUSDAmount.Add(borrowableUSDAmountForDeposit)
	}

	totalBorrowedUSDAmount := sdk.ZeroDec()
	for _, coin := range borrowBalances {
		lData := liqMap[coin.Denom]
		usdValue := sdk.NewDecFromInt(coin.Amount).Quo(sdk.NewDecFromInt(lData.conversionFactor)).Mul(lData.price)
		totalBorrowedUSDAmount = totalBorrowedUSDAmount.Add(usdValue)
	}

	// Validate that the proposed borrow's USD value is within user's borrowable limit
	if totalBorrowedUSDAmount.LTE(totalBorrowableUSDAmount) {
		// return err this position is not over LTV
	}

	// Sending coins to auction module with keeper address getting 5% of the profits
	err := k.SeizeDeposits(ctx, deposits, keeper, k.GetKeeperRewardPercentage(ctx))
	if err != nil {
		return err
	}

	return nil
}

// SeizeDeposits seizes a list of deposits and sends them to auction
func (k Keeper) SeizeDeposits(ctx sdk.Context, deposits []types.Deposit, keeper sdk.AccAddress, rewardPercentage sdk.Dec) error {
	for _, deposit := range deposits {
		keeperReward := rewardPercentage.MulInt(deposit.Amount.Amount).TruncateInt()
		keeperCoin := sdk.NewCoin(deposit.Amount.Denom, keeperReward) // TODO: will this cause dust
		auctionCoin := sdk.NewCoin(deposit.Amount.Denom, deposit.Amount.Amount.Sub(keeperReward))

		// Send auction amount to liquidation module account
		err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorMacc, sdk.NewCoins(auctionCoin))
		if err != nil {
			return err
		}

		// Send keeper their reward
		err = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorMacc, sdk.NewCoins(keeperCoin))
		if err != nil {
			return err
		}

		err = k.AuctionDeposit(ctx, deposit)
		if err != nil {
			return err
		}

		k.DeleteDeposit(ctx, deposit)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDepositLiquidation,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyDepositor, deposit.Depositor.String()),
				sdk.NewAttribute(types.AttributeKeyDepositCoins, deposit.Amount.String()),
			),
		)
	}
	return nil
}

// AuctionDeposit starts auction(s) for an individual deposit
func (k Keeper) AuctionDeposit(ctx sdk.Context, deposit types.Deposit) error {
	mm, _ := k.GetMoneyMarket(ctx, deposit.Amount.Denom)
	remainingAmount := deposit.Amount.Amount
	for remainingAmount.GT(mm.AuctionSize) {
		// _, err := k.auctionKeeper.StartCollateralAuction(
		// 	ctx, types.LiquidatorMacc, sdk.NewCoin(deposit.Amount.Denom, mm.AuctionSize),
		// 	sdk.NewCoin(principalDenom, debtAmount.Add(penalty)), []sdk.AccAddress{deposit.Depositor},
		// 	[]sdk.Int{mm.AuctionSize}, sdk.NewCoin(debtDenom, debtAmount),
		// )
		// if err != nil {
		// 	return err
		// }
		remainingAmount = remainingAmount.Sub(mm.AuctionSize)
	}

	// _, err := k.auctionKeeper.StartCollateralAuction(
	// 	ctx, types.LiquidatorMacc, sdk.NewCoin(deposit.Amount.Denom, remainingAmount),
	// 	sdk.NewCoin(principalDenom, debtAmount.Add(penalty)), []sdk.AccAddress{deposit.Depositor},
	// 	[]sdk.Int{remainingAmount}, sdk.NewCoin(debtDenom, debtAmount),
	// )
	// if err != nil {
	// 	return err
	// }

	return nil
}
