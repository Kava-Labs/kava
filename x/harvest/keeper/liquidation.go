package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/harvest/types"
)

// LiqData holds liquidation-related data
type LiqData struct {
	price            sdk.Dec
	ltv              sdk.Dec
	conversionFactor sdk.Int
}

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
	// Fetch deposits and parse coin denoms
	deposits := k.GetDepositsByUser(ctx, borrower)
	depositDenoms := []string{}
	for _, deposit := range deposits {
		depositDenoms = append(depositDenoms, deposit.Amount.Denom)
	}

	// Fetch borrow balances and parse coin denoms
	borrowBalances := k.GetBorrowBalance(ctx, borrower)
	borrowDenoms := getDenoms(borrowBalances)

	liqMap := make(map[string]LiqData)

	// Load required liquidation data for every deposit/borrow denom
	denoms := removeDuplicates(borrowDenoms, depositDenoms)
	for _, denom := range denoms {
		mm, found := k.GetMoneyMarket(ctx, denom)
		if !found {
			return sdkerrors.Wrapf(types.ErrMarketNotFound, "no market found for denom %s", denom)
		}

		priceData, err := k.pricefeedKeeper.GetCurrentPrice(ctx, mm.SpotMarketID)
		if err != nil {
			return err
		}

		liqMap[denom] = LiqData{priceData.Price, mm.BorrowLimit.LoanToValue, mm.ConversionFactor}
	}

	totalBorrowableUSDAmount := sdk.ZeroDec()
	totalDepositedUSDAmount := sdk.ZeroDec()
	for _, deposit := range deposits {
		lData := liqMap[deposit.Amount.Denom]
		usdValue := sdk.NewDecFromInt(deposit.Amount.Amount).Quo(sdk.NewDecFromInt(lData.conversionFactor)).Mul(lData.price)
		totalDepositedUSDAmount = totalDepositedUSDAmount.Add(usdValue)
		borrowableUSDAmountForDeposit := usdValue.Mul(lData.ltv)
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
		return sdkerrors.Wrapf(types.ErrBorrowNotLiquidatable, "borrowed %s <= borrowable %s", totalBorrowedUSDAmount, totalBorrowableUSDAmount)
	}

	// Sending coins to auction module with keeper address getting % of the profits
	borrow, _ := k.GetBorrow(ctx, borrower)
	err := k.SeizeDeposits(ctx, keeper, liqMap, deposits, borrowBalances, depositDenoms, borrowDenoms)
	if err != nil {
		return err
	}

	k.DeleteBorrow(ctx, borrow)

	// TODO: these 'oldDeposits' can still have funds in them. Where should we send the extra balance?
	for _, oldDeposit := range deposits {
		k.DeleteDeposit(ctx, oldDeposit)
	}

	return nil
}

// SeizeDeposits seizes a list of deposits and sends them to auction
func (k Keeper) SeizeDeposits(ctx sdk.Context, keeper sdk.AccAddress, liqMap map[string]LiqData,
	deposits []types.Deposit, borrowBalances sdk.Coins, dDenoms, bDenoms []string) error {

	// Seize % of every deposit and send to the keeper
	aucDeposits := sdk.Coins{}
	for _, deposit := range deposits {
		mm, _ := k.GetMoneyMarket(ctx, deposit.Amount.Denom)
		keeperReward := mm.KeeperRewardPercentage.MulInt(deposit.Amount.Amount).TruncateInt()
		keeperCoin := sdk.NewCoin(deposit.Amount.Denom, keeperReward)

		// Send keeper their reward
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, keeper, sdk.NewCoins(keeperCoin))
		if err != nil {
			return err
		}

		// Add remaining deposit coin to aucDeposits
		aucDeposits = aucDeposits.Add(sdk.NewCoin(deposit.Amount.Denom, deposit.Amount.Amount.Sub(keeperReward)))
	}

	// Build map to hold deposit coin USD valuations
	totalRemainingDepositedUSDAmount := sdk.ZeroDec()
	dUsdMap := make(map[string]sdk.Dec)
	for _, deposit := range aucDeposits {
		dData := liqMap[deposit.Denom]
		dCoinUsdValue := sdk.NewDecFromInt(deposit.Amount).Quo(sdk.NewDecFromInt(dData.conversionFactor)).Mul(dData.price)
		totalRemainingDepositedUSDAmount = totalRemainingDepositedUSDAmount.Add(dCoinUsdValue)
		dUsdMap[deposit.Denom] = dCoinUsdValue
	}

	// Build map to hold borrow coin USD valuations
	totalBorrowedUSDAmount := sdk.ZeroDec()
	bUsdMap := make(map[string]sdk.Dec)
	for _, bCoin := range borrowBalances {
		bData := liqMap[bCoin.Denom]
		bCoinUsdValue := sdk.NewDecFromInt(bCoin.Amount).Quo(sdk.NewDecFromInt(bData.conversionFactor)).Mul(bData.price)
		totalBorrowedUSDAmount = totalBorrowedUSDAmount.Add(bCoinUsdValue)
		bUsdMap[bCoin.Denom] = bCoinUsdValue
	}

	// The % by which the lot must be larger than the borrow
	ltv := totalBorrowedUSDAmount.Quo(totalRemainingDepositedUSDAmount)

	err := k.StartAuctions(ctx, deposits[0].Depositor, bDenoms, dDenoms, borrowBalances, aucDeposits, ltv, liqMap, bUsdMap, dUsdMap)
	if err != nil {
		return err
	}

	return nil
}

// StartAuctions attempts to start auctions for seized assets
func (k Keeper) StartAuctions(ctx sdk.Context, borrower sdk.AccAddress, borrowDenoms, depositDenoms []string,
	borrows, deposits sdk.Coins, ltv sdk.Dec, liqMap map[string]LiqData, bUsdMap, dUsdMap map[string]sdk.Dec) error {
	// Set up auction constants
	returnAddrs := []sdk.AccAddress{borrower}
	weights := []sdk.Int{sdk.NewInt(100)}
	debt := sdk.NewCoin("debt", sdk.ZeroInt())

	// 1. Attempt auctions where we can sell all of a borrowed asset type at once
	for _, bDenom := range borrowDenoms {
		bCoin := sdk.NewCoin(bDenom, borrows.AmountOf(bDenom))

		for _, dDenom := range depositDenoms {
			// Search for a deposit coin amount with USD valuation >= desired lot size USD valuation
			lotSizeUSD := bUsdMap[bDenom].Mul(ltv)
			if dUsdMap[dDenom].GTE(lotSizeUSD) {

				// Convert lot size USD to lot size native currency
				lotSizeNative := lotSizeUSD.MulInt(liqMap[dDenom].conversionFactor).Quo(liqMap[dDenom].price)
				lot := sdk.NewCoin(dDenom, lotSizeNative.TruncateInt())
				bid := bCoin

				// Start auction with this lot (deposit) and bid (borrow)
				err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorAccount, sdk.NewCoins(lot))
				if err != nil {
					return err
				}
				_, err = k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorAccount, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return err
				}

				// Adjust remaining value of remaining USD to be auctioned
				bUsdMap[bDenom] = sdk.ZeroDec()
				dUsdMap[dDenom] = dUsdMap[dDenom].Sub(lotSizeUSD)
				// Adjust amount of remaining bids/lots in native currencies
				borrows = borrows.Sub(sdk.NewCoins(bCoin))
				deposits = deposits.Sub(sdk.NewCoins(lot))
				break // No more borrow balance left for this denom, move to next borrow denom
			}
		}
	}

	// 2. Attempt auctions where we can sell all of a deposited asset type at once
	for _, dDenom := range depositDenoms {
		dCoin := sdk.NewCoin(dDenom, deposits.AmountOf(dDenom))

		// At this bid amount we'll sell all the collateral at a (1/ltv) ratio
		bidSize := dUsdMap[dDenom].Mul(ltv)
		for _, bDenom := range borrowDenoms {
			if bUsdMap[bDenom].GTE(bidSize) {
				// Convert USD value of bCoin back to native currency
				bidSizeNative := bidSize.Quo(liqMap[bDenom].price)
				bid := sdk.NewCoin(bDenom, bidSizeNative.MulInt(liqMap[bDenom].conversionFactor).TruncateInt())
				lot := dCoin

				// Start auction with this lot (deposit) and bid (borrow)
				err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorAccount, sdk.NewCoins(lot))
				if err != nil {
					return err
				}
				_, err = k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorAccount, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return err
				}

				// Adjust remaining value of remaining USD to be auctioned
				bUsdMap[bDenom] = bUsdMap[bDenom].Sub(bidSize)
				dUsdMap[dDenom] = sdk.ZeroDec()
				// Adjust amount of remaining bids/lots in native currencies
				borrows = borrows.Sub(sdk.NewCoins(bid))
				deposits = deposits.Sub(sdk.NewCoins(lot))
			}
		}
	}

	// 3. Attempt auctions where we can recover the remaining borrowed asset for some of the deposited asset
	for _, bDenom := range borrowDenoms {
		bCoin := sdk.NewCoin(bDenom, borrows.AmountOf(bDenom))
		// Already recovered all of this borrow asset, move to next asset
		if borrows.AmountOf(bDenom).Equal(sdk.ZeroInt()) {
			continue
		}

		// We need to raise this $ amount of 'borrow denom' using seized 'deposit denom'
		lotValueUSD := bUsdMap[bDenom].Quo(ltv)

		for _, dDenom := range depositDenoms {
			if dUsdMap[dDenom].GTE(lotValueUSD) {
				// Convert lot size USD to lot size native currency
				lotSizeNative := lotValueUSD.MulInt(liqMap[dDenom].conversionFactor).Quo(liqMap[dDenom].price)
				lot := sdk.NewCoin(dDenom, lotSizeNative.TruncateInt())
				bid := bCoin

				// Start auction with this lot (deposit) and bid (borrow)
				err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorAccount, sdk.NewCoins(lot))
				if err != nil {
					return err
				}
				_, err = k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorAccount, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return err
				}

				// Adjust remaining value of remaining USD to be auctioned
				bUsdMap[bDenom] = sdk.ZeroDec()
				dUsdMap[dDenom] = dUsdMap[dDenom].Sub(lotValueUSD)
				// Adjust amount of remaining bids/lots in native currencies
				borrows = borrows.Sub(sdk.NewCoins(bCoin))
				deposits = deposits.Sub(sdk.NewCoins(lot))
				break // No more borrow balance left for this denom, move to next borrow denom
			}
		}
	}

	return nil
}

func getDenoms(coins sdk.Coins) []string {
	denoms := []string{}
	for _, coin := range coins {
		denoms = append(denoms, coin.Denom)
	}
	return denoms
}

func removeDuplicates(one []string, two []string) []string {
	check := make(map[string]int)
	fullList := append(one, two...)

	res := []string{}
	for _, val := range fullList {
		check[val] = 1
	}

	for key := range check {
		res = append(res, key)
	}
	return res
}
