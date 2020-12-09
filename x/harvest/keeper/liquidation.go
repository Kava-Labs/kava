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
	currLtv := totalBorrowedUSDAmount.Quo(totalDepositedUSDAmount)
	err := k.SeizeDeposits(ctx, keeper, liqMap, deposits, borrowBalances, currLtv)
	if err != nil {
		return err
	}

	k.DeleteBorrow(ctx, borrow)

	return nil
}

// SeizeDeposits seizes a list of deposits and sends them to auction
func (k Keeper) SeizeDeposits(ctx sdk.Context, keeper sdk.AccAddress, liqMap map[string]LiqData, deposits []types.Deposit, borrowBalances sdk.Coins, ltv sdk.Dec) error {
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

	// Build maps to hold borrow and deposit coin USD valuations
	bUsdMap := make(map[string]sdk.Dec)
	for _, bCoin := range borrowBalances {
		bData := liqMap[bCoin.Denom]
		bCoinUsdValue := sdk.NewDecFromInt(bCoin.Amount).Quo(sdk.NewDecFromInt(bData.conversionFactor)).Mul(bData.price)
		bUsdMap[bCoin.Denom] = bCoinUsdValue
	}

	dUsdMap := make(map[string]sdk.Dec)
	for _, deposit := range aucDeposits {
		dData := liqMap[deposit.Denom]
		dCoinUsdValue := sdk.NewDecFromInt(deposit.Amount).Quo(sdk.NewDecFromInt(dData.conversionFactor)).Mul(dData.price)
		dUsdMap[deposit.Denom] = dCoinUsdValue
	}

	// Set up auction constants
	returnAddrs := []sdk.AccAddress{deposits[0].Depositor}
	weights := []sdk.Int{sdk.NewInt(100)}
	debt := sdk.NewCoin("debt", sdk.ZeroInt())

	// The % by which the lot must be larger than the borrow
	lotSizeFactor := sdk.OneDec().Quo(ltv)

	// Auction off any full lots (deposit USD values) >= bid (borrow USD value)
	for _, bCoin := range borrowBalances {
		for _, dCoin := range aucDeposits {
			minLotSize := dUsdMap[dCoin.Denom].Mul(lotSizeFactor)
			// Value of lot must be greater by (1/ltv)
			if minLotSize.GTE(bUsdMap[bCoin.Denom]) {

				// Convert USD back to native currency
				lotSizeNative := minLotSize.Quo(liqMap[dCoin.Denom].price)
				lot := sdk.NewCoin(dCoin.Denom, lotSizeNative.TruncateInt())
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

				bUsdMap[bCoin.Denom] = sdk.ZeroDec()
				dUsdMap[dCoin.Denom] = dUsdMap[dCoin.Denom].Sub(sdk.NewDec(minLotSize.Int64()))
				break // No more borrow balance left for this denom, move to next borrow denom
			}
		}
	}

	// For each deposit's USD value find a remaining USD borrow amount that can support it.
	for dKey, dValue := range dUsdMap {
		if dValue.IsZero() {
			continue
		}

		for bKey, bValue := range bUsdMap {
			if bValue.IsZero() {
				continue
			}

			// At this bid amount we'll sell all the collateral at a (1/ltv) ratio
			bidSize := dValue.Quo(lotSizeFactor)
			if bValue.GTE(bidSize) {
				// Convert USD value back to native currency
				bidSizeNative := bidSize.Quo(liqMap[bKey].price)
				lotSizeNative := dValue.Quo(liqMap[dKey].price)

				bid := sdk.NewCoin(bKey, bidSizeNative.TruncateInt())
				lot := sdk.NewCoin(dKey, lotSizeNative.TruncateInt())

				// Start auction with this lot (deposit) and bid (borrow)
				err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.LiquidatorAccount, sdk.NewCoins(lot))
				if err != nil {
					return err
				}
				_, err = k.auctionKeeper.StartCollateralAuction(ctx, types.LiquidatorAccount, lot, bid, returnAddrs, weights, debt)
				if err != nil {
					return err
				}

				bUsdMap[bKey] = bUsdMap[bKey].Sub(bidSizeNative)
				dUsdMap[dKey] = sdk.ZeroDec()
			}
			// TODO: Case where bValue is never large enough to cover bidSize.
			//		 Must either: sell less collateral to maintain (1/ltv) or sell at a different ratio.
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
